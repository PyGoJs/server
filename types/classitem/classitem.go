package classitem

import (
	"database/sql"
	"log"
	"time"
)

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

type ClassItem struct {
	Id          int
	MaxStudents int `sql:"max_students"`
	Sched       SchedItem
}

/*
To do:
Fetch schedule item
Create class_item if it does not exists
Return class_item with schedItem embedded

Maybe merge scheditem.go and classitem.go ?

*/

func Fetch(cid int, tm time.Time, db *sql.DB) (ClassItem, error) {
	end := tm.Unix()

	var s SchedItem
	var c ClassItem

	var ciId sql.NullInt64
	var ciMs sql.NullInt64 // Class_Item.Max_Students

	// SELECT s.id, s.day, s.start, s.end, s.description, s.staff FROM schedule_item AS s WHERE s.start>=? AND s.end<=? AND s.usestopped=0 AND s.cid=?;
	// "SELECT s.id, s.day, s.start, s.end, s.description, s.staff, c.id, c.max_students FROM schedule_item AS s, class_item as c WHERE s.id = c.siid AND s.start<=? AND s.end>=? AND s.usestopped=0 AND s.cid=?;"

	err := db.QueryRow(`
SELECT s.id, s.start, s.description, s.staff, class_item.id, class_item.max_students
FROM schedule_item AS s
LEFT JOIN class_item
ON s.id = class_item.siid 
 AND s.end>=? 
 AND s.usestopped=0 
 AND s.cid=?
 ORDER BY s.start LIMIT 1;
	`, end, cid).Scan(&s.Id, &s.StartInt, &s.Desc, &s.Staff, &ciId, &ciMs)

	if err != nil {
		log.Println("Error classitem.Fetch: ", err)
		return ClassItem{}, err
	}

	if ciId.Valid {
		c.Id = int(ciId.Int64)
		c.MaxStudents = int(ciMs.Int64)
	}

	// To do: If no ClassItem from query above, insert it.

	s.Start = time.Unix(int64(s.StartInt), 0)

	c.Sched = s
	return c, nil
}
