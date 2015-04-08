package handlers

import (
	"net/http"

	"github.com/pygojs/server/auth"
)

type pageAuthCheckKey struct {
	Valid bool       `json:"valid"`
	User  *auth.User `json:"user,omitempty"`
}

func AuthCheckKey(w http.ResponseWriter, r *http.Request) {
	var u auth.User
	var err error
	if u, err = auth.CheckKey(r.FormValue("authkey")); err != nil {
		writeJSON(w, r, pageAuthCheckKey{})
		return
	}

	p := pageAuthCheckKey{
		Valid: true,
		User:  &u,
	}
	writeJSON(w, r, p)
}
