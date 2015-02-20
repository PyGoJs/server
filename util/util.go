package util

import (
	"database/sql"
	"log"

	_ "github.com/ziutek/mymysql/godrv"
)

func Db() (*sql.DB, error) {
	db, err := sql.Open("mymysql", "pygojs/pygojs/NVH5KfDhD7GXf88p")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return db, nil
}
