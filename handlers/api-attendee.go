package handlers

import (
	"net/http"
	"strconv"

	"github.com/pygojs/server/auth"
	"github.com/pygojs/server/types/attendee"
)

// ApiAttendee HTTP handler writes []att.Att for the attendees, and students that could be/can attending,
// for the given ciid (class item id) (formvalue).
func ApiAttendee(w http.ResponseWriter, r *http.Request) {
	if _, err := auth.CheckKey(r.FormValue("authkey")); err != nil {
		p := pageErrorStr{
			Error: "invalid authkey",
		}
		writeJSON(w, r, p)
		return
	}

	var lid int

	if lid, _ = strconv.Atoi(r.FormValue("lid")); lid == 0 {
		p := pageErrorStr{
			Error: "invalid lid (l id)",
		}
		writeJSON(w, r, p)
		return
	}

	atts, err := att.FetchAll(0, lid)
	if err != nil {
		p := pageErrorStr{
			Error: "cannot fetching attendees",
		}
		writeJSON(w, r, p)
		return
	}

	writeJSON(w, r, atts)

}
