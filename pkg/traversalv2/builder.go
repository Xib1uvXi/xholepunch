package traversalv2

import "github.com/Xib1uvXi/xholepunch/pkg/traversalv2/holepunch"

func BuilderDemo(serverAddr string, natType int8) (*Client, error) {
	hpHandler := holepunch.NewV1Impl()
	cf := &DemoConnectFactory{}

	return NewClient(serverAddr, natType, hpHandler, cf), nil
}
