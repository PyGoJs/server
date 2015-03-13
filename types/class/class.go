package class

import (
	"database/sql"
	"log"
	"time"
)

type Class struct {
	Id               int
	Name             string
	Icsid            int
	SchedLastFetched time.Time
}

func Fetch(cid int) Class {
	// This will be replaced with an actual query to the name and icsid in the future.
	return Class{
		Id:               cid,
		SchedLastFetched: time.Now(),
	}
}

// MaxStudents
// TO DO: Illness
func MaxStudents(c Class, db *sql.DB) (count int, err error) {
	err = db.QueryRow("SELECT count(*) FROM student WHERE cid=?;", c.Id).Scan(&count)

	if err != nil {
		log.Println("ERROR in attendee.MaxStudents;", err)
	}

	return
}
