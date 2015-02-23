package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/pygojs/server/util"

	"github.com/pygojs/server/types/attendee"
	"github.com/pygojs/server/types/classitem"
	"github.com/pygojs/server/types/student"
)

type pageCheckin struct {
	Accepted     bool      `json:"accepted"`
	Error        int       `json:"error,omitempty"`
	MinTillStart int       `json:"minTillStart,omitempty"`
	Attendees    []att.Att `json:"attendees,omitempty"`
}

func Checkin(w http.ResponseWriter, r *http.Request) {

	if sleep, err := strconv.Atoi(r.FormValue("sleep")); err == nil && sleep > 0 {
		if dur, err := time.ParseDuration(strconv.Itoa(sleep) + "ms"); err == nil {
			time.Sleep(dur)
		}
	}

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
	defer db.Close()

	s, err := stu.Fetch(rfid, db)

	// Student not found. Because of that don't know classItem so can't give minTillStart or attendees.
	if err != nil {
		if err != sql.ErrNoRows {
			writeJSON(w, r, pageError{ErrStr: "can't fetch student"}, time.Time{})
			return
		}

		p := pageCheckin{
			Accepted: false,
			Error:    1,
		}
		writeJSON(w, r, p, time.Time{})
		return
	}

	c, err := classitem.Fetch(s.Cid, time.Now(), db)

	// ClassItem not found (no class item for student?).
	if err != nil {
		writeJSON(w, r, pageError{ErrStr: "can't fetch schedule_item"}, time.Time{})
		return
	}

	minTillStart := int(c.Sched.Start.Sub(time.Now()).Minutes())

	// Too long until next class
	if minTillStart > 15 {
		p := pageCheckin{
			Accepted:     false,
			Error:        2,
			MinTillStart: minTillStart,
		}
		writeJSON(w, r, p, time.Time{})
		return
	}

	// TO DO: create class item if it doesn't exist (only do this if minTillStart !> 15, so not above)
	if r.FormValue("save") != "" {
		att.Attent(s, c, minTillStart, db)
	}

	p := pageCheckin{
		Accepted: true,
	}

	if r.FormValue("attendees") != "" {
		p.MinTillStart = minTillStart
		atts, _ := att.Fetchs(c, db)
		p.Attendees = atts
	}

	writeJSON(w, r, p, time.Time{})
}
