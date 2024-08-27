package traversal

import (
	"fmt"
	"github.com/Xib1uvXi/xholepunch/pkg/util/json"
	"github.com/Xib1uvXi/xholepunch/pkg/util/netutil"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"net"
	"time"
)

var (
	ErrE2HDeadlineExceeded = errors.InternalServer("e2h timeout", "e2h timeout")
)

type e2hHandler struct {
	udpConn    *net.UDPConn
	localAddr  *net.UDPAddr
	remoteAddr *net.UDPAddr
	localNAT   int8
	remoteNAT  int8
	lAddr      string
	rAddr      string
	isActive   bool
}

func newE2HHandler(lAddr, rAddr string, localNAT, remoteNAT int8, isActive bool) (*e2hHandler, error) {
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

	log.Debugf("----------------- e2h %v listen on %s", isActive, localAddr.String())

	return &e2hHandler{
		udpConn:    udpConn,
		localAddr:  localAddr,
		remoteAddr: remoteAddr,
		localNAT:   localNAT,
		remoteNAT:  remoteNAT,
		lAddr:      lAddr,
		rAddr:      rAddr,
		isActive:   isActive,
	}, nil
}

// close
func (e *e2hHandler) Close() {
	if e.udpConn != nil {
		e.udpConn.Close()
	}
}

func (e *e2hHandler) HolePunch() (*HolePunchResult, error) {
	if e.isActive {
		return e.active()
	} else {
		return e.passive()
	}
}

// passive
func (e *e2hHandler) passive() (*HolePunchResult, error) {
	errCh := make(chan error)
	respCh := make(chan struct{})
	go func() {
		if err := e.asyncReceiver(respCh); err != nil {
			log.Debugf("receive error: %v", err)
			errCh <- err
		}
	}()

	rIp := e.remoteAddr.IP.String()
	rPort := e.remoteAddr.Port
	predictor := NewLinearPortPredictor(rPort)

	go func() {
		if err := e.sender(&HolePunchMessage{Empty: 1}, 2); err != nil {
			log.Debugf("send error: %v", err)
		}
	}()

	// port predict
	go func() {
		payload, err := json.StringifyJsonToBytesWithErr(&HolePunchMessage{Empty: 1})
		if err != nil {
			log.Errorf("stringify json error: %v", err)
			return
		}

		for i := 0; i < 20; i++ {
			nextPort := predictor.NextPort()
			newRAddr := fmt.Sprintf("%s:%d", rIp, nextPort)
			log.Debugf("predict next port: %s", newRAddr)

			target, err := net.ResolveUDPAddr("udp4", newRAddr)

			if err != nil {
				log.Errorf("resolve udp address error: %v", err)
				continue
			}

			_ = netutil.UdpSendByteMessage(e.udpConn, target, payload)
			_ = netutil.UdpSendByteMessage(e.udpConn, target, payload)
		}
	}()

	select {
	case <-respCh:
		return &HolePunchResult{
			LocalAddr:  e.localAddr.String(),
			RemoteAddr: e.remoteAddr.String(),
			LocalNAT:   e.localNAT,
			RemoteNAT:  e.remoteNAT,
		}, nil

	case err := <-errCh:
		log.Errorf("e2e error: %v", err)
		return nil, err
	case <-time.After(15 * time.Second):
		return nil, ErrE2HDeadlineExceeded
	}
}

// active
func (e *e2hHandler) active() (*HolePunchResult, error) {
	respCh := make(chan struct{})
	go func() {
		if err := e.asyncReceiver(respCh); err != nil {
			log.Debugf("receive error: %v", err)
		}
	}()

	go func() {
		select {
		case <-respCh:
			log.Infof("e2h receive hole punch message from %s", e.remoteAddr.String())

		case <-time.After(10 * time.Second):
			return
		}
	}()

	if err := e.sender(&HolePunchMessage{Empty: 1}, 6); err != nil {
		return nil, err
	}

	return &HolePunchResult{
		LocalAddr:  e.localAddr.String(),
		RemoteAddr: e.remoteAddr.String(),
		LocalNAT:   e.localNAT,
		RemoteNAT:  e.remoteNAT,
	}, nil

}

func (e *e2hHandler) sender(msg *HolePunchMessage, times int) error {
	payload, err := json.StringifyJsonToBytesWithErr(msg)
	if err != nil {
		log.Errorf("stringify json error: %v", err)
		return err
	}

	if e.isActive {
		log.Debugf("------------------ %s active send message: %s", e.localAddr.String(), e.remoteAddr.String())
	}

	for i := 0; i < times; i++ {
		if err := netutil.UdpSendByteMessage(e.udpConn, e.remoteAddr, payload); err != nil {
			log.Debugf("send message error: %v", err)
			continue
		}
	}

	return nil
}

func (e *e2hHandler) asyncReceiver(result chan struct{}) error {
	var resp HolePunchMessage
	addr, err := netutil.UdpReceiveMessage(e.udpConn, &resp)
	if err != nil {
		log.Errorf("receive udp message error: %v", err)
		return err
	}

	if addr.String() != e.rAddr {
		log.Errorf("address %s not match: %v", addr.String(), e.rAddr)
		return ErrAddrNotMatch
	}

	if resp.Empty != 1 {
		return ErrReceiveError
	}

	log.Infof("e2e receive hole punch message from %s", addr.String())

	result <- struct{}{}
	return nil
}
