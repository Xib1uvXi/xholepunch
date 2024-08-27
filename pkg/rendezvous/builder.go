package rendezvous

import "time"

func Builder(listenAddr string) (*Server, error) {
	meeting := NewMeetingImpl()

	wrm, err := NewWaitingRoomManager(meeting, 1*time.Minute, 3*time.Minute)
	if err != nil {
		return nil, err
	}

	frontDesk, err := NewFrontDesk(listenAddr, wrm)
	if err != nil {
		return nil, err
	}

	return NewServer(meeting, frontDesk, wrm), nil
}
