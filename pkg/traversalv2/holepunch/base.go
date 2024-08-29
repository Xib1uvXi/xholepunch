package holepunch

import (
	"github.com/Doraemonkeys/reliableUDP"
	"net"
	"time"
)

const udpTimeout = 10 * time.Second

type HolePunchMessage struct {
	Empty int8 `json:"empty"`
}

type BaseModel struct {
	udpConn    *net.UDPConn
	localAddr  *net.UDPAddr
	remoteAddr *net.UDPAddr
	localNAT   int8
	remoteNAT  int8
	lAddr      string
	rAddr      string
	isActive   bool
	rudpConn   *reliableUDP.ReliableUDP
}

func newBaseModel(lAddr, rAddr string, localNAT, remoteNAT int8, isActive bool) (*BaseModel, error) {
	localAddr, err := net.ResolveUDPAddr("udp4", lAddr)
	if err != nil {
		return nil, err
	}

	remoteAddr, err := net.ResolveUDPAddr("udp4", rAddr)
	if err != nil {
		return nil, err
	}

	udpConn, err := net.ListenUDP("udp4", localAddr)
	if err != nil {
		return nil, err
	}

	bm := &BaseModel{
		udpConn:    udpConn,
		localAddr:  localAddr,
		remoteAddr: remoteAddr,
		localNAT:   localNAT,
		remoteNAT:  remoteNAT,
		lAddr:      lAddr,
		rAddr:      rAddr,
		isActive:   isActive,
		rudpConn:   reliableUDP.New(udpConn),
	}

	bm.rudpConn.SetGlobalReceive()

	return bm, nil
}

func (bm *BaseModel) Close() {
	if bm.udpConn != nil {
		bm.udpConn.Close()
	}

	if bm.rudpConn != nil {
		time.Sleep(5 * time.Second)
		bm.rudpConn.Close()
	}
}

func (bm *BaseModel) WaitFor3RTT(rtt time.Duration) {
	for i := 0; i < 3; i++ {
		time.Sleep(rtt)
	}
}
