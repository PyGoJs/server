package schedule

import (
	"encoding/json"
	"errors"
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

// Time layouts for the fetched rawSchedule
const rsDateLayout = "Mon Jan 2 2006"
const rsDateTimeLayout = "Mon Jan 2 2006 15:04"

const save = true

// rawSched is a 'wrapper' containg an slice of rawSchedDay.
type rawSched struct {
	days []rawSchedDay
}

// rawSchedDay is used for Unmarshalling the schedule JSON from the API.
type rawSchedDay struct {
	Date   string
	Events []struct {
		Start       string
		End         string
		Classes     []string
		Staffs      []string
		Facilities  []string
		Description string
	}
}

// UpdateSched fetches the live schedule from the API,
// and saves the changes in the database.
func Update(c class.Class, tm time.Time) (bool, error) {
	yr, wk := tm.ISOWeek()
	// Next week if Saturday, or Friday from 5pm.
	/*if tm.Day() == 6 || (tm.Day() == 5 && tm.Hour() >= 17) {
		yr, wk = tm.Add(0, 0, 3).ISOWeek()
	}*/
	rs, err := fetchSched(c.IcsId, yr, wk)
	if err != nil {
		return false, err
	}

	// rawSched to []SchedItem
	sNew, err := rs.format()
	if err != nil {
		return false, err
	}

	sOld, err := FetchAll(c)
	if err != nil {
		return false, err
	}

	change, err := sNew.compareAndSave(sOld, c)
	if err != nil {
		return change, err
	}

	_, err = util.Db.Exec("UPDATE class SET schedlastfetched=? WHERE id=? LIMIT 1;", time.Now().Unix(), c.Id)
	if err != nil {
		log.Println("ERROR updating schedlastfetched of class, err:", err, c.Id)
		return false, err
	}

	return change, nil
}

// fetchSched returns the live schedule, fetched from the API, as rawSched.
func fetchSched(icsid, yr, wk int) (rawSched, error) {
	url := fmt.Sprintf("http://xedule.novaember.com/weekschedule.%d.json?year=%d&week=%d", icsid, yr, wk)
	// Make Novaember reload the schedule first. (Returns an error instead of valid JSON, so refetch it)
	_, err := http.Get(url + "&reload")
	if err != nil {
		log.Println("ERROR/Warning while fetching schedule, cannot tell Novaember to reload, err:", err)
	}
	res, err := http.Get(url)
	if err != nil {
		log.Println("ERROR while fetching schedule, err:", err)
		return rawSched{}, err
	}

	cont, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	if err != nil {
		log.Println("ERROR while parsing fetched schedule, err:", err)
		return rawSched{}, err
	}

	var rs []rawSchedDay
	err = json.Unmarshal(cont, &rs)

	if err != nil {
		log.Println("ERROR while unmarshaling fetched schedule, err:", err)
		return rawSched{}, err
	}

	return rawSched{days: rs}, nil
}

// format takes a rawSched and returns it as Sched.
func (rs rawSched) format() (Sched, error) {
	var s Sched
	var tmDay, tmStart, tmEnd time.Time
	for _, d := range rs.days {
		var err1, err2, err3 error
		tmDay, err1 = time.Parse(rsDateLayout, d.Date)

		for _, e := range d.Events {
			tmStart, err2 = time.Parse(rsDateTimeLayout, d.Date+" "+e.Start)
			tmStart = tmStart.In(util.Loc)
			tmEnd, err3 = time.Parse(rsDateTimeLayout, d.Date+" "+e.End)
			tmEnd = tmEnd.In(util.Loc)

			s.Items = append(s.Items, SchedItem{
				StartInt: int(tmStart.Unix() - tmDay.Unix()),
				EndInt:   int(tmEnd.Unix() - tmDay.Unix()),
				Day:      int(tmDay.Weekday()),
				Staff:    strings.Join(e.Staffs, ","),
				Fac:      strings.Join(e.Facilities, ","),
				Desc:     e.Description,
			})
		}

		if none, str := util.CheckErrs([]error{err1, err2, err3}); none == false {
			fmt.Println("ERROR, Cannot format rawSched to []SchedItem: ", str)
			return Sched{}, errors.New("cannot format rawSched to []SchedItem; " + str)
		}
	}

	return s, nil
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
