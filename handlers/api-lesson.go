package handlers

import (
	"net/http"
	"strconv"

	"github.com/pygojs/server/types/lesson"
)

// ApiClassItem HTTP handler writes []classitem.ClassItem for the given cid (class id)
// and yrwk (year week) (form values).
func ApiClassItem(w http.ResponseWriter, r *http.Request) {
	var cid int
	var err error

	var p pageErrorStr
	if cid, err = strconv.Atoi(r.FormValue("cid")); err != nil || cid <= 0 {
		p.Error = "invalid cid (classid) " // Extra space for next if statement
	}

	var yr int
	if yr, err = strconv.Atoi(r.FormValue("yrwk")); err != nil || yr <= 0 {
		p.Error += "invalid yearweek"
	}

	if p.Error != "" {
		writeJSON(w, r, p)
		return
	}

	ls, err := lesson.FetchAll(cid, yr)
	if err != nil {
		p := pageErrorStr{
			Error: "cannot fetch class items",
		}
		writeJSON(w, r, p)
		return
	}

	writeJSON(w, r, ls)
}
