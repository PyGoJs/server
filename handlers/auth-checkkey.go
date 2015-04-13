package handlers

import (
	"net/http"

	"github.com/pygojs/server/auth"
)

type pageCheckKey struct {
	Valid bool       `json:"valid"`
	User  *auth.User `json:"user,omitempty"`
}

// AuthCheckKey checks whether or not the given session key is valid.
// Writes the user information if the session is valid.
func AuthCheckKey(w http.ResponseWriter, r *http.Request) {
	var u auth.User
	var err error
	if u, err = auth.CheckKey(r.FormValue("authkey")); err != nil {
		writeJSON(w, r, pageCheckKey{})
		return
	}

	p := pageCheckKey{
		Valid: true,
		User:  &u,
	}
	writeJSON(w, r, p)
}
