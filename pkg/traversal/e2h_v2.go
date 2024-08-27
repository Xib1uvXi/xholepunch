package traversal

import (
	"fmt"
	"github.com/Doraemonkeys/reliableUDP"
	"github.com/Xib1uvXi/xholepunch/pkg/util/netutil"
	"github.com/go-kratos/kratos/v2/log"
	"net"
	"time"
)

const udpTimeout = 8 * time.Second

type e2hHandlerV2 struct {
	udpConn    *net.UDPConn
	localAddr  *net.UDPAddr
	remoteAddr *net.UDPAddr
	localNAT   int8
	remoteNAT  int8
	lAddr      string
	rAddr      string
	isActive   bool
	rudpConn   *reliableUDP.ReliableUDP
}

func newE2HHandlerV2(lAddr, rAddr string, localNAT, remoteNAT int8, isActive bool) (*e2hHandlerV2, error) {
	localAddr, err := net.ResolveUDPAddr("udp4", lAddr)
	if err != nil {
		return nil, err
	}

	remoteAddr, err := net.ResolveUDPAddr("udp4", rAddr)
	if err != nil {
		return nil, err
	}

	udpConn, err := net.ListenUDP("udp4", localAddr)
	if err != nil {
		return nil, err
	}

	v2 := &e2hHandlerV2{
		udpConn:    udpConn,
		localAddr:  localAddr,
		remoteAddr: remoteAddr,
		localNAT:   localNAT,
		remoteNAT:  remoteNAT,
		lAddr:      lAddr,
		rAddr:      rAddr,
		isActive:   isActive,
		rudpConn:   reliableUDP.New(udpConn),
	}

	v2.rudpConn.SetGlobalReceive()

	return v2, nil
}

func (e *e2hHandlerV2) Close() {
	if e.udpConn != nil {
		e.udpConn.Close()
	}

	if e.rudpConn != nil {
		e.rudpConn.Close()
	}
}

func (e *e2hHandlerV2) HolePunch() (*HolePunchResult, error) {
	if e.isActive {
		return e.active()
	} else {
		return e.passive()
	}
}

func (e *e2hHandlerV2) passive() (*HolePunchResult, error) {

	log.Infof("e2hHandlerV2 passive")

	rIp := e.remoteAddr.IP.String()
	rPort := e.remoteAddr.Port
	predictor := NewLinearPortPredictor(rPort)

	go func() {
		for i := 0; i < 20; i++ {
			newRAddr := fmt.Sprintf("%s:%d", rIp, predictor.NextPort())
			log.Debug("new_rAddr", newRAddr)
			_ = netutil.RUDPSendUnreliableMessage(e.rudpConn, newRAddr, &HolePunchMessage{Empty: 1})
			_ = netutil.RUDPSendUnreliableMessage(e.rudpConn, newRAddr, &HolePunchMessage{Empty: 1})
		}
	}()

	for {
		log.Infof("e2hHandlerV2 passive start receive message")
		var msg HolePunchMessage
		addr, err := netutil.RUDPReceiveAllMessage(e.rudpConn, udpTimeout, &msg)
		if err != nil {
			return nil, fmt.Errorf("easy 2 symmetric receive message error %w", err)
		}

		return &HolePunchResult{
			LocalAddr:  e.lAddr,
			RemoteAddr: addr.String(),
			LocalNAT:   e.localNAT,
			RemoteNAT:  e.remoteNAT,
		}, nil
	}
}

func (e *e2hHandlerV2) active() (*HolePunchResult, error) {

	log.Infof("e2hHandlerV2 active")

	for i := 0; i < 3; i++ {
		log.Infof("e2hHandlerV2 active send message")
		if err := netutil.RUDPSendMessage(e.rudpConn, e.rAddr, &HolePunchMessage{Empty: 1}, udpTimeout); err != nil {
			log.Debugf("symmetric2EasyNAT send message error: %v \n", err)
			continue
		}

		return &HolePunchResult{
			LocalAddr:  e.lAddr,
			RemoteAddr: e.rAddr,
			LocalNAT:   e.localNAT,
			RemoteNAT:  e.remoteNAT,
		}, nil
	}

	return nil, fmt.Errorf("symmetric 2 easy hole punching failed, no response")
}
