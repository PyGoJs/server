package handlers

import (
	"net/http"
	"strconv"

	"github.com/pygojs/server/types/classitem"
)

func ApiClassItem(w http.ResponseWriter, r *http.Request) {
	var cid int
	var err error

	var p pageError
	if cid, err = strconv.Atoi(r.FormValue("cid")); err != nil || cid <= 0 {
		p.ErrStr = "invalid cid (classid) " // Extra space for next if statement
	}

	var yr int
	if yr, err = strconv.Atoi(r.FormValue("yrwk")); err != nil || yr <= 0 {
		p.ErrStr += "invalid yearweek"
	}

	if p.ErrStr != "" {
		writeJSON(w, r, p)
		return
	}

	cis, err := classitem.FetchAll(cid, yr)
	if err != nil {
		p := pageError{
			ErrStr: "cannot fetch class items",
		}
		writeJSON(w, r, p)
		return
	}

	writeJSON(w, r, cis)
}
