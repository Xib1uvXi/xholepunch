package traversalv2

type ConnectFactory interface {
	Connect(localAddr string, remoteAddr string, isActive bool) error
}
