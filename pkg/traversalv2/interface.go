package traversalv2

import (
	"github.com/Xib1uvXi/xholepunch/pkg/rendezvous"
	"github.com/Xib1uvXi/xholepunch/pkg/traversalv2/holepunch"
)

type ConnectFactory interface {
	Connect(localAddr string, remoteAddr string, isActive bool) error
}

type HolePunchHandler interface {
	HolePunching(localAddr string, localNAT int8, msg *rendezvous.NegotiationMessage, remoteAddr string) (*holepunch.Result, error)
}
