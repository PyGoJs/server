package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/pygojs/server/handlers"
)

func main() {
	// https://gist.github.com/belbomemo/b5e7dad10fa567a5fe8a
	// " still preferred over ` because of escaping (there are like three(!) `'s in the string below).
	/*fmt.Println(
	"         ,_---~~~~~----._         \n" +
		"  _,,_,*^____      _____``*g*\"*, \n" +
		" / __/ /'     ^.  /      \\ ^@q   f\n" +
		"[  @f | @))    |  | @))   l  0 _/ \n" +
		" \\`/   \\~____ / __ \\_____/    \\   \n" +
		"  |           _l__l_           I  \n" +
		"  }          [______]           I \n" +
		"  ]            | | |            | \n" +
		"  ]             ~ ~             | \n" +
		"  |                            |  \n" +
		"   |                           |  \n")*/

	http.Handle("/checkin", logR(http.HandlerFunc(handlers.Checkin)))

	log.Println("Started")

	http.ListenAndServe(":13375", nil)

}

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
