package ws

import (
	"github.com/pygojs/server/types/attendee"
	"github.com/pygojs/server/types/lesson"
)

// inMsg is what incomming messages from conns are put in.
type inMsg struct {
	Page string // Conn send to server that page changed
}

// OutMsg is what outgoing messages are put in.
type OutMsg struct {
	dest struct { // Destination.
		page string `json:"-"` // Only send this msg to conns that are on this page.
	} `json:"-"`
	Error   string   `json:"error,omitempty"`
	Checkin struct { // Information about a check-in.
		//CiId int          `json:"ciid"`
		//Si   si.SchedItem `json:"si"`
		Ls  []lesson.Lesson `json:"ls"`
		Att att.Att         `json:"att"`
	} `json:"checkin,omitempty"`
}
