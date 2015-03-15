package schedule

import (
	"log"
	"time"

	"github.com/pygojs/server/util"

	"github.com/pygojs/server/types/class"
)

// Sched is a 'wrapper' containing a slice of SchedItem.
type Sched struct {
	Items []SchedItem
}

// SchedItem contains the facts about a schedule item.
type SchedItem struct {
	Id       int
	Day      int
	StartInt int `sql:"start"`
	Start    time.Time
	EndInt   int `sql:"end"`
	Created  int
	//UseStopped int
	Desc  string `sql:"description"`
	Fac   string `sql:"facillity"`
	Staff string
}

// FetchAll returns all the SchedItems for a given class.
// Used by updatesched.
func FetchAll(class class.Class) (Sched, error) {
	var sis Sched

	rows, err := util.Db.Query("SELECT id, day, start, end, description, facility, staff FROM schedule_item WHERE cid=? AND usestopped=0 LIMIT 50;", class.Id)
	if err != nil {
		log.Println("ERROR, cannot fetch schedule in schedule.FetchAll, err:", err)
		return sis, err
	}

	for rows.Next() {
		var si SchedItem

		err = rows.Scan(&si.Id, &si.Day, &si.StartInt, &si.EndInt, &si.Desc, &si.Fac, &si.Staff)
		if err != nil {
			log.Println("ERROR while formatting SchedItems in schedule.FetchAll, err:", err)
			return sis, err
		}

		sis.Items = append(sis.Items, si)
	}

	return sis, nil
}
