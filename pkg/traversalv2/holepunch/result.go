package holepunch

import (
	"fmt"
	"net"
)

type Result struct {
	LocalAddr  string `json:"local_addr"`
	RemoteAddr string `json:"remote_addr"`
	LocalNAT   int8   `json:"local_nat"`
	RemoteNAT  int8   `json:"remote_nat"`
}

func (r Result) String() string {
	return fmt.Sprintf("local addr: %s, remote addr: %s, local nat: %d, remote nat: %d", r.LocalAddr, r.RemoteAddr, r.LocalNAT, r.RemoteNAT)
}

func (r *Result) LocalIPAndPort() (string, int, error) {
	localAddr, err := net.ResolveUDPAddr("udp4", r.LocalAddr)
	if err != nil {
		return "", 0, err
	}

	return localAddr.IP.String(), localAddr.Port, nil
}

func (r *Result) RemoteIPAndPort() (string, int, error) {
	remoteAddr, err := net.ResolveUDPAddr("udp4", r.RemoteAddr)
	if err != nil {
		return "", 0, err
	}

	return remoteAddr.IP.String(), remoteAddr.Port, nil
}
