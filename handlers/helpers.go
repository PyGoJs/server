package handlers

import (
	"encoding/json"
	"net/http"
)

type pageErrorStr struct {
	Error string `json:"error"`
}

type pageErrorNum struct {
	Error int `json:"error"`
}

// writeJSON writes the given struct (v) on the given http ResponseWriter in JSON format.
// A time object is used for determining last-modified (headers).
// https://sourcegraph.com/blog/google-io-2014-building-sourcegraph-a-large-scale-code-search-engine-in-go
func writeJSON(w http.ResponseWriter, r *http.Request, v interface{}) error {

	data, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		return err
	}

	/*
		if lm != (time.Time{}) {
			if t, err := time.Parse(http.TimeFormat, r.Header.Get("If-Modified-Since")); err == nil && lm.Unix() <= t.Unix() {
				w.WriteHeader(http.StatusNotModified)
				return errors.New("not modified")
			}
			w.Header().Set("Last-Modified", lm.Format(http.TimeFormat))
		}
	*/

	w.Header().Set("content-type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_, err = w.Write(data)
	return err
}
