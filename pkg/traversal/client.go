package traversal

import (
	"context"
	"fmt"
	"github.com/Xib1uvXi/xholepunch/pkg/rendezvous"
	"github.com/Xib1uvXi/xholepunch/pkg/util/json"
	"github.com/Xib1uvXi/xholepunch/pkg/util/netutil"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/quic-go/quic-go"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"
)

type HolePunchResult struct {
	LocalAddr  string `json:"local_addr"`
	RemoteAddr string `json:"remote_addr"`
	LocalNAT   int8   `json:"local_nat"`
	RemoteNAT  int8   `json:"remote_nat"`
}

func (r HolePunchResult) String() string {
	return fmt.Sprintf("local addr: %s, remote addr: %s, local nat: %d, remote nat: %d", r.LocalAddr, r.RemoteAddr, r.LocalNAT, r.RemoteNAT)
}

type Client struct {
	serverAddr       string
	localAddr        string
	NATType          int8
	holePunchHandler HolePunchHandler
	connectFactory   ConnectFactory
	closeFn          []func()
}

type HolePunchHandler interface {
	HolePunching(localAddr string, localNAT int8, msg *rendezvous.NegotiationMessage, remoteAddr string) (*HolePunchResult, error)
}

type ConnectFactory interface {
	Connect(localAddr string, remoteAddr string, isActive bool) error
}

func NewClient(serverAddr string, natType int8, holePunchHandler HolePunchHandler, connectFactory ConnectFactory) *Client {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	localAddr := ":" + fmt.Sprint(rand.Intn(23000)+10000)

	return &Client{
		serverAddr:       serverAddr,
		localAddr:        localAddr,
		NATType:          natType,
		holePunchHandler: holePunchHandler,
		connectFactory:   connectFactory,
	}
}

func (c *Client) Close() {
	for _, fn := range c.closeFn {
		fn()
	}
}

// HolePunching
func (c *Client) HolePunching(token string) error {
	if err := c.connect(token); err != nil {
		log.Errorf("connect error: %v", err)
		return err
	}

	return nil
}

// connect
func (c *Client) connect(token string) error {
	tcpConn, err := net.Dial("tcp4", c.serverAddr)
	if err != nil {
		log.Errorf("dial tcp error: %v", err)
		return err
	}

	msg := &rendezvous.ConnectMessage{
		Token:   token,
		NATType: c.NATType,
	}

	if err := netutil.ConnSendMessage(tcpConn, msg); err != nil {
		log.Errorf("send message error: %v", err)
		_ = tcpConn.Close()
		return err
	}

	if err := c.negotiation(tcpConn); err != nil {
		log.Errorf("negotiation error: %v", err)
		_ = tcpConn
	}

	return nil
}

func (c *Client) negotiation(conn net.Conn) error {
	// set timeout
	_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	var msg rendezvous.NegotiationMessage
	if err := netutil.ConnReceiveMessage(conn, &msg); err != nil {
		log.Errorf("receive message error: %v", err)
		return err
	}

	// reset timeout
	_ = conn.SetReadDeadline(time.Time{})
	log.Infof("receive hole punching negotiation message: %v", msg)

	wg := sync.WaitGroup{}
	wg.Add(2)
	resultCh := make(chan *rendezvous.HolePunchMessage)
	errCh := make(chan error)
	go func() {
		var result rendezvous.HolePunchMessage
		if err := netutil.ConnReceiveMessage(conn, &result); err != nil {
			log.Errorf("receive message error: %v", err)
			errCh <- err
			return
		}

		resultCh <- &result
	}()

	go func() {
		if err := c.ackNegotiation(&msg); err != nil {
			log.Errorf("ack negotiation error: %v", err)
			return
		}
	}()

	select {
	case result := <-resultCh:
		log.Infof("receive hole punching message: %v", result)

		if result.Addr == "" {
			log.Errorf("receive hole punching message get remote add error: %v", result)
			return fmt.Errorf("receive hole punching message get remote add error")
		}

		// hole punching
		hpResult, err := c.holePunchHandler.HolePunching(c.localAddr, c.NATType, &msg, result.Addr)
		if err != nil {
			log.Errorf("hole punching error: %v", err)
			return err
		}

		log.Infof("hole punching result: %v", hpResult)

		if err := c.connectFactory.Connect(hpResult.LocalAddr, hpResult.RemoteAddr, msg.IsActive); err != nil {
			log.Errorf("create connect error: %v", err)
			return err
		}

	case err := <-errCh:
		log.Errorf("receive hole punching message error: %v", err)
		return err
	case <-time.After(30 * time.Second):
		log.Errorf("receive hole punching negotiation udp addr message timeout")
	}

	return nil
}

// ack negotiation
func (c *Client) ackNegotiation(msg *rendezvous.NegotiationMessage) error {
	quicServerAddr := fmt.Sprintf("%s%d", c.serverAddr[:strings.LastIndex(c.serverAddr, ":")+1], msg.ServerPort)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// fixme: need change c.localAddr to real local addr

	quicClient, err := quic.DialAddr(ctx, quicServerAddr, netutil.ClientTLSConfig(), nil)
	if err != nil {
		log.Errorf("dial quic server error: %v", err)
		return err
	}

	stream, err := quicClient.OpenStreamSync(ctx)
	if err != nil {
		log.Errorf("open stream error: %v", err)
		return err
	}

	c.closeFn = append(c.closeFn, func() {
		stream.Close()
		quicClient.CloseWithError(11, "closing connection")

	})

	_ = stream.SetWriteDeadline(time.Now().Add(5 * time.Second))

	payload, err := json.StringifyJsonToBytesWithErr(&rendezvous.CheckinMessage{Ack: 1})
	if err != nil {
		log.Errorf("stringify json to bytes error: %v", err)
		return err
	}

	_, err = stream.Write([]byte(payload))
	if err != nil {
		log.Errorf("write message error: %v", err)
		return err
	}

	return nil
}
