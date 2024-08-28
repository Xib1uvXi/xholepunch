package holepunch

import (
	"github.com/Xib1uvXi/xholepunch/pkg/rendezvous"
	"github.com/Xib1uvXi/xholepunch/pkg/types"
	"github.com/go-kratos/kratos/v2/errors"
)

var (
	ErrNoSupportHard2Hard   = errors.BadRequest("no support hard2hard", "no support hard2hard")
	ErrUnknownHolePunchType = errors.BadRequest("unknown hole punch type", "unknown hole punch type")
)

type v1Impl struct {
}

func NewV1Impl() HolePunchHandler {
	return &v1Impl{}
}

func (h *v1Impl) HolePunching(localAddr string, localNAT int8, msg *rendezvous.NegotiationMessage, remoteAddr string) (*Result, error) {
	if types.NATType(localNAT) == types.Symmetric && types.NATType(msg.RemoteNATType) == types.Symmetric {
		return nil, ErrNoSupportHard2Hard
	}

	// easy 2 easy
	if types.NATType(localNAT) != types.Symmetric && types.NATType(msg.RemoteNATType) != types.Symmetric {
		handler, err := NewEasy2Easy(localAddr, remoteAddr, localNAT, msg.RemoteNATType, msg.IsActive)
		if err != nil {
			return nil, err
		}

		defer handler.Close()

		return handler.HolePunch()
	}

	if types.NATType(localNAT) == types.Symmetric && types.NATType(msg.RemoteNATType) != types.Symmetric {
		handler, err := NewPortRestrictedCone2Symmetric(localAddr, remoteAddr, localNAT, msg.RemoteNATType, msg.IsActive)
		if err != nil {
			return nil, err
		}

		defer handler.Close()

		return handler.HolePunch()
	}

	if types.NATType(localNAT) != types.Symmetric && types.NATType(msg.RemoteNATType) == types.Symmetric {
		handler, err := NewPortRestrictedCone2Symmetric(localAddr, remoteAddr, localNAT, msg.RemoteNATType, msg.IsActive)
		if err != nil {
			return nil, err
		}

		defer handler.Close()

		return handler.HolePunch()
	}

	return nil, ErrUnknownHolePunchType
}
