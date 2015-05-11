package lesson

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/pygojs/server/types/class"
	"github.com/pygojs/server/types/client"
	"github.com/pygojs/server/types/schedule"
	"github.com/pygojs/server/util"
)

// Lesson only exists when there is at least one student attending the SchedItem.
// MaxStudents is saved because illness and because the amount of students in a class can change.
type Lesson struct {
	Id       int `json:"id"`
	MaxStus  int `sql:"max_students" json:"maxstus"`
	AmntStus int `json:"amntstus",omitempty` // OmitEmpty for /nextclass

	// YrWk: A lesson is attached to a schedule item, but a schedule item doens't describe the year/week.
	YrWk  int                `json:"-"`
	Sched schedule.SchedItem `json:"sched"`
}

// NextC fetches and returns the next or current Lesson for the given Class.
func NextC(c class.Class, tm time.Time) (Lesson, error) {
	ls, err := next("AND s.cid="+strconv.Itoa(c.Id), 1, tm)
	if err != nil || len(ls) == 0 {
		return Lesson{}, err
	}
	return ls[0], err

}

// NextCl fetches and returns the next or current Lesson for the given Client (facility).
func NextCl(cl client.Client, tm time.Time) ([]Lesson, error) {
	return next("AND s.facility = '"+cl.Fac+"'", 15, tm)
}

// next is used by NextC and NextCl for fetching the actual Lesson.
func next(sqlend string, limit int, tm time.Time) ([]Lesson, error) {
	// (Get the UnixTime from the start of this day and subtract it from the given tm.Unix,
	//  so we end up with the amount of seconds since the start of the day.)
	utsDay := time.Date(tm.Year(), tm.Month(), tm.Day(), 0, 0, 0, 0, util.Loc).Unix()
	end := tm.Unix() - utsDay
	day := int(tm.Weekday())
	yr, wk := tm.ISOWeek()
	yrWk, _ := strconv.Atoi(strconv.Itoa(yr) + strconv.Itoa(wk))

	// Get the schedule-items with the end-time that is the closest to tm time (but still in the future).
	rows, err := util.Db.Query(`
SELECT s.id, s.cid, s.day, s.start, s.end, s.description, s.facility, s.staff, lesson.id, lesson.max_students
FROM schedule_item AS s
LEFT JOIN lesson
ON s.id = lesson.siid AND lesson.yearweek=?
WHERE s.end>=? 
 AND s.day=?
 AND s.usestopped=0 
 `+sqlend+`
 GROUP BY s.cid
 ORDER BY s.start LIMIT `+strconv.Itoa(limit)+`;
	`, yrWk, end, day)
	if err != nil {
		log.Println("Error classitem.Fetch: ", err)
		return []Lesson{}, err
	}

	var ls []Lesson

	var count int
	for rows.Next() {
		var l Lesson
		var si schedule.SchedItem

		// It is not certain whether a lesson for the schedItem exists or not.
		var lId sql.NullInt64 // Lesson.Id
		var lMs sql.NullInt64 // Lesson.MaxStudents

		err = rows.Scan(&si.Id, &si.Cid, &si.Day, &si.StartInt, &si.EndInt,
			&si.Desc, &si.Fac, &si.Staff, &lId, &lMs)
		if err != nil {
			log.Println("ERROR while formatting Lesson in next, err:", err)
			continue
		}

		if lId.Valid {
			l.Id = int(lId.Int64)
			l.MaxStus = int(lMs.Int64)
			l.YrWk = yrWk
		}

		si.Start = time.Unix(utsDay+int64(si.StartInt), 0)

		l.Sched = si
		ls = append(ls, l)
		count++
	}

	if count == 0 {
		return []Lesson{}, errors.New("no more classes for today")
	}

	return ls, nil
}

