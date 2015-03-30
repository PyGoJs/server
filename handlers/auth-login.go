package handlers

import (
	"net/http"

	"github.com/pygojs/server/auth"
)

type pageLogin struct {
	User auth.User `json:"user"`
	Key  string    `json:"key"`
}

func AuthLogin(w http.ResponseWriter, r *http.Request) {
	login := r.FormValue("login")
	pass := r.FormValue("password")

	if login == "" || pass == "" {
		p := pageError{
			ErrStr: "form value login and password required",
		}
		writeJSON(w, r, p)
		return
	}

	var u auth.User
	var key string
	var err error
	if u, key, err = auth.Login(login, pass); err != nil {
		p := pageError{
			ErrStr: "invalid login or password",
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
		p := pageError{
			ErrStr: "invalid authkey",
		}
		writeJSON(w, r, p)
		return
	}*/
}
