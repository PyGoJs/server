// ClassItem only exists when there is at least one student attending the SchedItem.
package classitem

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/pygojs/server/types/class"
	"github.com/pygojs/server/types/schedule"
	"github.com/pygojs/server/util"
)

// ClassItem only exists when there is at least one student attending the SchedItem.
// MaxStudents is saved because illness and because the amount of students in a class can change.
type ClassItem struct {
	Id          int
	MaxStudents int `sql:"max_students"`
	Sched       schedule.SchedItem
}

// Next returns the next or current classitem for the given Class.
func Next(c class.Class, tm time.Time, db *sql.DB) (ClassItem, error) {
	// (Get the UnixTime from the start of this day and subtract is from the given tm.Unix,
	//  so we end up with the amount of seconds since the start of the day.)
	utsDay := time.Date(tm.Year(), tm.Month(), tm.Day(), 0, 0, 0, 0, util.Loc).Unix()
	end := tm.Unix() - utsDay
	day := int(tm.Weekday())

	var si schedule.SchedItem
	var ci ClassItem

	// It is not certain whether a classItem for the schedItem exists or not.
	var ciId sql.NullInt64 // ClassItem.Id
	var ciMs sql.NullInt64 // ClassItem.MaxStudents

	// Get the sched with the end-time that is the closest to tm time (but is still in the future).
	err := db.QueryRow(`
SELECT s.id, s.day, s.start, s.description, s.staff, class_item.id, class_item.max_students
FROM schedule_item AS s
LEFT JOIN class_item
ON s.id = class_item.siid 
WHERE s.end>=? 
 AND s.day=?
 AND s.usestopped=0 
 AND s.cid=?
 ORDER BY s.start LIMIT 1;
	`, end, day, c.Id).Scan(&si.Id, &si.Day, &si.StartInt, &si.Desc, &si.Staff, &ciId, &ciMs)

	if err != nil {
		if err != sql.ErrNoRows {
			log.Println("Error classitem.Fetch: ", err)
		}
		return ClassItem{}, err
	}

	if ciId.Valid {
		ci.Id = int(ciId.Int64)
		ci.MaxStudents = int(ciMs.Int64)
	}

	si.Start = time.Unix(utsDay+int64(si.StartInt), 0)

	ci.Sched = si
	return ci, nil
}

// Create makes a new class_item for the given SchedItem in the database, and returns the ClassItem.
func Create(si schedule.SchedItem, c class.Class, db *sql.DB) (ClassItem, error) {
	var ci ClassItem
	ci.Sched = si

	maxStu, _ := class.MaxStudents(c, db)
	ci.MaxStudents = maxStu

	r, err := db.Exec("INSERT INTO class_item (siid, cid, max_students) VALUES (?, ?, ?);", si.Id, c.Id, maxStu)

	if err != nil {
		log.Println("ERROR, cannot insert new class_item in classitem.Fetch, err:", err)
		return ci, err
	}

	id, _ := r.LastInsertId()
	ci.Id = int(id)

	fmt.Println(" Created classitem succesfully", ci.Id, ci.MaxStudents)

	return ci, nil
}
