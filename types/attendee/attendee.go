package att

import (
	"database/sql"
	"log"
	"strconv"
	"strings"

	"github.com/pygojs/server/types/classitem"
)

// Attendee
type Att struct {
	Id        int  `json:"-"`
	Ciid      int  `json:"ciid,omitempty"`
	Sid       int  `json:",omitempty"`
	Attent    bool `json:"attent"`
	MinsEarly int  `sql:"mins_early" json:"minsEarly,omitempty"`
	Stu       Stu  `json:"stu"`
}

// Student
type Stu struct {
	Id   int    `json:"-"`
	Name string `json:"name"`
	Rfid string `json:"rfid"`
	Cid  int    `json:"-"`
}

// Fetch returns the student with the given rfid.
func FetchStu(rfid string, db *sql.DB) (Stu, error) {
	var s Stu

	/*err := db.QueryRow(`
	SELECT s.id, s.name, s.rfid, s.cid, c.name FROM student AS s, class AS c
	WHERE s.rfid=? AND s.cid = c.id LIMIT 1;
		`, rfid).Scan(&s.Id, &s.Name, &s.Rfid, &s.Cid, &c.Name)*/
	err := db.QueryRow("SELECT id, name, rfid, cid FROm student WHERE rfid=? LIMIT 1;", rfid).Scan(&s.Id, &s.Name, &s.Rfid, &s.Cid)

	if err != nil {
		log.Println("Error student.Fetch:", err)
		return Stu{}, err
	}

	return s, nil
}

// Fetchs returns the attendees for the given classitem.
// Not (yet) attending students are also given.
func Fetchs(ci classitem.ClassItem, db *sql.DB) ([]Att, error) {
	var atts []Att

	// Lamers tought us this!11

	// Fetch the attendee_item's and the students owning them.
	rows, err := db.Query("SELECT attendee_item.id, attendee_item.mins_early, student.id, student.name, student.rfid FROM student, attendee_item WHERE attendee_item.ciid=? AND attendee_item.sid = student.id", ci.Id)

	// It should only error during development, so I dare to 'handle' this error this way.
	if err != nil {
		log.Fatal(err)
		return atts, err
	}

	// Workaround, change it sometime.
	sids := []string{"0"}

	// Loop over the fetched items, and put them into atts.
	for rows.Next() {
		var a Att
		var s Stu
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

	// Fetch the students that did/are not attent this.
	rows, err = db.Query("SELECT student.name, student.rfid FROM student, class_item WHERE class_item.id = ? AND class_item.cid = student.cid AND student.id NOT IN ("+strings.Join(sids, ",")+") LIMIT 1337;", ci.Id)
	if err != nil {
		log.Fatal(err)
		return atts, err
	}

	// Loop over the fetched students, and also put them into atts.
	for rows.Next() {
		var a Att
		var s Stu
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

// Attent makes the given student attent the given classitem.
// MinutesEarly is positive when the class is about to start, negative if the student is not on time.
// The ID of the new attendee_item is returned.
func Attent(s Stu, c classitem.ClassItem, minsEarly int, db *sql.DB) int64 {
	r, err := db.Exec("INSERT INTO attendee_item (ciid, sid, mins_early) VALUES (?, ?, ?);", c.Id, s.Id, minsEarly)
	if err != nil {
		log.Fatal(err)
		return 0
	}

	lastid, _ := r.LastInsertId()
	return lastid
}
