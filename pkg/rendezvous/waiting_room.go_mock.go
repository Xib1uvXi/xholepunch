// Code generated by MockGen. DO NOT EDIT.
// Source: waiting_room.go
//
// Generated by this command:
//
//	mockgen -source=waiting_room.go -destination=waiting_room.go_mock.go -package=rendezvous
//

// Package rendezvous is a generated GoMock package.
package rendezvous

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockMeetingHandler is a mock of MeetingHandler interface.
type MockMeetingHandler struct {
	ctrl     *gomock.Controller
	recorder *MockMeetingHandlerMockRecorder
}

// MockMeetingHandlerMockRecorder is the mock recorder for MockMeetingHandler.
type MockMeetingHandlerMockRecorder struct {
	mock *MockMeetingHandler
}

// NewMockMeetingHandler creates a new mock instance.
func NewMockMeetingHandler(ctrl *gomock.Controller) *MockMeetingHandler {
	mock := &MockMeetingHandler{ctrl: ctrl}
	mock.recorder = &MockMeetingHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMeetingHandler) EXPECT() *MockMeetingHandlerMockRecorder {
	return m.recorder
}

// Meeting mocks base method.
func (m *MockMeetingHandler) Meeting(token string, conn1, conn2 *holePunchConn) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Meeting", token, conn1, conn2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Meeting indicates an expected call of Meeting.
func (mr *MockMeetingHandlerMockRecorder) Meeting(token, conn1, conn2 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Meeting", reflect.TypeOf((*MockMeetingHandler)(nil).Meeting), token, conn1, conn2)
}