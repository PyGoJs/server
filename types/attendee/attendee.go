package att

import (
	"database/sql"
	"log"
	"strconv"
	"strings"

	"github.com/pygojs/server/types/classitem"
	"github.com/pygojs/server/types/student"
)

type Att struct {
	Id        int     `json:"-"`
	Ciid      int     `json:"ciid,omitempty"`
	Sid       int     `json:",omitempty"`
	Attent    bool    `json:"attent"`
	MinsEarly int     `sql:"mins_early" json:"minsEarly,omitempty"`
	Stu       stu.Stu `json:"stu"`
}

func Fetchs(ci classitem.ClassItem, db *sql.DB) ([]Att, error) {
	var atts []Att

	// Lamers tought us this!11
	rows, err := db.Query("SELECT attendee_item.id, attendee_item.mins_early, student.id, student.name, student.rfid FROM student, attendee_item WHERE attendee_item.ciid=? AND attendee_item.sid = student.id", ci.Id)
	if err != nil {
		log.Fatal(err)
	}

	sids := []string{"0"}

	for rows.Next() {
		var a Att
		var s stu.Stu
		err = rows.Scan(&a.Id, &a.MinsEarly, &s.Id, &s.Name, &s.Rfid)
		if err != nil {
			log.Fatal(err)
			return []Att{}, nil
		}
		a.Attent = true
		a.Stu = s
		atts = append(atts, a)
		sids = append(sids, strconv.Itoa(s.Id))
	}

	// Fetch the students that did not attent this.
	rows, err = db.Query("SELECT student.name, student.rfid FROM student, class_item WHERE class_item.id = ? AND class_item.cid = student.cid AND student.id NOT IN ("+strings.Join(sids, ",")+") LIMIT 1337;", ci.Id)
	for rows.Next() {
		var a Att
		var s stu.Stu
		err = rows.Scan(&s.Name, &s.Rfid)
		if err != nil {
			log.Fatal(err)
			return []Att{}, nil
		}
		a.Attent = false
		a.Stu = s
		atts = append(atts, a)
	}

	return atts, nil
}

func Attent(s stu.Stu, c classitem.ClassItem, minsEarly int, db *sql.DB) {
	_, err := db.Query("INSERT INTO attendee_item (ciid, sid, mins_early) VALUES (?, ?, ?);", c.Id, s.Id, minsEarly)
	if err != nil {
		log.Fatal(err)
	}
}
