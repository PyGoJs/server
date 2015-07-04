package main

// Used for checking a person in, without having to Curl.

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/pygojs/server/util"
)

func main() {
	name := flag.String("n", "", "name of student")
	clid := flag.String("c", "AjuSY$U9e", "client id")

	flag.Parse()

	assert(*name != "", "no name given")

	util.LoadConfig("config.json")
	util.CreateDb()

	var rfid string
	var err error
	err = util.Db.QueryRow("SELECT rfid FROM student WHERE name=? LIMIT 1;", *name).Scan(&rfid)
	assert(err == nil, err)

	url := fmt.Sprintf("http://dev.pygojs.remi.im/checkin?clientid=%s&rfid=%s", *clid, strings.Replace(rfid, " ", "%20", -1))
	fmt.Println(url)
	resp, err := http.Get(url)
	assert(err == nil, err)

	fmt.Println("Rfid tag:", "'"+rfid+"'")

	defer resp.Body.Close()
	cont, err := ioutil.ReadAll(resp.Body)
	assert(err == nil, err)

	fmt.Println(string(cont))
}

// assert makes sure the condition as b is true.
// If not true, it prints given error message and exits the program.
// (msg is interface to allow type error and string.)
// (err.Error() can be nil if no error)
func assert(b bool, msg interface{}) {
	if !b {
		log.Println("ERROR:", msg)
		os.Exit(1)
	}
}
