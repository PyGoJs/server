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
	"github.com/pygojs/server/types/schedule"
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

	// Time Now
	tn := time.Now()

	// Class
	c := class.Fetch(s.Cid)

	// Update the schedule if it was last fetched >=30 minutes ago.
	if tn.Sub(c.SchedLastFetched).Minutes() >= 30 {
		schedule.Update(c, tn, db)
	}

	// ClassItem, includes SchedItem. ClassItem can be empty, SchedItem may not.
	ci, err := classitem.Next(c, tn.AddDate(0, 0, -1), db)

	// ClassItem not found (no class item for student?).
	if err != nil {
		p := pageCheckin{
			Accepted: false,
			Error:    4,
		}
		writeJSON(w, r, p, time.Time{})
		return
	}

	// Check if this student is already attending this class_item.
	// When ci.Id is 0 nobody can be attending the class, so no need to check.
	if ci.Id != 0 {
		if id, _ := s.IsAttending(ci, db); id != 0 {
			p := pageCheckin{
				Accepted: false,
				Error:    3,
			}
			writeJSON(w, r, p, time.Time{})
			return
		}
	}

	minTillStart := int(ci.Sched.Start.Sub(tn).Minutes())

	if r.FormValue("save") != "" {
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

		// Create the classItem is there is none.
		if ci.Id == 0 {
			ci, _ = classitem.Create(ci.Sched, c, db)
		}
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
