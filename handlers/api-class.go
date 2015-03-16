package handlers

import (
	"net/http"
	"time"

	"github.com/pygojs/server/types/class"
)

func ApiClass(w http.ResponseWriter, r *http.Request) {
	cs, err := class.FetchAll()
	if err != nil {
		p := pageError{
			ErrStr: "not your fault",
		}
		writeJSON(w, r, p, time.Time{})
		return
	}
	writeJSON(w, r, cs, time.Time{})
}
