package handlers

import (
	"net/http"

	"github.com/pygojs/server/types/class"
)

// ApiClass HTTP handler writes []class.Class containing all classes in the database.
func ApiClass(w http.ResponseWriter, r *http.Request) {
	cs, err := class.FetchAll()
	if err != nil {
		p := pageErrorStr{
			Error: "not your fault",
		}
		writeJSON(w, r, p)
		return
	}
	writeJSON(w, r, cs)
}
