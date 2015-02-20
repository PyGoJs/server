package stu

import "database/sql"

type Stu struct {
	Id   int    `json:"-"`
	Name string `json:"name"`
	Rfid string `json:"rfid"`
}

func Fetch(rfid string, db *sql.DB) (Stu, error) {
	var s Stu
	err := db.QueryRow("SELECT id, name, rfid FROM student WHERE rfid=? LIMIT 1;", rfid).Scan(&s.Id, &s.Name, &s.Rfid)
	if err != nil {
		return Stu{}, err
	}
	return s, nil
}
