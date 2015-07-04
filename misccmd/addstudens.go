package main

// Add list of students to the database, without having to SQL manually.

import (
	"fmt"
	"strings"

	"github.com/pygojs/server/util"
)

var st00f = `Name1 Surname1
Name2 Surname2`

func main() {
	util.LoadConfig("config.json")
	util.CreateDb()
	for _, n := range strings.Split(st00f, "\n") {
		n2 := strings.Split(n, " ")
		first := n2[0]
		last := string([]rune(n2[1])[0])
		name := first + " " + last
		rfid := strings.ToLower(first+last) + "rfid"
		fmt.Println(name, rfid)

		// fmt.Println(util.Db.Query("INSERT INTO student (name, rfid, cid, createdyrwk) VALUES (?, ?, 1, 201516);", name, rfid))
	}
}
