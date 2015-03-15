package schedule

import (
	"testing"
	"time"

	"github.com/pygojs/server/types/class"
)

/*func TestFetchAll(t *testing.T) {
	db, err := util.Db()
	if err != nil {
		t.Fail()
		return
	}

	c := class.Class{Id: 1, Icsid: 14327}

	sis, err := FetchAll(c, db)
	if err != nil {
		t.Fail()
	}

	t.Log(sis)
}*/

func TestUpdate(t *testing.T) {
	c := class.Class{Id: 1, Icsid: 14327}

	err = Update(c, time.Now())
	if err != nil {
		t.Fail()
	}
}
