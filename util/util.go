package util

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/ziutek/mymysql/godrv"
)

var Loc, _ = time.LoadLocation("Europe/Amsterdam")

func Db() (*sql.DB, error) {
	// Database password in version control, yey
	// Temporary though. A config struct should be made soon.
	db, err := sql.Open("mymysql", "pygojs/pygojs/NVH5KfDhD7GXf88p")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return db, nil
}

func CheckErrs(errs []error) (bool, string) {
	var str string
	for i, e := range errs {
		if e != nil {
			str += fmt.Sprintf("%d: %s. \n", i, e)
		}
	}
	return len(str) == 0, str
}
