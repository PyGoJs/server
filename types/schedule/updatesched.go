package schedule

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/pygojs/server/types/class"
	"github.com/pygojs/server/util"
)

// Time layouts for the fetched rawSchedule
const rsDateLayout = "Mon Jan 2 2006"
const rsDateTimeLayout = "Mon Jan 2 2006 15:04"

type schedItemUpd struct {
	SchedItem
	matching bool
}

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
func UpdateSched(c class.Class, tm time.Time, db *sql.DB) error {
	yr, wk := tm.ISOWeek()
	rs, err := fetchSched(c.Icsid, yr, wk+1)
	if err != nil {
		return err
	}

	// rawSched to []SchedItem
	siNew, err := rs.format()
	if err != nil {
		return err
	}

	siOld, err := FetchAll(c, db)
	if err != nil {
		return err
	}

	_, _ = siNew, siOld

	// To do
	// Query old items and compare
	// Think about renaming classitem.Fetch to something along the lines of 'classitem.Next'.

	return nil
}

// fetchSched returns the live schedule, fetched from the API, as rawSched.
func fetchSched(icsid, yr, wk int) (rawSched, error) {
	res, err := http.Get(fmt.Sprintf("http://xedule.novaember.com/weekschedule.%d.json?year=%d&week=%d", icsid, yr, wk))
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

// format takes a rawSched and returns it as schedItemUpd.
func (rs rawSched) format() ([]schedItemUpd, error) {
	var siu []schedItemUpd
	var tmDay, tmStart, tmEnd time.Time
	for _, d := range rs.days {
		var err1, err2, err3 error
		tmDay, err1 = time.Parse(rsDateLayout, d.Date)

		for _, e := range d.Events {
			tmStart, err2 = time.Parse(rsDateTimeLayout, d.Date+" "+e.Start)
			tmStart = tmStart.In(util.Loc)
			tmEnd, err3 = time.Parse(rsDateTimeLayout, d.Date+" "+e.End)
			tmEnd = tmEnd.In(util.Loc)

			siu = append(siu, schedItemUpd{SchedItem{
				StartInt: int(tmStart.Unix() - tmDay.Unix()),
				EndInt:   int(tmEnd.Unix() - tmDay.Unix()),
				Day:      int(tmDay.Weekday()),
				Staff:    strings.Join(e.Staffs, ","),
				Fac:      strings.Join(e.Facilities, ","),
				Desc:     e.Description,
			}, false})
		}

		if none, str := util.CheckErrs([]error{err1, err2, err3}); none == false {
			fmt.Println("ERROR, Cannot format rawSched to []SchedItem: ", str)
			return []schedItemUpd{}, errors.New("cannot format rawSched to []SchedItem; " + str)
		}
	}

	return siu, nil
}

// compareAndSave does not work as of yet.
func (siNew Sched) compareAndSave(siOld Sched) {
	newToSave := siNew.Items
	// MatchingOldIds will contain the ids's of the old items that should not be removed.
	//var matchingOldIds []int
	change := true

	// Only loop if there are oldies
	if len(siOld.Items) > 0 {
		newToSave = []SchedItem{}
		change = false
		for _, n := range siNew.Items {
			match := false
		oldLoop:
			for _, o := range siOld.Items {
				if n.Day == o.Day && n.StartInt == o.StartInt && n.EndInt == o.EndInt &&
					n.Desc == o.Desc && n.Fac == o.Fac && n.Staff == o.Staff {
					match = true

					//matchingOldIds = append(matchingOldIds, o.Id)
					break oldLoop
				}
			}

			// No match means the schedule item is new and should be saved.
			if match == false {
				newToSave = append(newToSave, n)
				change = true
				fmt.Println(" New:", n)
			}
		}

		// Old items that have a match
		var remIds []int
		_, _ = remIds, change
	}
}
