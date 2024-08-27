package traversal

import (
	"github.com/Xib1uvXi/xholepunch/pkg/util/json"
	"github.com/Xib1uvXi/xholepunch/pkg/util/netutil"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"net"
	"time"
)

var (
	ErrAddrNotMatch        = errors.InternalServer("address not match", "address not match")
	ErrReceiveError        = errors.InternalServer("receive error", "receive error")
	ErrE2EDeadlineExceeded = errors.InternalServer("e2e timeout", "e2e timeout")
)

type e2eHandler struct {
	udpConn    *net.UDPConn
	localAddr  *net.UDPAddr
	remoteAddr *net.UDPAddr
	localNAT   int8
	remoteNAT  int8
	lAddr      string
	rAddr      string
}

func newE2EHandler(lAddr, rAddr string, localNAT, remoteNAT int8) (*e2eHandler, error) {
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

	return &e2eHandler{
		udpConn:    udpConn,
		localAddr:  localAddr,
		remoteAddr: remoteAddr,
		localNAT:   localNAT,
		remoteNAT:  remoteNAT,
		lAddr:      lAddr,
		rAddr:      rAddr,
	}, nil
}

// close
func (e *e2eHandler) Close() {
	if e.udpConn != nil {
		e.udpConn.Close()
	}
}

func (e *e2eHandler) HolePunch() (*HolePunchResult, error) {
	errCh := make(chan error)
	respCh := make(chan struct{})
	go func() {
		if err := e.asyncReceiver(respCh); err != nil {
			log.Debugf("receive error: %v", err)
		}
	}()

	go func() {
		if err := e.sender(&HolePunchMessage{Empty: 1}); err != nil {
			log.Debugf("send error: %v", err)
			return
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
	case <-time.After(10 * time.Second):
		return nil, ErrE2EDeadlineExceeded
	}
}

func (e *e2eHandler) sender(msg *HolePunchMessage) error {
	payload, err := json.StringifyJsonToBytesWithErr(msg)
	if err != nil {
		log.Errorf("stringify json error: %v", err)
		return err
	}

	for i := 0; i < 6; i++ {
		if err := netutil.UdpSendByteMessage(e.udpConn, e.remoteAddr, payload); err != nil {
			log.Debugf("send message error: %v", err)
			continue
		}
	}

	return nil
}

func (e *e2eHandler) asyncReceiver(result chan struct{}) error {
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
