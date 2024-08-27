package rendezvous

import (
	"github.com/go-kratos/kratos/v2/log"
	"math/rand"
	"time"
)

type Server struct {
	meetingHandler MeetingHandler
	frontDesk      *FrontDesk
	waitingRoom    *WaitingRoomManager
}

func NewServer(meetingHandler MeetingHandler, frontDesk *FrontDesk, waitingRoom *WaitingRoomManager) *Server {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	return &Server{
		meetingHandler: meetingHandler,
		frontDesk:      frontDesk,
		waitingRoom:    waitingRoom,
	}
}

func (s *Server) Serve() {
	s.frontDesk.Serve()
}

func (s *Server) Close() {
	if err := s.frontDesk.Close(); err != nil {
		log.Errorf("front desk close error: %v", err)
	}

	if err := s.waitingRoom.Close(); err != nil {
		log.Errorf("waiting room close error: %v", err)
	}
}
