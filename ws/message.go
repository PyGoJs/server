package ws

import "github.com/pygojs/server/types/attendee"

type inMsg struct {
	Page string // Conn send to server that page changed
}

type OutMsg struct {
	Error   string `json:"error,omitempty"`
	Checkin struct {
		CiId int     `json:"ciid"`
		Att  att.Att `json:"att"`
	} `json:"checkin,omitempty"`
}
