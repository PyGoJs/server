package schedule

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pygojs/server/types/class"
	"github.com/pygojs/server/util"
)

const save = true

// rawSched is used for Unmarshalling the schedule JSON from the API.
type rawSched struct {
	Days []struct {
		Day    time.Weekday
		Events []struct {
			Start   int64
			End     int64
			Desc    string
			Classes []string
			Facs    []string
			Staffs  []string
		}
	}
}

const schedUrl = "http://xeduleapi.remi.im/schedule.json?aid=%d&year=%d&week=%d&nocache=true"

func Update(c class.Class, tm time.Time) (bool, error) {
	yr, wk := tm.ISOWeek()

	resp, err := http.Get(fmt.Sprintf(schedUrl, c.Id, yr, wk))
	if err != nil {
		log.Println("ERROR fetching schedule:", err, c)
		return false, err
	}
	defer resp.Body.Close()

	cont, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("ERROR reading from fetched schedule:", err, c)
		return false, err
	}

	var rs rawSched
	json.Unmarshal(cont, &rs)

	// Format the rawSched into a regular Sched.
	var sNew Sched
	for _, d := range rs.Days {
		var utsDay int64
		for _, e := range d.Events {
			if utsDay == 0 {
				utsDay = time.Date(tm.Year(), tm.Month(), tm.Day(), 0, 0, 0, 0, util.Loc).Unix()
			}
			sNew.Items = append(sNew.Items, SchedItem{
				StartInt: int(e.Start - utsDay),
				EndInt:   int(e.End - utsDay),
				Start:    time.Unix(e.Start, 0),
				Day:      int(d.Day),
				Staff:    strings.Join(e.Staffs, ","),
				Fac:      strings.Join(e.Facs, ","),
				Desc:     e.Desc,
			})
		}
	}

	// Fetch the current/old schedule from the database.
	sOld, err := FetchAll(c)
	if err != nil {
		return false, err
	}

	// Compare and save changes.
	change, err := sNew.compareAndSave(sOld, c)
	if err != nil {
		return change, err
	}

	// Save current time as 'Schedule last fetched' time.
	_, err = util.Db.Exec("UPDATE class SET schedlastfetched=? WHERE id=? LIMIT 1;", time.Now().Unix(), c.Id)
	if err != nil {
		log.Println("ERROR updating schedlastfetched of class, err:", err, c.Id)
		return false, err
	}

	return change, nil
}

func UpdateAll(tm time.Time) error {
	cs, err := class.FetchAll()
	if err != nil {
		return err
	}

	for _, c := range cs {
		_, err = Update(c, tm)
		if err != nil {
			return err
		}
	}

	return nil
}

// compareAndSave saves the differences between the given siNew and siOld.
// New items are created, items that are no longer used are disabled.
// (Not the most efficient and clean code.)
func (siNew Sched) compareAndSave(siOld Sched, c class.Class) (bool, error) {
	newToSave := siNew.Items
	// MatchingOldIds will contain the ids's of the old items that should not be removed.
	change := true

	// Only loop if there are oldies
	if len(siOld.Items) > 0 {
		var matchingOldIds []int
		newToSave = []SchedItem{}
		change = false

		for _, n := range siNew.Items {
			match := false
		oldLoop:
			for _, o := range siOld.Items {
				if n.Day == o.Day && n.StartInt == o.StartInt && n.EndInt == o.EndInt &&
					n.Desc == o.Desc && n.Fac == o.Fac && n.Staff == o.Staff {
					match = true

					matchingOldIds = append(matchingOldIds, o.Id)
					break oldLoop
				}
			}

			// No match means the schedule item is new and should be saved.
			if match == false {
				newToSave = append(newToSave, n)
				change = true
			}
		}

		// Old items that should be removed (have no match).
		var remIds []string
		for _, o := range siOld.Items {
			rem := true

		oldIdsLoop:
			for _, oId := range matchingOldIds {
				if o.Id == oId {
					rem = false
					break oldIdsLoop
				}
			}
			if rem == true {
				remIds = append(remIds, strconv.Itoa(o.Id))
			}
		}

		// Remove the old items (well not really remove, but setting usestopped).
		if len(remIds) > 0 {
			change = true
			ut := time.Now().Unix()
			sql := fmt.Sprintf("UPDATE schedule_item SET usestopped=? WHERE id IN (%s);", strings.Join(remIds, ","))
			if save {
				_, err := util.Db.Exec(sql, ut)
				if err != nil {
					log.Println("ERROR while updating schedule; cannot set usestopped for ids ", remIds, ", err:", err)
					return false, err
				}
			}
		}
	}

	if len(newToSave) > 0 {
		var sqlstr string
		ut := time.Now().Unix()

		for _, n := range newToSave {
			if len(sqlstr) > 0 {
				sqlstr += ", "
			}

			// cid day start end created description facility staff
			sqlstr += fmt.Sprintf("(%d, %d, %d, %d, %d, '%s', '%s', '%s')",
				c.Id, n.Day, n.StartInt, n.EndInt, ut, n.Desc, n.Fac, n.Staff)
		}

		sql := fmt.Sprintf(
			"INSERT INTO schedule_item (cid, day, start, end, created, description, facility, staff) VALUES %s;",
			sqlstr)

		if save {
			_, err := util.Db.Exec(sql)
			if err != nil {
				log.Println("ERROR while updating schedule, cannot insert items, err:", err, c.Id)
				log.Println(sql)
				return false, err
			}
		}
	}

	return change, nil
}
