package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/pygojs/server/util"

	"github.com/pygojs/server/types/attendee"
	"github.com/pygojs/server/types/student"
)

type pageCheckin struct {
	Accepted     bool      `json:"accepted"`
	Error        int       `json:"error,omitempty"`
	MinTillStart int       `json:"minTillStart,omitempty"`
	Attendees    []att.Att `json:"attendees,omitempty"`
}

// Hardcoding the ciid until package class exists
const ciid = 1

func Checkin(w http.ResponseWriter, r *http.Request) {

	var err error

	var cid int
	var rfid string
	if cid, err = strconv.Atoi(r.FormValue("clientid")); err != nil {
		writeJSON(w, r, pageError{ErrStr: "invalid clientid"}, time.Time{})
		return
	}

	if rfid = r.FormValue("rfid"); rfid == "" {
		writeJSON(w, r, pageError{ErrStr: "invalid rfid"}, time.Time{})
		return
	}

	_ = cid

	db, err := util.Db()
	if err != nil {
		writeJSON(w, r, pageError{ErrStr: "not your fault :( (database is down?)"}, time.Time{})
		return
	}

	s, err := stu.Fetch(rfid, db)

	// Student not found. Because of that don't know classItem so can't give minTillStart or attendees.
	if err != nil {
		/*if err != sql.ErrNoRows {
			p := pageError{
				ErrStr: "database error again"
			}
			writeJSON(w, r, p, time.Time{})
			return
		}*/
		p := pageCheckin{
			Accepted: false,
			Error:    1,
		}
		writeJSON(w, r, p, time.Time{})
		return
	}

	_ = s
	//att.Attent(s, ciid, -1337, db)

	p := pageCheckin{
		Accepted: true,
	}

	if r.FormValue("attendees") != "" {
		p.MinTillStart = -1337
		atts, _ := att.Fetchs(ciid, db)
		p.Attendees = atts
	}

	writeJSON(w, r, p, time.Time{})
}
