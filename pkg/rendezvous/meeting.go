package rendezvous

import (
	"github.com/Xib1uvXi/xholepunch/pkg/types"
	"github.com/Xib1uvXi/xholepunch/pkg/util/json"
	"github.com/Xib1uvXi/xholepunch/pkg/util/netutil"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/quic-go/quic-go"
	"net"
	"sync"
	"time"
)

type CheckinMessage struct {
	Ack int8 `json:"ack"`
}

type NegotiationMessage struct {
	LocalPublicAddr  string `json:"local_public_addr"`
	RemotePublicAddr string `json:"remote_public_addr"`
	RemoteNATType    int8   `json:"remote_nat_type"`
	ServerPort       int    `json:"server_port"`
	IsActive         bool   `json:"is_active"`
}

type HolePunchMessage struct {
	Addr string `json:"addr"`
}

type meetingImpl struct {
}

func NewMeetingImpl() MeetingHandler {
	return &meetingImpl{}
}

func (m *meetingImpl) Meeting(token string, conn1, conn2 *holePunchConn) error {
	var conn1Active bool
	// check is active
	if types.NATType(conn1.natType) == types.Symmetric && types.NATType(conn2.natType) != types.Symmetric {
		conn1Active = true
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	var addr1, addr2 string

	go func() {
		defer wg.Done()
		addrCh1, err := m.genMeetingAddr(conn1, conn2, conn1Active)
		if err != nil {
			log.Errorf("gen meeting addr error: %v", err)
			return
		}

		select {
		case addr1 = <-addrCh1:
			log.Infof("conn1 addr: %s", addr1)
			return
		case <-time.After(30 * time.Second):
			log.Errorf("get conn1 meeting udp addr timeout")
		}

	}()

	go func() {
		defer wg.Done()
		addrCh2, err := m.genMeetingAddr(conn2, conn1, !conn1Active)
		if err != nil {
			log.Errorf("gen meeting addr error: %v", err)
			return
		}

		select {
		case addr2 = <-addrCh2:
			log.Infof("conn2 addr: %s", addr2)
			return

		case <-time.After(30 * time.Second):
			log.Errorf("get conn2 meeting udp addr timeout")
		}
	}()

	wg.Wait()

	if addr1 == "" || addr2 == "" {
		log.Errorf("get udp addr error, addr1 %s addr2 %s", addr1, addr2)
	}

	if conn1Active {
		if err := m.startPunching(conn2, addr1); err != nil {
			log.Errorf("start punching error: %v", err)
			return err
		}

		if err := m.startPunching(conn1, addr2); err != nil {
			log.Errorf("start punching error: %v", err)
			return err
		}
	} else {
		if err := m.startPunching(conn1, addr2); err != nil {
			log.Errorf("start punching error: %v", err)
			return err
		}

		if err := m.startPunching(conn2, addr1); err != nil {
			log.Errorf("start punching error: %v", err)
			return err
		}
	}

	return nil
}

// startPunching
func (m *meetingImpl) startPunching(conn1 *holePunchConn, addr2 string) error {
	if err := netutil.ConnSendMessage(conn1.conn, &HolePunchMessage{Addr: addr2}); err != nil {
		log.Errorf("send hole punch message error: %v", err)
		return err
	}

	return nil
}

// genMeeting
func (m *meetingImpl) genMeetingAddr(conn1 *holePunchConn, conn2 *holePunchConn, IsActive bool) (chan string, error) {
	var addr = make(chan string)

	udpConn, err := netutil.UDPRandListen()
	if err != nil {
		log.Errorf("new quic udp error: %v", err)
		return nil, err
	}

	if err := m.getUDPAddr(udpConn, addr); err != nil {
		log.Errorf("get udp addr error: %v", err)
		return nil, err
	}

	if err := m.negotiation(conn1, conn2, udpConn.LocalAddr().(*net.UDPAddr).Port, IsActive); err != nil {
		log.Errorf("negotiation error: %v", err)
		return nil, err
	}

	return addr, nil
}

// new quic udp
func (m *meetingImpl) getUDPAddr(udpConn *net.UDPConn, addr chan string) error {
	var udpAddr = make(chan string)

	quicListener, err := netutil.NewReliableUDPServer(udpConn)
	if err != nil {
		log.Errorf("new reliable udp server error: %v", err)
		return err
	}

	handler := func(conn quic.Connection, stream quic.Stream) error {
		data := make([]byte, 1024)
		n, err := stream.Read(data)
		if err != nil {
			log.Errorf("stream read error: %v", err)
			return err
		}

		var checkinMsg CheckinMessage
		if err := json.ParseJsonFromBytes(data[:n], &checkinMsg); err != nil {
			log.Errorf("parse json from bytes error: %v", err)
			return err
		}

		if checkinMsg.Ack != 1 {
			log.Errorf("checkin message ack error: %v", checkinMsg.Ack)
			return errors.BadRequest("checkin message ack error", "ack error")
		}

		log.Debugf("checkin message ack: %v", conn.RemoteAddr().String())

		// adderss
		udpAddr <- conn.RemoteAddr().String()
		return nil
	}

	quicListener.SetStreamHandler(handler)
	if err := quicListener.Start(); err != nil {
		log.Errorf("quic listener start error: %v", err)
		_ = udpConn.Close()
		return err
	}

	go func() {
		defer quicListener.Close()

		select {
		case uaddr := <-udpAddr:
			addr <- uaddr
			return

		case <-time.After(30 * time.Second):
			log.Errorf("get udp addr timeout")
			return
		}
	}()

	return nil
}

// Negotiation
func (m *meetingImpl) negotiation(conn *holePunchConn, peerConn *holePunchConn, serverPort int, IsActive bool) error {
	msg := &NegotiationMessage{
		LocalPublicAddr:  conn.conn.RemoteAddr().String(),
		RemotePublicAddr: peerConn.conn.RemoteAddr().String(),
		RemoteNATType:    int8(peerConn.natType),
		ServerPort:       serverPort,
		IsActive:         IsActive,
	}

	if err := netutil.ConnSendMessage(conn.conn, msg); err != nil {
		log.Errorf("send negotiation message error: %v", err)
		return err
	}

	return nil
}
