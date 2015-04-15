// ClassItem only exists when there is at least one student attending the SchedItem.
package classitem

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/pygojs/server/types/class"
	"github.com/pygojs/server/types/client"
	"github.com/pygojs/server/types/schedule"
	"github.com/pygojs/server/util"
)

// ClassItem only exists when there is at least one student attending the SchedItem.
// MaxStudents is saved because illness and because the amount of students in a class can change.
type ClassItem struct {
	Id       int                `json:"id"`
	MaxStus  int                `sql:"max_students" json:"maxstus"`
	AmntStus int                `json:"amntstus",omitempty` // OmitEmpty for /nextclass
	YrWk     int                `json:"-"`
	Sched    schedule.SchedItem `json:"sched"`
}

// NextC fetches and returns the next or current ClassItem for the given Class.
func NextC(c class.Class, tm time.Time) (ClassItem, error) {
	cis, err := next("AND s.cid="+strconv.Itoa(c.Id), 1, tm)
	if err != nil || len(cis) == 0 {
		return ClassItem{}, err
	}
	return cis[0], err

}

// NextCl fetches and returns the next or current ClassItem for the given Client (facility).
func NextCl(cl client.Client, tm time.Time) ([]ClassItem, error) {
	return next("AND s.facility = '"+cl.Fac+"'", 15, tm)
}

// next is used by NextC and NextCl for fetching the actual ClassItem.
func next(sqlend string, limit int, tm time.Time) ([]ClassItem, error) {
	// (Get the UnixTime from the start of this day and subtract is from the given tm.Unix,
	//  so we end up with the amount of seconds since the start of the day.)
	utsDay := time.Date(tm.Year(), tm.Month(), tm.Day(), 0, 0, 0, 0, util.Loc).Unix()
	end := tm.Unix() - utsDay
	day := int(tm.Weekday())
	yr, wk := tm.ISOWeek()
	yrWk, _ := strconv.Atoi(strconv.Itoa(yr) + strconv.Itoa(wk))

	// Get the sched with the end-time that is the closest to tm time (but is still in the future).
	rows, err := util.Db.Query(`
SELECT s.id, s.cid, s.day, s.start, s.end, s.description, s.facility, s.staff, class_item.id, class_item.max_students
FROM schedule_item AS s
LEFT JOIN class_item
ON s.id = class_item.siid AND class_item.yearweek=?
WHERE s.end>=? 
 AND s.day=?
 AND s.usestopped=0 
 `+sqlend+`
 GROUP BY s.cid
 ORDER BY s.start LIMIT `+strconv.Itoa(limit)+`;
	`, yrWk, end, day)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Println("Error classitem.Fetch: ", err)
		}
		return []ClassItem{}, err
	}

	var cis []ClassItem

	for rows.Next() {
		var ci ClassItem
		var si schedule.SchedItem

		// It is not certain whether a classItem for the schedItem exists or not.
		var ciId sql.NullInt64 // ClassItem.Id
		var ciMs sql.NullInt64 // ClassItem.MaxStudents

		err = rows.Scan(&si.Id, &si.Cid, &si.Day, &si.StartInt, &si.EndInt,
			&si.Desc, &si.Fac, &si.Staff, &ciId, &ciMs)
		if err != nil {
			log.Println("ERROR while formatting ClassItem in next, err:", err)
			continue
		}

		if ciId.Valid {
			ci.Id = int(ciId.Int64)
			ci.MaxStus = int(ciMs.Int64)
			ci.YrWk = yrWk
		}

		si.Start = time.Unix(utsDay+int64(si.StartInt), 0)

		ci.Sched = si
		cis = append(cis, ci)
	}
	return cis, nil
}

