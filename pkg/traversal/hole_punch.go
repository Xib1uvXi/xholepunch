package traversal

import (
	"github.com/Xib1uvXi/xholepunch/pkg/rendezvous"
	"github.com/Xib1uvXi/xholepunch/pkg/types"
	"github.com/go-kratos/kratos/v2/errors"
)

const (
	// HolePunchTypeNone represents no hole punch type
	hard2hard = iota + 1
	easy2easy
	easy2hard
)

var (
	ErrNoSupportHard2Hard   = errors.BadRequest("no support hard2hard", "no support hard2hard")
	ErrUnknownHolePunchType = errors.BadRequest("unknown hole punch type", "unknown hole punch type")
)

type HolePunchMessage struct {
	Empty int8 `json:"empty"`
}

type holePunchImpl struct {
}

func NewHolePunchImpl() HolePunchHandler {
	return &holePunchImpl{}
}

func (h *holePunchImpl) HolePunching(localAddr string, localNAT int8, msg *rendezvous.NegotiationMessage, remoteAddr string) (*HolePunchResult, error) {

	switch h.checkHolePunchType(localNAT, msg.RemoteNATType) {
	case hard2hard:
		return nil, ErrNoSupportHard2Hard

	case easy2easy:
		handler, err := newE2EHandler(localAddr, remoteAddr, localNAT, msg.RemoteNATType)
		if err != nil {
			return nil, err
		}

		defer handler.Close()

		return handler.HolePunch()

	case easy2hard:
		handler, err := newE2HHandlerV2(localAddr, remoteAddr, localNAT, msg.RemoteNATType, msg.IsActive)
		if err != nil {
			return nil, err
		}

		defer handler.Close()

		return handler.HolePunch()

	default:
		return nil, ErrUnknownHolePunchType
	}
}

func (h *holePunchImpl) checkHolePunchType(localNAT int8, remoteNAT int8) int {
	if types.NATType(localNAT) == types.Symmetric && types.NATType(remoteNAT) == types.Symmetric {
		return hard2hard
	}

	if types.NATType(localNAT) != types.Symmetric && types.NATType(remoteNAT) != types.Symmetric {
		return easy2easy
	}

	if types.NATType(localNAT) == types.Symmetric && types.NATType(remoteNAT) != types.Symmetric {
		return easy2hard
	}

	if types.NATType(localNAT) != types.Symmetric && types.NATType(remoteNAT) == types.Symmetric {
		return easy2hard
	}

	return 0
}
