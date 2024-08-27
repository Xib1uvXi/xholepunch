package rendezvous

import (
	"fmt"
	"github.com/Xib1uvXi/xholepunch/pkg/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestNewWaitingRoomManager(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMeetingHandler := NewMockMeetingHandler(ctrl)

	wrm, err := NewWaitingRoomManager(mockMeetingHandler, time.Second, 2*time.Second)
	require.NoError(t, err)
	defer wrm.Close()
}

func TestWaitingRoomManager_CreateWaitingRoom(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMeetingHandler := NewMockMeetingHandler(ctrl)

	wrm, err := NewWaitingRoomManager(mockMeetingHandler, time.Second, 2*time.Second)
	require.NoError(t, err)
	defer wrm.Close()

	mockConn := NewMockConn(ctrl)

	mockConn.EXPECT().Close().Times(1)

	token := "test-token"

	require.NoError(t, wrm.CreateWaitingRoom(mockConn, token, types.FullConeNATType))

	time.Sleep(3 * time.Second)

	require.False(t, wrm.tokenMap.Contains(token))
}

func TestWaitingRoomManager_JoinWaitingRoom(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMeetingHandler := NewMockMeetingHandler(ctrl)
	wrm, err := NewWaitingRoomManager(mockMeetingHandler, time.Second, 2*time.Second)
	require.NoError(t, err)
	defer wrm.Close()

	mockConn := NewMockConn(ctrl)
	mockConn.EXPECT().Close().Times(1)

	token := "test-token"

	require.NoError(t, wrm.CreateWaitingRoom(mockConn, token, types.FullConeNATType))

	mockConn2 := NewMockConn(ctrl)
	mockConn2.EXPECT().Close().Times(1)
	mockMeetingHandler.EXPECT().Meeting(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(fmt.Errorf("ssss"))
	require.NoError(t, wrm.JoinWaitingRoom(token, mockConn2, types.FullConeNATType))

	time.Sleep(3 * time.Second)
}
