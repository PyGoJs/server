package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/pygojs/server/util"

	"github.com/pygojs/server/types/attendee"
	"github.com/pygojs/server/types/class"
	"github.com/pygojs/server/types/classitem"
)

type pageCheckin struct {
	Accepted     bool      `json:"accepted"`
	Error        int       `json:"error,omitempty"`
	MinTillStart int       `json:"minTillStart,omitempty"`
	Attendees    []att.Att `json:"attendees,omitempty"`
}

// Checkin should be read as check-in (not as "I'm there like checkin' out this nub's project")
func Checkin(w http.ResponseWriter, r *http.Request) {

	// Sleep for given amount of milliseconds when get variable 'sleep' is set.
	// (Looks complicated (or just confusing) because of checking if valid int and time parsing.)
	if sleep, err := strconv.Atoi(r.FormValue("sleep")); err == nil && sleep > 0 {
		if dur, err := time.ParseDuration(strconv.Itoa(sleep) + "ms"); err == nil {
			time.Sleep(dur)
		}
	}

	var err error

	var clid int
	var rfid string
	if clid, err = strconv.Atoi(r.FormValue("clientid")); err != nil {
		writeJSON(w, r, pageError{ErrStr: "invalid clientid"}, time.Time{})
		return
	}

	if rfid = r.FormValue("rfid"); rfid == "" {
		writeJSON(w, r, pageError{ErrStr: "invalid rfid"}, time.Time{})
		return
	}

	_ = clid

	db, err := util.Db()
	if err != nil {
		writeJSON(w, r, pageError{ErrStr: "not your fault :( (database is down?)"}, time.Time{})
		return
	}
	defer db.Close()

	s, err := att.FetchStu(rfid, db)

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

	// Class
	c := class.Fetch(s.Cid)

	// ClassItem
	ci, err := classitem.Fetch(c, time.Now(), db)

	// ClassItem not found (no class item for student?).
	if err != nil {
		writeJSON(w, r, pageError{ErrStr: "can't fetch schedule_item"}, time.Time{})
		return
	}

	minTillStart := int(ci.Sched.Start.Sub(time.Now()).Minutes())

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
		/*if ci.Id == 0 {
			classitem.Create(ci.Sched, db)
		}*/
		att.Attent(s, ci, minTillStart, db)
	}

	p := pageCheckin{
		Accepted: true,
	}

	if r.FormValue("attendees") != "" {
		p.MinTillStart = minTillStart
		atts, _ := att.Fetchs(ci, db)
		p.Attendees = atts
	}

	writeJSON(w, r, p, time.Time{})
}
