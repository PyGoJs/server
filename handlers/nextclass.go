package handlers

import (
	"net/http"
	"time"

	"github.com/pygojs/server/types/attendee"
	"github.com/pygojs/server/types/classitem"
	"github.com/pygojs/server/types/client"
	"github.com/pygojs/server/util"
)

type pageNextClass struct {
	classitem.ClassItem
	MinTillStart int       `json:"mintillstart"`
	Attendees    []att.Att `json:"attendees"`
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

	tn := time.Now()

	ci, err := classitem.NextCl(cl, tn)

	// Schedule item not found (no more classes for today)
	if err != nil {
		p := pageErrorNum{
			Error: 4,
		}
		writeJSON(w, r, p)
		return
	}

	atts, _ := att.FetchAll(ci)
	minTillStart := int(ci.Sched.Start.Sub(tn).Minutes())

	p := pageNextClass{ci, minTillStart, atts}

	writeJSON(w, r, p)
}
