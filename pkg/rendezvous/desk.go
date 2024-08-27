package rendezvous

import (
	"context"
	"github.com/Xib1uvXi/xholepunch/pkg/util/netutil"
	"github.com/go-kratos/kratos/v2/log"
	"net"
	"sync"
)

//go:generate mockgen -source=desk.go -destination=desk.go_mock.go -package=rendezvous

type ConnectHandler interface {
	HandleConnect(conn *net.TCPConn, token string, nat int8) error
}

type ConnectMessage struct {
	Token   string `json:"token"`
	NATType int8   `json:"nat_type"`
}

func (m *ConnectMessage) Reset() *ConnectMessage {
	m.Token = ""
	m.NATType = 0
	return m
}

type FrontDesk struct {
	ctx            context.Context
	cancel         context.CancelFunc
	listenAddr     string
	tcpListener    *net.TCPListener
	connectHandler ConnectHandler
	msgPool        sync.Pool
}

func NewFrontDesk(listenAddr string, connectHandler ConnectHandler) (*FrontDesk, error) {
	ctx, cancel := context.WithCancel(context.Background())
	desk := &FrontDesk{
		ctx:            ctx,
		cancel:         cancel,
		listenAddr:     listenAddr,
		connectHandler: connectHandler,
		msgPool:        sync.Pool{New: func() interface{} { return &ConnectMessage{} }},
	}

	// new tcp listener
	tcpAddr, err := net.ResolveTCPAddr("tcp4", listenAddr)
	if err != nil {
		log.Errorf("resolve tcp addr error %v", err)
		return nil, err
	}

	tcpListener, err := net.ListenTCP("tcp4", tcpAddr)
	if err != nil {
		log.Errorf("listen tcp error: %v", err)
		return nil, err
	}

	desk.tcpListener = tcpListener

	return desk, nil
}

func (f *FrontDesk) Serve() {
	go f.acceptLoop()
}

func (f *FrontDesk) Close() error {
	f.cancel()
	return f.tcpListener.Close()
}

// acceptLoop
func (f *FrontDesk) acceptLoop() {
	log.Infof("front desk start to listen on %s", f.listenAddr)

	for {
		select {
		case <-f.ctx.Done():
			log.Infof("front desk stop to listen on %s", f.listenAddr)
			return
		default:
			conn, err := f.tcpListener.AcceptTCP()
			if err != nil {
				log.Errorf("accept tcp error: %v", err)
				continue
			}

			go f.handleConnection(conn)
		}
	}
}

func (f *FrontDesk) handleConnection(conn *net.TCPConn) {
	msg := f.msgPool.Get().(*ConnectMessage)
	defer func() {
		f.msgPool.Put(msg.Reset())
	}()

	msg.Reset()
	if err := netutil.ConnReceiveMessage(conn, msg); err != nil {
		log.Errorf("remote addr: %s,receive message error: %v", conn.RemoteAddr().String(), err)
		_ = conn.Close()
		return
	}

	if msg.Token == "" {
		log.Errorf("receive message invalid token: %s， remote addr: %s", msg.Token, conn.RemoteAddr().String())
		_ = conn.Close()
		return
	}

	if msg.NATType == 0 {
		log.Errorf("receive message invalid nat type: %d, remote addr: %s", msg.NATType, conn.RemoteAddr().String())
		_ = conn.Close()
		return
	}

	log.Infof("receive message token: %s, remote addr: %s", msg.Token, conn.RemoteAddr().String())

	if err := f.connectHandler.HandleConnect(conn, msg.Token, msg.NATType); err != nil {
		log.Errorf("handle connect error: %v", err)
		_ = conn.Close()
		return
	}
}
