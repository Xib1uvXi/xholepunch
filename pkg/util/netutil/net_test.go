package netutil

import (
	"github.com/stretchr/testify/require"
	"net"
	"testing"
)

func TestUDPRandListen(t *testing.T) {
	_, err := UDPRandListen()
	require.NoError(t, err)
}

func TestTCPRandListen(t *testing.T) {
	_, err := TCPRandListen()
	require.NoError(t, err)
}

type TestMessage struct {
	Ack   int8   `json:"ack"`
	Token string `json:"token"`
}

func TestConnSendMessage(t *testing.T) {
	cServer, err := TCPRandListen()
	require.NoError(t, err)

	var done = make(chan struct{})

	go func() {
		conn, err := cServer.Accept()
		require.NoError(t, err)
		var msg TestMessage
		err = ConnReceiveMessage(conn, &msg)
		require.NoError(t, err)
		require.Equal(t, int8(1), msg.Ack)
		require.Equal(t, "test-token123213123123213213213124vdsgvfdsgsfedfg", msg.Token)
		close(done)
	}()

	conn, err := net.Dial("tcp", cServer.Addr().String())
	require.NoError(t, err)

	err = ConnSendMessage(conn, &TestMessage{Ack: 1, Token: "test-token123213123123213213213124vdsgvfdsgsfedfg"})
	require.NoError(t, err)

	<-done
}
