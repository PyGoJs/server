package main

import (
	"log"
	"net/http"

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
	http.HandleFunc("/checkin", handlers.Checkin)

	log.Println("Started")

	http.ListenAndServe(":13375", nil)

}
