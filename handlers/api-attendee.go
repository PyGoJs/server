package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/pygojs/server/types/attendee"
	"github.com/pygojs/server/types/classitem"
)

func ApiAttendee(w http.ResponseWriter, r *http.Request) {
	var ciid int

	if ciid, _ = strconv.Atoi(r.FormValue("ciid")); ciid == 0 {
		p := pageError{
			ErrStr: "invalid ciid (class item id)",
		}
		writeJSON(w, r, p, time.Time{})
		return
	}

	atts, err := att.FetchAll(classitem.ClassItem{Id: ciid})
	if err != nil {
		p := pageError{
			ErrStr: "cannot fetching attendees",
		}
		writeJSON(w, r, p, time.Time{})
		return
	}

	writeJSON(w, r, atts, time.Time{})

}