// Afters gives a []Lesson containing the classes after the given l.
// It only contains lessons that are directly after the l, and in the same facility.
func (l Lesson) Afters(c class.Class) ([]Lesson, error) {
	ls := []Lesson{l}

	rows, err := util.Db.Query(`
SELECT s.id, s.day, s.start, s.end, s.description, s.facility, s.staff, lesson.id, lesson.max_students
FROM schedule_item AS s
LEFT JOIN lesson
ON s.id = lesson.siid AND lesson.yearweek=?
WHERE s.start>=? 
 AND s.day=?
 AND s.usestopped=0 
 AND s.cid=?
 AND s.facility=?
 ORDER BY s.start LIMIT 10;
	`, l.YrWk, l.Sched.EndInt, l.Sched.Day, c.Id, l.Sched.Fac)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Println("ERROR fetching Afters in classitem, err:", err)
		}
		return ls, err
	}

	// lastEnd is used for checking if there is a break between scheduleitems.
	lastEnd := l.Sched.EndInt

	for rows.Next() {
		var tl Lesson // TempLesson
		var si schedule.SchedItem

		// It is not certain whether a lesson for the schedItem exists or not.
		var lId sql.NullInt64 // Lesson.Id
		var lMs sql.NullInt64 // Lesson.MaxStudents

		err = rows.Scan(&si.Id, &si.Day, &si.StartInt, &si.EndInt, &si.Desc, &si.Fac, &si.Staff,
			&lId, &lMs)
		if err != nil {
			log.Println("ERROR while formatting Lessons in classitem.Afters, err:", err)
			continue
		}

		// Check if there is a break before this item.
		if si.StartInt > lastEnd {
			// Stop the loop because we reached a break, no more lessons directly after.
			break
		}

		// Ignore duplicates. Easy fix for dirty devving database.
		if si.EndInt == lastEnd {
			fmt.Println("Duplicate nub")
			continue
		}

		lastEnd = si.EndInt

		if lId.Valid {
			tl.Id = int(lId.Int64)
			tl.MaxStus = int(lMs.Int64)
			tl.YrWk = l.YrWk
		}

		tl.Sched = si
		ls = append(ls, tl)

	}
	return ls, nil
}

// FetchAll gives the lessons in the yearweek for the given class.
func FetchAll(cid, yrwk int) ([]Lesson, error) {
	var ls []Lesson

	rows, err := util.Db.Query(`
SELECT s.day, s.start, s.end, s.description, s.facility, s.staff, c.id, c.max_students
FROM schedule_item AS s, lesson AS c
WHERE c.cid = ? AND c.yearweek = ? AND c.siid = s.id
	`, cid, yrwk)
	if err != nil {
		log.Println("ERROR while FetchAll classItem, err:", err)
		return ls, err
	}

	for rows.Next() {
		var si schedule.SchedItem
		var l Lesson

		err = rows.Scan(&si.Day, &si.StartInt, &si.EndInt, &si.Desc, &si.Fac, &si.Staff,
			&l.Id, &l.MaxStus)
		if err != nil {
			log.Println("ERROR while formatting Lesson in FetchAll, err:", err)
			return ls, err
		}

		// Query amount of attending students.
		err = util.Db.QueryRow("SELECT COUNT(id) FROM attendee WHERE lid=? LIMIT 50;",
			l.Id).Scan(&l.AmntStus)
		if err != nil {
			log.Println("ERROR/warning while fetching attendee count for classitem, err:", err)
		}

		l.Sched = si

		ls = append(ls, l)
	}

	return ls, nil
}

// Create makes a new lesson for the given SchedItem in the database, and returns the Lesson.
func (l *Lesson) Create(c class.Class, tm time.Time) error {
	yr, wk := tm.ISOWeek()

	maxStu, _ := class.MaxStudents(c)
	l.MaxStus = maxStu

	r, err := util.Db.Exec("INSERT INTO lesson (siid, cid, max_students, yearweek) VALUES (?, ?, ?, ?);",
		l.Sched.Id, c.Id, maxStu, strconv.Itoa(yr)+strconv.Itoa(wk))

	if err != nil {
		log.Println("ERROR, cannot insert new lesson in classitem.Fetch, err:", err)
		return err
	}

	id, _ := r.LastInsertId()
	l.Id = int(id)

	return nil
}
