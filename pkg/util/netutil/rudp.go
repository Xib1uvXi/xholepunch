package netutil

import (
	"github.com/Doraemonkeys/reliableUDP"
	"github.com/Xib1uvXi/xholepunch/pkg/util/json"
	"net"
	"time"
)

// 发送超时时间为timeout,如果timeout为0则默认为4秒
func RUDPSendMessage(conn *reliableUDP.ReliableUDP, addr string, msg interface{}, timeout time.Duration) error {
	data, err := json.StringifyJsonToBytesWithErr(msg)
	if err != nil {
		return err
	}

	raddr, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		return err
	}
	err = conn.Send(raddr, data, timeout)
	if err != nil {
		return err
	}
	return nil
}

func RUDPSendUnreliableMessage(conn *reliableUDP.ReliableUDP, addr string, msg interface{}) error {
	data, err := json.StringifyJsonToBytesWithErr(msg)
	if err != nil {
		return err
	}

	raddr, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		return err
	}
	err = conn.SendUnreliable(data, raddr)
	if err != nil {
		return err
	}
	return nil
}

func RUDPReceiveAllMessage(conn *reliableUDP.ReliableUDP, timeout time.Duration, msg interface{}) (*net.UDPAddr, error) {
	data, addr, err := conn.ReceiveAll(timeout)
	if err != nil {
		return nil, err
	}

	err = json.ParseJsonFromBytes(data, msg)
	if err != nil {
		return nil, err
	}

	return addr, nil
}
