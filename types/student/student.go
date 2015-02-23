package stu

import (
	"database/sql"
	"log"
)

type Stu struct {
	Id   int    `json:"-"`
	Name string `json:"name"`
	Rfid string `json:"rfid"`
	Cid  int    `json:"-"`
}

func Fetch(rfid string, db *sql.DB) (Stu, error) {
	var s Stu
	err := db.QueryRow("SELECT id, name, rfid, cid FROM student WHERE rfid=? LIMIT 1;", rfid).Scan(&s.Id, &s.Name, &s.Rfid, &s.Cid)
	if err != nil {
		log.Println("Error student.Fetch:", err)
		return Stu{}, err
	}
	return s, nil
}
