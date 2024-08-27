package rendezvous

import (
	"context"
	"github.com/Xib1uvXi/xholepunch/pkg/types"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/panjf2000/ants/v2"
	"github.com/zekroTJA/timedmap/v2"
	"net"
	"time"
)

//go:generate mockgen -source=waiting_room.go -destination=mock/waiting_room.go_mock.go -package=rendezvous

var (
	ErrTokenNotFound = errors.BadRequest("token not found", "token not found")
)

type MeetingHandler interface {
	Meeting(token string, conn1, conn2 *holePunchConn) error
}

type holePunchConn struct {
	conn    net.Conn
	natType types.NATType
}

type RoomKeyCard struct {
	exitCh  chan struct{}
	entryCh chan *holePunchConn
}

type WaitingRoomManager struct {
	ctx            context.Context
	cancel         context.CancelFunc
	tokenMap       *timedmap.TimedMap[string, *RoomKeyCard]
	meetingHandler MeetingHandler
	runningPool    *ants.Pool

	tokenTTL time.Duration
}

func (wrm *WaitingRoomManager) HandleConnect(conn *net.TCPConn, token string, nat int8) error {
	if wrm.tokenMap.Contains(token) {
		return wrm.JoinWaitingRoom(token, conn, types.NATType(nat))
	}

	return wrm.CreateWaitingRoom(conn, token, types.NATType(nat))
}

func NewWaitingRoomManager(meetingHandler MeetingHandler, tokenTTL time.Duration, tokenCleanInterval time.Duration) (*WaitingRoomManager, error) {
	ctx, cancel := context.WithCancel(context.Background())
	wrm := &WaitingRoomManager{
		ctx:            ctx,
		cancel:         cancel,
		tokenMap:       timedmap.New[string, *RoomKeyCard](tokenCleanInterval),
		meetingHandler: meetingHandler,
		tokenTTL:       tokenTTL,
	}

	// goroutine init alloc 2kb stack, 100000  * 2kb = 200mb
	pool1, err := ants.NewPool(100000, ants.WithPreAlloc(true))

	if err != nil {
		return nil, err
	}

	wrm.runningPool = pool1

	return wrm, nil
}

func (wrm *WaitingRoomManager) Close() error {
	wrm.cancel()
	wrm.runningPool.Release()

	return nil
}

// CreateWaitingRoom Create a new token and a channel to wait for a connection
func (wrm *WaitingRoomManager) CreateWaitingRoom(tcpConn net.Conn, token string, nat types.NATType) error {
	exitCh := make(chan struct{})
	entryCh := make(chan *holePunchConn)

	wrm.tokenMap.Set(token, &RoomKeyCard{exitCh: exitCh, entryCh: entryCh}, wrm.tokenTTL, func(value *RoomKeyCard) {
		close(value.exitCh)
		log.Infof("token %s expired", token)
	})

	if err := wrm.runningPool.Submit(func() { wrm.waitConnect(token, &holePunchConn{conn: tcpConn, natType: nat}, entryCh, exitCh) }); err != nil {
		log.Errorf("create waiting room, conn pool submit error: %s", err)
		return err
	}

	return nil
}

// JoinWaitingRoom Join the waiting room with the token
func (wrm *WaitingRoomManager) JoinWaitingRoom(token string, conn net.Conn, natType types.NATType) error {
	value, ok := wrm.tokenMap.GetValue(token)
	if !ok {
		return ErrTokenNotFound.WithMetadata(map[string]string{"token": token})
	}

	value.entryCh <- &holePunchConn{conn: conn, natType: natType}

	return nil
}

func (wrm *WaitingRoomManager) waitConnect(token string, conn1 *holePunchConn, entryCh chan *holePunchConn, exitCh chan struct{}) {
	log.Debugf("waiting for token %s", token)

	select {
	case <-wrm.ctx.Done():
		log.Infof("waiting for token %s srv canceled", token)
		_ = conn1.conn.Close()
		return
	case <-exitCh:
		log.Infof("token %s room exit", token)
		_ = conn1.conn.Close()
		return

	case hpConn := <-entryCh:
		log.Infof("token %s room entry", token)

		handler := func() {
			log.Infof("token %s room entry handler", token)
			if err := wrm.meetingHandler.Meeting(token, conn1, hpConn); err != nil {
				log.Errorf("meeting handler error: %s", err)
				_ = conn1.conn.Close()
				_ = hpConn.conn.Close()
			}
		}

		if err := wrm.runningPool.Submit(handler); err != nil {
			log.Errorf("meeting pool submit error: %s", err)
			_ = conn1.conn.Close()
			_ = hpConn.conn.Close()
		}
	}

	return
}
