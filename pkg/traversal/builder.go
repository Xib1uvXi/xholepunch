package traversal

import "github.com/go-kratos/kratos/v2/log"

type DemoConnectFactory struct{}

func (d *DemoConnectFactory) Connect(localAddr string, remoteAddr string, isActive bool) error {
	log.Infof("connect from %s to %s, is active %v", localAddr, remoteAddr, isActive)
	return nil
}

func Builder(serverAddr string, natType int8) (*Client, error) {
	hpHandler := NewHolePunchImpl()
	cf := &DemoConnectFactory{}

	return NewClient(serverAddr, natType, hpHandler, cf), nil
}

func BuilderV2(serverAddr string, natType int8) (*ClientV2, error) {
	hpHandler := NewHolePunchImpl()
	cf := &DemoConnectFactory{}

	return NewClientV2(serverAddr, natType, hpHandler, cf), nil
}
