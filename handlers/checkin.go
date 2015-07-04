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
	"github.com/pygojs/server/types/client"
	"github.com/pygojs/server/types/lesson"
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

// Checkin (check-in) HTTP handler tries to check-in the student with the given rfid in the facility
// of the given clientid (form values). Check-ins are only really saved when save (form value) is given.
// pageCheckin is writen. MinTillStart and Attendees are only given when attendees (form value) is given.
func Checkin(w http.ResponseWriter, r *http.Request) {
	// This is not nice. Fix it.
	util.LogS("%s checkin: %s:%s:%s", util.Ip(*r), r.FormValue("clientid"),
		r.FormValue("rfid"), r.FormValue("nosave"))

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
		writeJSON(w, r, pageErrorStr{Error: "invalid clientid"})
		return
	}

	if rfid = r.FormValue("rfid"); rfid == "" {
		writeJSON(w, r, pageErrorStr{Error: "invalid rfid"})
		return
	}

	s, err := att.FetchStu(rfid)

	// Student not found. Because of that don't know classItem so can't give minTillStart or attendees.
	if err != nil {
		if err != sql.ErrNoRows {
			writeJSON(w, r, pageErrorStr{Error: "can't fetch student"})
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

	// Debug time from configuration file
	if util.Cfg().Debug.Enabled == true {
		tn = util.Cfg().Debug.Tm
	}

	// Fetch class of this student
	c, err := class.Fetch(s.Cid)

	if err != nil {
		writeJSON(w, r, pageErrorStr{Error: "can't fetch class"})
		return
	}

	// Update the schedule if it was last fetched >=15 minutes ago.
	//if tn.Sub(c.SchedLastFetched).Minutes() >= 15 {
	change, _ := schedule.Update(c, tn)
	fmt.Println(" Schedule updated,", change)
	//}

	// Lesson, includes SchedItem. Lesson can be empty, SchedItem may not.
	l, err := lesson.NextC(c, tn)

	// Lesson/ScheduleItem not found (no more classes today?).
	if err != nil {
		p := pageCheckin{
			Accepted: false,
			Error:    4,
		}
		writeJSON(w, r, p)
		return
	}

	fmt.Println(l)

	// Student not checking into the facility/room he or she should be attending.
	if l.Sched.Fac != cl.Fac {
		p := pageCheckin{
			Accepted: false,
			Error:    5,
		}
		writeJSON(w, r, p)
		return
	}

	// Check if this student is already attending this class_item.
	// When ci.Id is 0 nobody is (and can be) attending the class (up to now), so no need to check.
	if l.Id != 0 {
		if id, _ := s.IsAttending(l); id != 0 {
			p := pageCheckin{
				Accepted: false,
				Error:    3,
			}
			writeJSON(w, r, p)
			return
		}
	}

	minTillStart := int(l.Sched.Start.Sub(tn).Minutes())

	// LessonsAfter
	ls, _ := l.Afters(c)

	// NoSave for testing
	save := true
	if r.FormValue("nosave") != "" {
		save = false
	}

	if save {
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

		// Loop over the Lessons that this Stu should be checked-into.
		// (Think about doing this loop into ci.Create and s.Attent and just give a []ci and []s)
		for i, l := range ls {
			// Create the classItem if there is none.
			if l.Id == 0 {
				l.Create(c, tn)
				fmt.Println(" Lesson created, ", l.Id, l.MaxStus)
			}
			mts := 0
			if i == 0 {
				mts = minTillStart
			}

			// Make the Stu attent this class.
			lastId := s.Attent(l, mts)
			if lastId != 0 {
				l.MaxStus++
			}
		}
	}

	p := pageCheckin{
		Accepted: true,
	}

	// Give a list of attendees for this classItem if it is requested
	if r.FormValue("attendees") != "" {
		p.MinTillStart = minTillStart
		atts, _ := att.FetchAll(0, l.Id)
		p.Attendees = atts
	}

	if save {
		// Push message to website viewers
		var wsMsg ws.OutMsg
		// wsMsg.Checkin.CiId = ci.Id
		wsMsg.Checkin.Ls = ls
		wsMsg.Checkin.Att.Items = append(wsMsg.Checkin.Att.Items, att.AttItem{
			MinsEarly: minTillStart,
		})
		s.Rfid = ""
		wsMsg.Checkin.Att.Stu = s
		ws.Wss.Broadcast(strings.ToLower(c.Name), wsMsg)
	}

	writeJSON(w, r, p)
}
