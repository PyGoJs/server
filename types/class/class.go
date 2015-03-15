package class

import (
	"log"
	"time"

	"github.com/pygojs/server/util"
)

type Class struct {
	Id                  int
	Name                string
	IcsId               int
	SchedLastFetched    time.Time
	SchedLastFetchedInt int64
}

// Fetch returns the Class for the given class Id.
func Fetch(cid int) (Class, error) {
	var c Class

	err := util.Db.QueryRow("SELECT id, name, icsid, schedlastfetched FROM class WHERE id=? LIMIT 1;",
		cid).Scan(&c.Id, &c.Name, &c.IcsId, &c.SchedLastFetchedInt)

	if err != nil {
		log.Println("ERROR; cannot fetch class, err:", err, cid)
		return c, err
	}

	c.SchedLastFetched = time.Unix(c.SchedLastFetchedInt, 0)

	return c, nil

	// This will be replaced with an actual query to the name and icsid in the future.
	/*return Class{
		Id:               cid,
		SchedLastFetched: time.Now(),
	}*/
}

// MaxStudents
// TO DO: Illness
func MaxStudents(c Class) (count int, err error) {
	err = util.Db.QueryRow("SELECT count(*) FROM student WHERE cid=?;", c.Id).Scan(&count)

	if err != nil {
		log.Println("ERROR in attendee.MaxStudents;", err)
	}

	return
}
