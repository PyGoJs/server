package att

import (
	"database/sql"
	"log"

	"github.com/pygojs/server/types/student"
)

type Att struct {
	Id        int     `json:"id"`
	Ciid      int     `json:"ciid"`
	Sid       int     `json:",omitempty"`
	MinsEarly int     `sql:"mins_early" json:"minsEarly"`
	Stu       stu.Stu `json:"stu"`
}

func Fetchs(ciid int, db *sql.DB) ([]Att, error) {
	var atts []Att

	// Lamers tought us this!11
	rows, err := db.Query("SELECT attendee_item.id, attendee_item.ciid, attendee_item.mins_early, student.id, student.name, student.rfid FROM student, attendee_item WHERE attendee_item.ciid=? AND attendee_item.sid = student.id", ciid)
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		var a Att
		var s stu.Stu
		err = rows.Scan(&a.Id, &a.Ciid, &a.MinsEarly, &s.Id, &s.Name, &s.Rfid)
		if err != nil {
			log.Fatal(err)
			return []Att{}, nil
		}
		a.Stu = s
		atts = append(atts, a)
	}

	return atts, nil
}

func Attent(s stu.Stu, ciid int, minsEarly int, db *sql.DB) {
	_, err := db.Query("INSERT INTO attendee_item (ciid, sid, mins_early) VALUES (?, ?, ?);", ciid, s.Id, minsEarly)
	if err != nil {
		log.Fatal(err)
	}
}
