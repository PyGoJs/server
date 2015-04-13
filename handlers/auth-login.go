package handlers

import (
	"net/http"

	"github.com/pygojs/server/auth"
)

type pageLogin struct {
	User auth.User `json:"user"`
	Key  string    `json:"key"`
}

// AuthLogin creates a new session for the given login/password pair, if it is valid.
// The session key and the user information is written.
func AuthLogin(w http.ResponseWriter, r *http.Request) {
	login := r.FormValue("login")
	pass := r.FormValue("password")

	if login == "" || pass == "" {
		p := pageErrorStr{
			Error: "form value login and password required",
		}
		writeJSON(w, r, p)
		return
	}

	var u auth.User
	var key string
	var err error
	if u, key, err = auth.Login(login, pass); err != nil {
		p := pageErrorStr{
			Error: "invalid login or password",
		}
		writeJSON(w, r, p)
		return
	}

	p := pageLogin{
		User: u,
		Key:  key,
	}
	writeJSON(w, r, p)

	/*if u, err := auth.CheckKey(r.FormValue("authkey")); err != nil {
		p := pageErrorStr{
			Error: "invalid authkey",
		}
		writeJSON(w, r, p)
		return
	}*/
}
