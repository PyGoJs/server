package att

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/pygojs/server/types/classitem"
	"github.com/pygojs/server/util"
)

// Attendee
type Att struct {
	Id int `json:"-"`
	//Ciid int `json:"ciid,omitempty"`
	//Sid       int  `json:",omitempty"`
	Attent    bool `json:"attent"`
	MinsEarly int  `sql:"mins_early" json:"minsEarly,omitempty"`
	Stu       Stu  `json:"stu"`
}

// Student
type Stu struct {
	Id   int    `json:"-"`
	Name string `json:"name"`
	Rfid string `json:"rfid,omitempty"`
	Cid  int    `json:"-"`
}

// Fetch returns the student with the given rfid.
func FetchStu(rfid string) (Stu, error) {
	var s Stu

	/*err := db.QueryRow(`
	SELECT s.id, s.name, s.rfid, s.cid, c.name FROM student AS s, class AS c
	WHERE s.rfid=? AND s.cid = c.id LIMIT 1;
		`, rfid).Scan(&s.Id, &s.Name, &s.Rfid, &s.Cid, &c.Name)*/
	err := util.Db.QueryRow("SELECT id, name, rfid, cid FROm student WHERE rfid=? LIMIT 1;", rfid).Scan(&s.Id, &s.Name, &s.Rfid, &s.Cid)

	if err != nil {
		log.Println("Error student.Fetch:", err)
		return Stu{}, err
	}

	return s, nil
}

// FetchAll returns the attendees for the given classitem.
// Not (yet) attending students are also given.
func FetchAll(cid int, isCiid bool) ([]Att, error) {

	if cid == 0 {
		log.Println("ERROR cid 0 in attendee.FetchAll")
		return []Att{}, errors.New("cid may not be 0")
	}

	var atts []Att

	// Workaround, change it sometime (the '0').
	// Contains the StudentID's that already have an attendee item, and should not be fetched with the second query.
	sids := []string{"0"}
	sql := `SELECT student.name, student.rfid FROM student
WHERE student.cid = ` + strconv.Itoa(cid) + ` LIMIT 1337;`

	// ClassItem can (or could) sometimes be empty, when fetched by classitem.Fetch
	if isCiid {
		ciid := cid

		// Fetch the attendee_item's and the students owning them.
		rows, err := util.Db.Query("SELECT attendee_item.id, attendee_item.mins_early, student.id, student.name, student.rfid FROM student, attendee_item WHERE attendee_item.ciid=? AND attendee_item.sid = student.id;", cid)

		// It should only error during development, so I dare to 'handle' this error this way.
		if err != nil {
			log.Println("ERROR cannot fetch attendees in att.Fetchs, err:", err)
			return atts, err
		}

		// Loop over the fetched items, and put them into atts.
		for rows.Next() {
			var a Att
			var s Stu
			err = rows.Scan(&a.Id, &a.MinsEarly, &s.Id, &s.Name, &s.Rfid)
			if err != nil {
				log.Fatal(err)
				return []Att{}, err
			}
			a.Attent = true
			a.Stu = s
			atts = append(atts, a)
			sids = append(sids, strconv.Itoa(s.Id))
		}
		sql = `SELECT student.name, student.rfid FROM student, class_item 
WHERE class_item.id = ` + strconv.Itoa(ciid) + ` AND class_item.cid = student.cid AND class_item.yearweek >= student.createdyrwk
AND student.id NOT IN (` + strings.Join(sids, ",") + `) LIMIT 1337;`
		fmt.Println("Ciid", ciid)
	}

	// Fetch the students that did/are not attent this.
	rows, err := util.Db.Query(sql)
	if err != nil {
		log.Println("ERROR cannot fetch students in att.Fetchs, err:", err)
		return atts, err
	}

	// Loop over the fetched students, and also put them into atts.
	for rows.Next() {
		var a Att
		var s Stu
		err = rows.Scan(&s.Name, &s.Rfid)
		if err != nil {
			log.Fatal(err)
			return []Att{}, err
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
func (s Stu) Attent(ci classitem.ClassItem, minsEarly int) int64 {
	// Note: Uncomment things to make it work on a []ClassItem.
	// Change the argument 'ci ClassItem' to 'cis []ClassItem'.

	/*if len(cis) == 0 {
		return 0
	}*/

	var firstId int64
	//for _, ci := range cis {
	if ci.Id == 0 {
		log.Println("ClassItem Id is 0")
		return 0
	}

	r, err := util.Db.Exec("INSERT INTO attendee_item (ciid, sid, mins_early) VALUES (?, ?, ?);", ci.Id, s.Id, minsEarly)
	if err != nil {
		log.Fatal(err)
		//continue
		return 0
	}

	//if firstId == 0 {
	firstId, _ = r.LastInsertId()
	//}
	//}
	return firstId
}

// IsAttending returns the ID of the attendee_item of the given Stu, for the given classItem.
// Given ID is zero if the Stu is not attending the classItem.
func (s Stu) IsAttending(ci classitem.ClassItem) (int, error) {
	var id int

	err := util.Db.QueryRow("SELECT id FROM attendee_item WHERE sid=? AND ciid=? LIMIT 1;", s.Id, ci.Id).Scan(&id)

	if err != nil {
		if err != sql.ErrNoRows {
			log.Println("ERROR while checking is student is attendee class_item, err:",
				err, s.Id, ci.Id)
		}
		return 0, err
	}

	return id, nil
}
