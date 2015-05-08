package class

import (
	"log"
	"time"

	"github.com/pygojs/server/util"
)

type Class struct {
	Id                  int       `json:"id"`
	Name                string    `json:"name"`
	SchedLastFetched    time.Time `json:"-"`
	SchedLastFetchedInt int64     `json:"-"`
}

// Fetch returns the Class for the given class Id.
func Fetch(cid int) (Class, error) {
	var c Class

	err := util.Db.QueryRow("SELECT id, name, schedlastfetched FROM class WHERE id=? LIMIT 1;",
		cid).Scan(&c.Id, &c.Name, &c.SchedLastFetchedInt)

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

func FetchAll() ([]Class, error) {
	var cs []Class

	rows, err := util.Db.Query("SELECT id, name FROM class LIMIT 1337;")

	if err != nil {
		log.Println("ERROR: cannot fetchall class, err:", err)
		return cs, err
	}

	for rows.Next() {
		var c Class

		err = rows.Scan(&c.Id, &c.Name)
		if err != nil {
			log.Println("ERROR while formatting Class in FetchAll, err:", err)
			return cs, err
		}

		cs = append(cs, c)
	}

	return cs, nil
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
