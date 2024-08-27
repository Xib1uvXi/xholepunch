package netutil

import (
	"fmt"
	"github.com/Xib1uvXi/xholepunch/pkg/util/json"
	"github.com/go-kratos/kratos/v2/log"
	"math/rand"
	"net"
)

const (
	randPortStart = 20000
	randPortInc   = 10000

	maxPayloadSize = 1024 * 1024
)

func UdpSendMessage(conn *net.UDPConn, target *net.UDPAddr, data interface{}) error {
	payload, err := json.StringifyJsonToBytesWithErr(data)
	if err != nil {
		log.Errorf("stringify json error: %v", err)
		return err
	}

	_, err = conn.WriteToUDP(payload, target)

	if err != nil {
		log.Errorf("write message error: %v", err)
		return err
	}

	return nil
}
func UdpSendByteMessage(conn *net.UDPConn, target *net.UDPAddr, data []byte) error {
	_, err := conn.WriteToUDP(data, target)

	if err != nil {
		log.Errorf("write message error: %v", err)
		return err
	}

	return nil
}

func UdpReceiveMessage(conn *net.UDPConn, data interface{}) (*net.UDPAddr, error) {
	payload := make([]byte, maxPayloadSize)
	n, addr, err := conn.ReadFromUDP(payload)
	if err != nil {
		log.Errorf("read message error: %v", err)
		return nil, err
	}

	err = json.ParseJsonFromBytes(payload[:n], data)
	if err != nil {
		log.Errorf("parse json error: %v", err)
		return nil, err
	}

	return addr, nil
}

func ConnSendMessage(conn net.Conn, data interface{}) error {
	payload, err := json.StringifyJsonToBytesWithErr(data)
	if err != nil {
		log.Errorf("stringify json error: %v", err)
		return err
	}

	_, err = conn.Write(payload)
	if err != nil {
		log.Errorf("write message error: %v", err)
		return err
	}

	return nil
}

func ConnReceiveMessage(conn net.Conn, data interface{}) error {
	payload := make([]byte, maxPayloadSize)
	n, err := conn.Read(payload)
	if err != nil {
		log.Errorf("read message error: %v", err)
		return err
	}

	err = json.ParseJsonFromBytes(payload[:n], data)
	if err != nil {
		log.Errorf("%s parse json error: %v", string(payload[:n]), err)
		return err
	}

	return nil
}

func UDPRandListen() (*net.UDPConn, error) {
	randPort := rand.Intn(randPortStart) + randPortInc
	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf(":%d", randPort))

	if err != nil {
		return nil, err
	}

	udpConn, err := net.ListenUDP("udp4", addr)

	if err != nil {
		return nil, err
	}

	return udpConn, nil
}

func TCPRandListen() (*net.TCPListener, error) {
	randPort := rand.Intn(randPortStart) + randPortInc
	addr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf(":%d", randPort))

	if err != nil {
		return nil, err
	}

	tcpConn, err := net.ListenTCP("tcp4", addr)

	if err != nil {
		return nil, err
	}

	return tcpConn, nil
}
