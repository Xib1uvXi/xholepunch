package holepunch

import (
	"fmt"
	"github.com/Xib1uvXi/xholepunch/pkg/util/netutil"
	"github.com/go-kratos/kratos/v2/log"
)

type Easy2Easy struct {
	*BaseModel
}

func NewEasy2Easy(lAddr, rAddr string, localNAT, remoteNAT int8, isActive bool) (*Easy2Easy, error) {
	baseModel, err := newBaseModel(lAddr, rAddr, localNAT, remoteNAT, isActive)
	if err != nil {
		return nil, err
	}

	return &Easy2Easy{
		BaseModel: baseModel,
	}, nil
}

func (hp *Easy2Easy) HolePunch() (*Result, error) {
	if hp.isActive {
		return hp.active()
	}

	return hp.passive()
}

// passive
func (hp *Easy2Easy) passive() (*Result, error) {

	// waiting peer connect
	for i := 0; i < 2; i++ {
		if err := netutil.RUDPSendUnreliableMessage(hp.rudpConn, hp.rAddr, &HolePunchMessage{Empty: 1}); err != nil {
			log.Debugf("easy2easy passive send message error: %v", err)
			continue
		}
	}

	for {
		var msg HolePunchMessage
		newAddr, err := netutil.RUDPReceiveAllMessage(hp.rudpConn, udpTimeout, &msg)
		if err != nil {
			return nil, fmt.Errorf("hole punching error %w", err)
		}

		return &Result{
			LocalAddr:  hp.lAddr,
			RemoteAddr: newAddr.String(),
			LocalNAT:   hp.localNAT,
			RemoteNAT:  hp.remoteNAT,
		}, nil
	}
}

// active
func (hp *Easy2Easy) active() (*Result, error) {
	for i := 0; i < 3; i++ {
		log.Infof("e2hHandlerV2 active send message")
		if err := netutil.RUDPSendMessage(hp.rudpConn, hp.rAddr, &HolePunchMessage{Empty: 1}, udpTimeout); err != nil {
			log.Debugf("symmetric2EasyNAT send message error: %v \n", err)
			continue
		}

		return &Result{
			LocalAddr:  hp.lAddr,
			RemoteAddr: hp.rAddr,
			LocalNAT:   hp.localNAT,
			RemoteNAT:  hp.remoteNAT,
		}, nil
	}

	return nil, fmt.Errorf("easy 2 easy hole punching failed, no response")
}
