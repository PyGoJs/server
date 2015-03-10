package schedule

import (
	"testing"
	"time"

	"github.com/pygojs/server/types/class"

	"github.com/pygojs/server/util"
)

func TestFetchAll(t *testing.T) {
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
}

func TestUpdateSched(t *testing.T) {
	db, err := util.Db()
	if err != nil {
		t.Fail()
		return
	}

	c := class.Class{Id: 1, Icsid: 14327}

	err = UpdateSched(c, time.Now(), db)
	if err != nil {
		t.Fail()
	}
}
