package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/pygojs/server/handlers"
	"github.com/pygojs/server/types/client"
	"github.com/pygojs/server/util"
)

func main() {

	http.Handle("/checkin", logR(http.HandlerFunc(handlers.Checkin)))

	log.Println("Started")

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

	http.ListenAndServe(":13375", nil)

}

// logR logs a HTTP request
func logR(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		// Proxy stuff
		if ip == "127.0.0.1" {
			ip = r.Header.Get("X-FORWARDED-FOR")
		}
		fmt.Printf("%s %s %s%s\n", time.Now().Format("0102-15:04"), ip, r.Method, r.URL)
		h.ServeHTTP(w, r)
	})
}
