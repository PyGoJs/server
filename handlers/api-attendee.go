package handlers

import (
	"net/http"
	"strconv"

	"github.com/pygojs/server/auth"
	"github.com/pygojs/server/types/attendee"
	"github.com/pygojs/server/types/classitem"
)

func ApiAttendee(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CheckKey(r.FormValue("authkey")); err != nil {
		p := pageError{
			ErrStr: "invalid authkey",
		}
		writeJSON(w, r, p)
		return
	}

	var ciid int

	if ciid, _ = strconv.Atoi(r.FormValue("ciid")); ciid == 0 {
		p := pageError{
			ErrStr: "invalid ciid (class item id)",
		}
		writeJSON(w, r, p)
		return
	}

	atts, err := att.FetchAll(classitem.ClassItem{Id: ciid})
	if err != nil {
		p := pageError{
			ErrStr: "cannot fetching attendees",
		}
		writeJSON(w, r, p)
		return
	}

	writeJSON(w, r, atts)

}
