package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type pageError struct {
	ErrInt int    `json:"errornum,omitempty"`
	ErrStr string `json:"error,omitempty"`
}

// https://sourcegraph.com/blog/google-io-2014-building-sourcegraph-a-large-scale-code-search-engine-in-go
// lm: Last Modified
func writeJSON(w http.ResponseWriter, r *http.Request, v interface{}, lm time.Time) error {

	data, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		return err
	}

	if lm != (time.Time{}) {
		if t, err := time.Parse(http.TimeFormat, r.Header.Get("If-Modified-Since")); err == nil && lm.Unix() <= t.Unix() {
			w.WriteHeader(http.StatusNotModified)
			return errors.New("not modified")
		}
		w.Header().Set("Last-Modified", lm.Format(http.TimeFormat))
	}

	w.Header().Set("content-type", "application/json; charset=utf-8")
	_, err = w.Write(data)
	return err
}
