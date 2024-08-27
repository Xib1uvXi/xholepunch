package rendezvous

import (
	"fmt"
	"github.com/Doraemonkeys/reliableUDP"
	"github.com/Xib1uvXi/xholepunch/pkg/types"
	"github.com/Xib1uvXi/xholepunch/pkg/util/netutil"
	"github.com/go-kratos/kratos/v2/log"
	"net"
	"sync"
	"time"
)

type meetingImplV2 struct {
}

func NewMeetingImplV2() MeetingHandler {
	return &meetingImplV2{}

}

func (m *meetingImplV2) Meeting(token string, conn1, conn2 *holePunchConn) error {
	var conn1Active bool
	// check is active
	if types.NATType(conn1.natType) == types.Symmetric && types.NATType(conn2.natType) != types.Symmetric {
		conn1Active = true
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	var addr1, addr2 string

	ch1, err := m.getMeetingAddr(conn1, conn2, conn1Active)
	if err != nil {
		log.Errorf("get meeting addr error: %v", err)
		return err
	}

	ch2, err := m.getMeetingAddr(conn2, conn1, !conn1Active)
	if err != nil {
		log.Errorf("get meeting addr error: %v", err)
		return err
	}

	go func() {
		defer wg.Done()
		select {
		case addr1 = <-ch1:
			log.Infof("conn1 addr: %s", addr1)
		case <-time.After(10 * time.Second):
			log.Errorf("get conn1 meeting udp addr timeout")
		}
	}()

	go func() {
		defer wg.Done()
		select {
		case addr2 = <-ch2:
			log.Infof("conn2 addr: %s", addr2)
		case <-time.After(10 * time.Second):
			log.Errorf("get conn2 meeting udp addr timeout")
		}
	}()

	wg.Wait()

	if addr1 == "" || addr2 == "" {
		log.Errorf("get udp addr error, addr1 %s addr2 %s", addr1, addr2)
		return fmt.Errorf("get udp addr error")
	}

	if conn1Active {
		if err := m.startPunching(conn2, addr1); err != nil {
			log.Errorf("start punching error: %v", err)
			return err
		}

		if err := m.startPunching(conn1, addr2); err != nil {
			log.Errorf("start punching error: %v", err)
			return err
		}
	} else {
		if err := m.startPunching(conn1, addr2); err != nil {
			log.Errorf("start punching error: %v", err)
			return err
		}

		if err := m.startPunching(conn2, addr1); err != nil {
			log.Errorf("start punching error: %v", err)
			return err
		}
	}

	return nil
}

// getMeetingAddr
func (m *meetingImplV2) getMeetingAddr(conn1, conn2 *holePunchConn, isActive bool) (chan string, error) {
	var addr = make(chan string)

	serverPort, err := m.getUDPAddr(addr)
	if err != nil {
		log.Errorf("get udp addr error: %v", err)
		return nil, err
	}

	if err := m.negotiation(conn1, conn2, serverPort, isActive); err != nil {
		log.Errorf("negotiation error: %v", err)
		return nil, err
	}

	return addr, nil
}

// getUDPAddr
func (m *meetingImplV2) getUDPAddr(addrCh chan string) (int, error) {
	tempUdpConn, err := netutil.UDPRandListen()
	if err != nil {
		log.Debug("listen udp error", err)
		return 0, err
	}

	tempRudpConn := reliableUDP.New(tempUdpConn)
	tempRudpConn.SetGlobalReceive()

	randPort := tempUdpConn.LocalAddr().(*net.UDPAddr).Port

	go func() {
		var checkinMsg CheckinMessage
		addr, err := netutil.RUDPReceiveAllMessage(tempRudpConn, 5*time.Second, &checkinMsg)
		if err != nil {
			log.Debug("receive udp message error", err)
			tempRudpConn.Close()
			return
		}

		addrCh <- addr.String()
		tempRudpConn.Close()
	}()

	return randPort, nil
}

// Negotiation
func (m *meetingImplV2) negotiation(conn *holePunchConn, peerConn *holePunchConn, serverPort int, IsActive bool) error {
	msg := &NegotiationMessage{
		LocalPublicAddr:  conn.conn.RemoteAddr().String(),
		RemotePublicAddr: peerConn.conn.RemoteAddr().String(),
		RemoteNATType:    int8(peerConn.natType),
		ServerPort:       serverPort,
		IsActive:         IsActive,
	}

	if err := netutil.ConnSendMessage(conn.conn, msg); err != nil {
		log.Errorf("send negotiation message error: %v", err)
		return err
	}

	return nil
}

// startPunching
func (m *meetingImplV2) startPunching(conn1 *holePunchConn, addr2 string) error {
	if err := netutil.ConnSendMessage(conn1.conn, &HolePunchMessage{Addr: addr2}); err != nil {
		log.Errorf("send hole punch message error: %v", err)
		return err
	}

	return nil
}
