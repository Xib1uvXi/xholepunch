package holepunch

import (
	"fmt"
	"github.com/Xib1uvXi/xholepunch/pkg/traversalv2/predictor"
	"github.com/Xib1uvXi/xholepunch/pkg/util/netutil"
	"github.com/go-kratos/kratos/v2/log"
)

type PortRestrictedCone2Symmetric struct {
	*BaseModel
}

func NewPortRestrictedCone2Symmetric(lAddr, rAddr string, localNAT, remoteNAT int8, isActive bool) (*PortRestrictedCone2Symmetric, error) {
	baseModel, err := newBaseModel(lAddr, rAddr, localNAT, remoteNAT, isActive)
	if err != nil {
		return nil, err
	}

	return &PortRestrictedCone2Symmetric{
		BaseModel: baseModel,
	}, nil
}

func (hp *PortRestrictedCone2Symmetric) HolePunch() (*Result, error) {
	if hp.isActive {
		return hp.symmetric()
	}

	return hp.portRestrictedCone()
}

func (hp *PortRestrictedCone2Symmetric) Close() {
	hp.BaseModel.Close()
}

// Symmetric
func (hp *PortRestrictedCone2Symmetric) symmetric() (*Result, error) {
	for i := 0; i < 3; i++ {
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

	return nil, fmt.Errorf("symmetric 2 easy hole punching failed, no response")
}

// PortRestrictedCone
func (hp *PortRestrictedCone2Symmetric) portRestrictedCone() (*Result, error) {
	rIp := hp.remoteAddr.IP.String()
	rPort := hp.remoteAddr.Port
	ppPredictor := predictor.NewPseudorandomPortPredictor(rPort, 460)

	go func() {
		for {
			pport := ppPredictor.NextPort()
			if pport == 0 {
				break
			}

			newRAddr := fmt.Sprintf("%s:%d", rIp, pport)
			//log.Debugf("Pseudorandom 2 symmetric send message to %s", newRAddr)
			_ = netutil.RUDPSendUnreliableMessage(hp.rudpConn, newRAddr, &HolePunchMessage{Empty: 1})
			_ = netutil.RUDPSendUnreliableMessage(hp.rudpConn, newRAddr, &HolePunchMessage{Empty: 1})
		}

		log.Debugf("Pseudorandom 2 symmetric send all message done")
	}()

	for {
		var msg HolePunchMessage
		addr, err := netutil.RUDPReceiveAllMessage(hp.rudpConn, udpTimeout, &msg)
		if err != nil {
			return nil, fmt.Errorf("easy 2 symmetric receive message error %w", err)
		}

		return &Result{
			LocalAddr:  hp.lAddr,
			RemoteAddr: addr.String(),
			LocalNAT:   hp.localNAT,
			RemoteNAT:  hp.remoteNAT,
		}, nil
	}

}
