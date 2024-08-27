package rendezvous

import (
	"github.com/Xib1uvXi/xholepunch/pkg/types"
	"github.com/Xib1uvXi/xholepunch/pkg/util/netutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"net"
	"testing"
	"time"
)

func TestNewFrontDesk(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConnecthandler := NewMockConnectHandler(ctrl)

	serverAddr := "127.0.0.1:7779"
	desk, err := NewFrontDesk(serverAddr, mockConnecthandler)
	require.NoError(t, err)
	defer desk.Close()

	desk.Serve()

	mockConnecthandler.EXPECT().HandleConnect(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	// client
	conn, err := net.Dial("tcp", serverAddr)
	require.NoError(t, err)
	defer conn.Close()

	msg := &ConnectMessage{
		Token:   "test-token",
		NATType: int8(types.Symmetric),
	}

	require.NoError(t, netutil.ConnSendMessage(conn, msg))

	time.Sleep(1 * time.Second)

}
