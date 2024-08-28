package traversalv2

import (
	"fmt"
	"github.com/Doraemonkeys/reliableUDP"
	"github.com/Xib1uvXi/xholepunch/pkg/rendezvous"
	"github.com/Xib1uvXi/xholepunch/pkg/util/netutil"
	"github.com/go-kratos/kratos/v2/log"
	"math/rand"
	"net"
	"strings"
	"time"
)

const getRemoteAddrTimeout = 10 * time.Second

type Client struct {
	serverAddr       string
	NATType          int8
	holePunchHandler HolePunchHandler
	connectFactory   ConnectFactory
	cleanup          []func()

	localAddr string
}

func NewClient(serverAddr string, natType int8, holePunchHandler HolePunchHandler, connectFactory ConnectFactory) *Client {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	localAddr := ":" + fmt.Sprint(rand.Intn(23000)+10000)

	log.Debugf("client rand localAddr: %s", localAddr)

	return &Client{
		serverAddr:       serverAddr,
		localAddr:        localAddr,
		NATType:          natType,
		holePunchHandler: holePunchHandler,
		connectFactory:   connectFactory,
	}
}

func (c *Client) HolePunching(token string) error {
	conn, err := c.connect(token)
	if err != nil {
		return err
	}

	negotiationMessage, err := c.negotiation(conn)
	if err != nil {
		return err
	}

	addrChan := make(chan string)

	go func() {
		c.getRemoteAddr(conn, addrChan)
	}()

	if err := c.sendMsgToNewServerPort(negotiationMessage); err != nil {
		return err
	}

	var targetRemoteAddr string
	select {
	case targetRemoteAddr = <-addrChan:
		log.Debugf("receive public addr: %s \n", targetRemoteAddr)
	case <-time.After(getRemoteAddrTimeout):
		return fmt.Errorf("receive remote public addr timeout")
	}

	// hole punching
	hpResult, err := c.holePunchHandler.HolePunching(c.localAddr, c.NATType, negotiationMessage, targetRemoteAddr)
	if err != nil {
		log.Errorf("hole punching error: %v", err)
		return err
	}

	log.Infof("hole punching result: %v", hpResult)

	if err := c.connectFactory.Connect(hpResult.LocalAddr, hpResult.RemoteAddr, negotiationMessage.IsActive); err != nil {
		log.Errorf("create connect error: %v", err)
		return err
	}

	return nil
}

func (c *Client) Close() {
	for _, fn := range c.cleanup {
		fn()
	}
}

// connect
func (c *Client) connect(token string) (conn net.Conn, err error) {
	tcpConn, err := net.Dial("tcp4", c.serverAddr)
	if err != nil {
		log.Errorf("dial tcp error: %v", err)
		return nil, err
	}

	c.cleanup = append(c.cleanup, func() { tcpConn.Close() })

	if err := netutil.ConnSendMessage(tcpConn, &rendezvous.ConnectMessage{
		Token:   token,
		NATType: c.NATType,
	}); err != nil {
		log.Errorf("send message error: %v", err)
		return nil, err
	}

	return tcpConn, nil
}

func (c *Client) negotiation(conn net.Conn) (*rendezvous.NegotiationMessage, error) {
	// set timeout
	_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	var msg rendezvous.NegotiationMessage
	if err := netutil.ConnReceiveMessage(conn, &msg); err != nil {
		log.Errorf("receive message error: %v", err)
		return nil, err
	}

	// reset timeout
	_ = conn.SetReadDeadline(time.Time{})
	log.Infof("receive hole punching negotiation message: %v", msg)

	return &msg, nil
}

// send msg to new server port
func (c *Client) sendMsgToNewServerPort(msg *rendezvous.NegotiationMessage) error {
	newServerAddr := fmt.Sprintf("%s%d", c.serverAddr[:strings.LastIndex(c.serverAddr, ":")+1], msg.ServerPort)
	log.Debugf("new server addr: %s", newServerAddr)
	lAddr, err := net.ResolveUDPAddr("udp4", c.localAddr)
	if err != nil {
		log.Debugf("resolve udp addr error: %v \n", err)
		return err
	}

	udpConn, err := net.ListenUDP("udp4", lAddr)
	if err != nil {
		log.Debugf("listen udp error: %v \n", err)
		return err
	}

	defer udpConn.Close()
	rudpConn := reliableUDP.New(udpConn)
	defer rudpConn.Close()

	if err := netutil.RUDPSendMessage(rudpConn, newServerAddr, &rendezvous.CheckinMessage{Ack: 1}, 5*time.Second); err != nil {
		log.Debugf("send message to new server port error: %v", err)
		return err
	}

	return nil
}

// getRemoteAddr
func (c *Client) getRemoteAddr(conn net.Conn, remoteAddrCh chan string) {
	var result rendezvous.HolePunchMessage
	if err := netutil.ConnReceiveMessage(conn, &result); err != nil {
		log.Errorf("get remote addr receive message error: %v", err)
		return
	}

	if result.Addr == "" {
		log.Errorf("get remote addr error: %v", result)
		return
	}

	remoteAddrCh <- result.Addr
}