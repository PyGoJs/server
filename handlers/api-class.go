package handlers

import (
	"net/http"

	"github.com/pygojs/server/types/class"
)

func ApiClass(w http.ResponseWriter, r *http.Request) {
	cs, err := class.FetchAll()
	if err != nil {
		p := pageError{
			ErrStr: "not your fault",
		}
		writeJSON(w, r, p)
		return
	}
	writeJSON(w, r, cs)
}
