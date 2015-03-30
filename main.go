package main

import (
	"log"
	"net/http"

	"github.com/pygojs/server/auth"
	"github.com/pygojs/server/handlers"
	"github.com/pygojs/server/types/client"
	"github.com/pygojs/server/util"
	"github.com/pygojs/server/ws"
)

func main() {

	log.Println("Started")

	// Handlers
	http.Handle("/checkin", http.HandlerFunc(handlers.Checkin))

	// Handlers - Api
	http.Handle("/api/class", logR(http.HandlerFunc(handlers.ApiClass)))
	http.Handle("/api/classitem", logR(http.HandlerFunc(handlers.ApiClassItem)))
	http.Handle("/api/attendee", logR(http.HandlerFunc(handlers.ApiAttendee)))

	// Handlers - Auth
	http.Handle("/auth/login", logR(http.HandlerFunc(handlers.AuthLogin)))

	err := util.LoadConfig("config.json")
	// LoadConfig (and lots of other methods) logs the error.
	if err != nil {
		return
	}

	_, err = util.CreateDb()
	if err != nil {
		return
	}

	client.UpdateCache()

	// Create/start the websocket server and set the global var in ws to it
	// (so other stuff can do stuff with ws (see near the end of handlers.Checkin)).
	ws.Wss = ws.NewServer("/ws")

	go auth.Run()

	http.ListenAndServe(util.Cfg().Http.Addr, nil)

}

// logR logs a HTTP request
func logR(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		util.LogS("%s %s%s", util.Ip(*r), r.Method, r.URL)
		h.ServeHTTP(w, r)
	})
}
