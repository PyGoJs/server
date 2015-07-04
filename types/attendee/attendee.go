package att

import (
	"database/sql"
	"log"
	"strconv"
	"strings"

	"github.com/pygojs/server/types/lesson"
	"github.com/pygojs/server/util"
)

// Attendee
type Att struct {
	Id int `json:"-"`
	//Ciid int `json:"ciid,omitempty"`
	//Sid       int  `json:",omitempty"`
	Attent bool      `json:"attent"`
	Stu    Stu       `json:"stu"`
	Items  []AttItem `json:"items"`
}

// Attendee item, see Att
type AttItem struct {
	Type      int `json:"type"`
	MinsEarly int `json:"minsEarly"`
}

// Student
type Stu struct {
	Id   int    `json:"-"`
	Name string `json:"name"`
	Rfid string `json:"rfid,omitempty"`
	Cid  int    `json:"-"`
}

// FetchStu returns the student with the given rfid.
func FetchStu(rfid string) (Stu, error) {
	var s Stu

	/*err := db.QueryRow(`
	SELECT s.id, s.name, s.rfid, s.cid, c.name FROM student AS s, class AS c
	WHERE s.rfid=? AND s.cid = c.id LIMIT 1;
		`, rfid).Scan(&s.Id, &s.Name, &s.Rfid, &s.Cid, &c.Name)*/
	err := util.Db.QueryRow("SELECT id, name, rfid, cid FROM student WHERE rfid=? LIMIT 1;", rfid).Scan(&s.Id, &s.Name, &s.Rfid, &s.Cid)

	if err != nil {
		log.Println("Error student.Fetch:", err)
		return Stu{}, err
	}

	return s, nil
}

// FetchAll returns the attendees for the given class id or location id.
// Not (yet) attending students are also given.
// When class id is set (and location id is not) only the regular students
// are given, with no attendee information.
func FetchAll(cid int, lid int) ([]Att, error) {

	var atts []Att

	// Workaround, change it sometime (the '0').
	// Contains the StudentID's that already have an attendee item,
	// and should not be fetched with the second query.
	sids := []string{"0"}

	fetchedAtts := false

	// // Lesson can (or could) sometimes be empty, when fetched by lesson.Fetch
	if lid > 0 {
		fetchedAtts = true

		// Fetch the attendee's and the students owning them.
		rows, err := util.Db.Query(`
SELECT attendee.id, attendee.attent, student.id, student.name, student.rfid 
FROM student, attendee 
WHERE attendee.lid=? AND attendee.sid = student.id 
ORDER BY attendee.id IS NOT NULL, student.name;`, lid)

		// It should only error during development, so I dare to 'handle' errors this way.
		if err != nil {
			log.Println("ERROR cannot fetch attendees in att.FetchAll, err:", err)
			return atts, err
		}

		// Loop over the fetched items, and put them into atts.
		for rows.Next() {
			var a Att
			var s Stu
			var attent int

			err = rows.Scan(&a.Id, &attent, &s.Id, &s.Name, &s.Rfid)
			if err != nil {
				log.Println("ERROR formatting attendee:", err)
				return []Att{}, err
			}

			// (attent is a tinyint in database, but should be a boolean)
			if attent > 0 {
				a.Attent = true
			}

			a.Stu = s

			rItems, err := util.Db.Query("SELECT type, mins_early FROM attendee_item WHERE aid=? ORDER BY mins_early DESC;", a.Id)

			if err != nil {
				log.Println("ERROR fetching attendee_items in FetchAll:", err, cid, lid)
				return []Att{}, err
			}

			for rItems.Next() {
				var ai AttItem
				err = rItems.Scan(&ai.Type, &ai.MinsEarly)
				if err != nil {
					log.Println("ERROR formatting attendee_item:", err)
					return []Att{}, err
				}

				a.Items = append(a.Items, ai)
			}

			atts = append(atts, a)
			sids = append(sids, strconv.Itoa(s.Id))
		}

	}

	// SQL string for fetching students.
	// Gets overwritten by another SQL string if there are attending students.
	sql := `SELECT student.name, student.rfid FROM student
WHERE student.cid = ` + strconv.Itoa(cid) + ` ORDER BY student.name LIMIT 1337;`

	// Change the SQL code for fetching the students if there are atts, so only students not attending are fetched.
	if fetchedAtts {
		sql = `
SELECT student.name, student.rfid FROM student, lesson 
WHERE lesson.id = ` + strconv.Itoa(lid) + ` AND lesson.cid = student.cid 
	AND lesson.yearweek >= student.createdyrwk
	AND student.id NOT IN (` + strings.Join(sids, ",") + `) ORDER BY student.name LIMIT 1337;`

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
			log.Println("ERROR formatting students:", err)
			return []Att{}, err
		}
		a.Attent = false
		a.Stu = s
		atts = append(atts, a)
	}

	return atts, nil
}

// Attent makes the given student attent the given lesson.
// MinutesEarly is positive when the class is about to start, negative if the student is not on time.
// The ID of the new attendee is returned.
func (s Stu) Attent(l lesson.Lesson, minsEarly int) int64 {
	// I think there might be a way to fetch lesson without the ID?
	if l.Id == 0 {
		log.Println("Lesson Id is 0")
		return 0
	}

	// Create the attentee row.
	r, err := util.Db.Exec("INSERT INTO attendee (lid, sid, attent) VALUES (?, ?, 1);", l.Id, s.Id)
	if err != nil {
		log.Println("ERROR creating attendee in Attent:", err, s, l)
		return 0
	}

	aid, _ := r.LastInsertId()

	// Create the attendee_item row.
	_, err = util.Db.Exec("INSERT INTO attendee_item (aid, type, mins_early) VALUES (?, 0, ?);", aid, minsEarly)
	if err != nil {
		log.Println("ERROR creating attendee item in Attent:", err, s, l)
		return 0
	}

	return aid
}

// IsAttending returns the ID of the given Stu's attendee, for the given lesson.
// Returned ID is zero if the Stu is not attending the lesson.
func (s Stu) IsAttending(l lesson.Lesson) (int, error) {
	var id int

	err := util.Db.QueryRow("SELECT id FROM attendee WHERE sid=? AND lid=? LIMIT 1;", s.Id, l.Id).Scan(&id)

	if err != nil && err != sql.ErrNoRows {
		log.Println("ERROR while checking is student is attendee lesson, err:", err, s.Id, l.Id)
		return 0, err
	}

	return id, nil
}
