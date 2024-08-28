package holepunch

import "github.com/Xib1uvXi/xholepunch/pkg/rendezvous"

type HolePunchHandler interface {
	HolePunching(localAddr string, localNAT int8, msg *rendezvous.NegotiationMessage, remoteAddr string) (*Result, error)
}
