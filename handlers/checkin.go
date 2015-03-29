package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pygojs/server/types/attendee"
	"github.com/pygojs/server/types/class"
	"github.com/pygojs/server/types/classitem"
	"github.com/pygojs/server/types/client"
	"github.com/pygojs/server/types/schedule"
	"github.com/pygojs/server/util"
	"github.com/pygojs/server/ws"
)

type pageCheckin struct {
	Accepted     bool      `json:"accepted"`
	Error        int       `json:"error,omitempty"`
	MinTillStart int       `json:"minTillStart,omitempty"`
	Attendees    []att.Att `json:"attendees,omitempty"`
}

// Checkin should be read as check-in (not as "I'm here like checkin' out this nub's project")
func Checkin(w http.ResponseWriter, r *http.Request) {
	// This is not nice. Fix it.
	util.LogS("%s checkin: %s:%s:%s:%s", util.Ip(*r), r.FormValue("clientid"), r.FormValue("rfid"), r.FormValue("save"), r.FormValue("attendees"))

	// Sleep for given amount of milliseconds when get variable 'sleep' is set.
	// (Looks complicated (or just confusing) because of checking if valid int and time parsing.)
	if sleep, err := strconv.Atoi(r.FormValue("sleep")); err == nil && sleep > 0 {
		if dur, err := time.ParseDuration(strconv.Itoa(sleep) + "ms"); err == nil {
			time.Sleep(dur)
		}
	}

	var err error

	var cl client.Client
	var rfid string

	var ok bool

	if cl, ok = client.Get(r.FormValue("clientid")); ok == false {
		writeJSON(w, r, pageError{ErrStr: "invalid clientid"})
		return
	}

	if rfid = r.FormValue("rfid"); rfid == "" {
		writeJSON(w, r, pageError{ErrStr: "invalid rfid"})
		return
	}

	s, err := att.FetchStu(rfid)

	// Student not found. Because of that don't know classItem so can't give minTillStart or attendees.
	if err != nil {
		if err != sql.ErrNoRows {
			writeJSON(w, r, pageError{ErrStr: "can't fetch student"})
			return
		}

		p := pageCheckin{
			Accepted: false,
			Error:    1,
		}
		writeJSON(w, r, p)
		return
	}

	// Time Now
	tn := time.Now()
	// tn = time.Date(2015, 3, 23, 10, 20, 0, 0, util.Loc)

	// Fetch class of this student
	c, err := class.Fetch(s.Cid)

	if err != nil {
		writeJSON(w, r, pageError{ErrStr: "can't fetch class"})
		return
	}

	// Update the schedule if it was last fetched >=30 minutes ago.
	if tn.Sub(c.SchedLastFetched).Minutes() >= 30 {
		change, _ := schedule.Update(c, tn)
		fmt.Println(" Schedule updated,", change)
	}

	// ClassItem, includes SchedItem. ClassItem can be empty, SchedItem may not.
	ci, err := classitem.Next(c, tn)

	// ClassItem/ScheduleItem not found (no more classes today?).
	if err != nil {
		p := pageCheckin{
			Accepted: false,
			Error:    4,
		}
		writeJSON(w, r, p)
		return
	}

	// Student not checking into the facility/room he or she should be attending.
	if ci.Sched.Fac != "" && ci.Sched.Fac != cl.Fac {
		p := pageCheckin{
			Accepted: false,
			Error:    5,
		}
		writeJSON(w, r, p)
		return
	}

	// Check if this student is already attending this class_item.
	// When ci.Id is 0 nobody is (and can be) attending the class (up to now), so no need to check.
	if ci.Id != 0 {
		if id, _ := s.IsAttending(ci); id != 0 {
			p := pageCheckin{
				Accepted: false,
				Error:    3,
			}
			writeJSON(w, r, p)
			return
		}
	}

	minTillStart := int(ci.Sched.Start.Sub(tn).Minutes())

	// For easier testing
	if r.FormValue("save") != "" {
		// Too long until next class
		if minTillStart > 15 {
			p := pageCheckin{
				Accepted:     false,
				Error:        2,
				MinTillStart: minTillStart,
			}
			writeJSON(w, r, p)
			return
		}

		// Create the classItem if there is none.
		if ci.Id == 0 {
			ci, _ = classitem.Create(ci.Sched, c, tn)
			fmt.Println(" ClassItem created, ", ci.Id, ci.MaxStus)
		}
		lastId := att.Attent(s, ci, minTillStart)
		if lastId != 0 {
			ci.MaxStus++
		}
	}

	p := pageCheckin{
		Accepted: true,
	}

	if r.FormValue("attendees") != "" {
		p.MinTillStart = minTillStart
		atts, _ := att.FetchAll(ci)
		p.Attendees = atts
	}

	var wsMsg ws.OutMsg
	// wsMsg.Checkin.CiId = ci.Id
	wsMsg.Checkin.Ci = ci
	wsMsg.Checkin.Att.MinsEarly = minTillStart
	s.Rfid = ""
	wsMsg.Checkin.Att.Stu = s
	ws.Wss.Broadcast(strings.ToLower(c.Name), wsMsg)

	writeJSON(w, r, p)
}
