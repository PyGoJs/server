package classitem

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/pygojs/server/types/class"
)

// SchedItem contains the facts about the class's schedule (what, where, with, when).
type SchedItem struct {
	Id       int
	StartInt int `sql:"start"`
	Start    time.Time
	//End        int
	//Created    int
	//UseStopped int
	Desc  string `sql:"description"`
	Staff string
}

// ClassItem only exists when there is at least one student attending the SchedItem.
// MaxStudents is saved because illness and because the amount of students in a class can change.
type ClassItem struct {
	Id          int
	MaxStudents int `sql:"max_students"`
	Sched       SchedItem
}

// Fetch returns the current or next classitem for the given ClassId.
func Fetch(c class.Class, tm time.Time, db *sql.DB) (ClassItem, error) {
	end := tm.Unix()

	var si SchedItem
	var ci ClassItem

	// It is not certain whether a classItem for the schedItem exists or not.
	var ciId sql.NullInt64 // ClassItem.Id
	var ciMs sql.NullInt64 // ClassItem.MaxStudents

	err := db.QueryRow(`
SELECT s.id, s.start, s.description, s.staff, class_item.id, class_item.max_students
FROM schedule_item AS s
LEFT JOIN class_item
ON s.id = class_item.siid 
 AND s.end>=? 
 AND s.usestopped=0 
 AND s.cid=?
 ORDER BY s.start LIMIT 1;
	`, end, c.Id).Scan(&si.Id, &si.StartInt, &si.Desc, &si.Staff, &ciId, &ciMs)

	if err != nil {
		log.Println("Error classitem.Fetch: ", err)
		return ClassItem{}, err
	}

	if ciId.Valid {
		ci.Id = int(ciId.Int64)
		ci.MaxStudents = int(ciMs.Int64)
	} else {
		// If no ClassItem from query above, insert it.
		maxStu, _ := class.MaxStudents(c, db)
		r, err := db.Exec("INSERT INTO class_item (siid, cid, max_students) VALUES (?, ?, ?);", si.Id, c.Id, maxStu)

		if err != nil {
			log.Println("ERROR, cannot insert new class_item in classitem.Fetch, err:", err)
		}

		id, _ := r.LastInsertId()
		ci.Id = int(id)
		ci.MaxStudents = maxStu

		fmt.Println(" Created classitem succesfully", ci.Id, ci.MaxStudents)
	}

	si.Start = time.Unix(int64(si.StartInt), 0)

	ci.Sched = si
	return ci, nil
}