// Afters gives a []ClassItem containing the classes after the given ci.
// It only contains classes that are directly after the ci, and in the same facility.
func (ci ClassItem) Afters(c class.Class) ([]ClassItem, error) {
	cis := append([]ClassItem{}, ci)

	rows, err := util.Db.Query(`
SELECT s.id, s.day, s.start, s.end, s.description, s.facility, s.staff, class_item.id, class_item.max_students
FROM schedule_item AS s
LEFT JOIN class_item
ON s.id = class_item.siid AND class_item.yearweek=?
WHERE s.start>=? 
 AND s.day=?
 AND s.usestopped=0 
 AND s.cid=?
 AND s.facility=?
 ORDER BY s.start LIMIT 10;
	`, ci.YrWk, ci.Sched.EndInt, ci.Sched.Day, c.Id, ci.Sched.Fac)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Println("ERROR fetching Afters in classitem, err:", err)
		}
		return cis, err
	}

	// lastEnd is used for checking if there is a break between scheduleitems.
	lastEnd := ci.Sched.EndInt

	for rows.Next() {
		var tci ClassItem // TempClassItem
		var si schedule.SchedItem

		// It is not certain whether a classItem for the schedItem exists or not.
		var ciId sql.NullInt64 // ClassItem.Id
		var ciMs sql.NullInt64 // ClassItem.MaxStudents

		err = rows.Scan(&si.Id, &si.Day, &si.StartInt, &si.EndInt, &si.Desc, &si.Fac, &si.Staff,
			&ciId, &ciMs)
		if err != nil {
			log.Println("ERROR while formatting ClassItems in classitem.Afters, err:", err)
			continue
		}

		// Check if there is a break before this item.
		if si.StartInt > lastEnd {
			break
		}

		// Ignore duplicates. Easy fix for dirty devving database.
		if si.EndInt == lastEnd {
			fmt.Println("Duplicate nub")
			continue
		}

		lastEnd = si.EndInt

		if ciId.Valid {
			tci.Id = int(ciId.Int64)
			tci.MaxStus = int(ciMs.Int64)
			tci.YrWk = ci.YrWk
		}

		tci.Sched = si
		cis = append(cis, tci)

	}
	return cis, nil
}

func FetchAll(cid, yrwk int) ([]ClassItem, error) {
	var cis []ClassItem

	rows, err := util.Db.Query(`
SELECT s.day, s.start, s.end, s.description, s.facility, s.staff, c.id, c.max_students
FROM schedule_item AS s, class_item AS c
WHERE c.cid = ? AND c.yearweek = ? AND c.siid = s.id
	`, cid, yrwk)
	if err != nil {
		log.Println("ERROR while FetchAll classItem, err:", err)
		return cis, err
	}

	for rows.Next() {
		var si schedule.SchedItem
		var ci ClassItem

		err = rows.Scan(&si.Day, &si.StartInt, &si.EndInt, &si.Desc, &si.Fac, &si.Staff,
			&ci.Id, &ci.MaxStus)
		if err != nil {
			log.Println("ERROR while formatting ClassItem in FetchAll, err:", err)
			return cis, err
		}

		err = util.Db.QueryRow("SELECT COUNT(id) FROM attendee_item WHERE ciid=? LIMIT 50;",
			ci.Id).Scan(&ci.AmntStus)
		if err != nil {
			log.Println("ERROR/warning while fetching attendee count for classitem, err:", err)
		}

		ci.Sched = si

		cis = append(cis, ci)
	}

	return cis, nil
}

// Create makes a new class_item for the given SchedItem in the database, and returns the ClassItem.
func (ci ClassItem) Create(c class.Class, tm time.Time) (ClassItem, error) {
	yr, wk := tm.ISOWeek()

	maxStu, _ := class.MaxStudents(c)
	ci.MaxStus = maxStu
	fmt.Println("SchedId", ci.Sched.Id)

	r, err := util.Db.Exec("INSERT INTO class_item (siid, cid, max_students, yearweek) VALUES (?, ?, ?, ?);",
		ci.Sched.Id, c.Id, maxStu, strconv.Itoa(yr)+strconv.Itoa(wk))

	if err != nil {
		log.Println("ERROR, cannot insert new class_item in classitem.Fetch, err:", err)
		return ci, err
	}

	id, _ := r.LastInsertId()
	ci.Id = int(id)

	return ci, nil
}
