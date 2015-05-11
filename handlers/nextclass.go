package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/pygojs/server/types/attendee"
	"github.com/pygojs/server/types/class"
	"github.com/pygojs/server/types/client"
	"github.com/pygojs/server/types/lesson"
	"github.com/pygojs/server/util"
)

type pageNextLesson struct {
	lesson.Lesson
	MinTillStart int         `json:"mintillstart"`
	Class        class.Class `json:"class"`
	Attendees    []att.Att   `json:"attendees"`
}

type pageNextClass struct {
	Items []pageNextLesson `json:"items"`
}

// NextClass HTTP handler writes pageNextClass as JSON,
// containing the information for the next class given in the facility of the given clientid (formvalue).
// Attendees will be null if no class-item has been created yet (no-ane has check-into that class yet).
func NextClass(w http.ResponseWriter, r *http.Request) {
	util.LogS("%s nextclass: %s", util.Ip(*r), r.FormValue("clientid"))

	var ok bool
	var cl client.Client

	if cl, ok = client.Get(r.FormValue("clientid")); ok == false {
		writeJSON(w, r, pageErrorStr{Error: "invalid clientid"})
		return
	}

	// Current time
	tn := time.Now()

	// Debug time from config file
	if util.Cfg().Debug.Enabled == true {
		tn = util.Cfg().Debug.Tm
	}

	ls, err := lesson.NextCl(cl, tn)

	// Schedule item not found (no more classes for today)
	if err != nil {
		p := pageErrorNum{
			Error: 4,
		}
		writeJSON(w, r, p)
		return
	}

	var items []pageNextLesson

	// Loop over the class/schedule-items.
	// There might be multiple classes in the same facility
	// (when facility is "" for example).
	for _, l := range ls {

		minTillStart := int(l.Sched.Start.Sub(tn).Minutes())

		c, _ := class.Fetch(l.Sched.Cid)

		lid := l.Id
		cid := 0
		if l.Id == 0 {
			cid = c.Id
		}

		atts, _ := att.FetchAll(cid, lid)

		fmt.Println(len(atts))

		item := pageNextLesson{
			l,
			minTillStart,
			c,
			atts,
		}

		items = append(items, item)
	}

	// p := pageNextClass{items}

	writeJSON(w, r, items)
}
