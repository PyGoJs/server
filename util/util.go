package util

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	_ "github.com/ziutek/mymysql/godrv"
)

type Config struct {
	Http struct {
		Addr string `json:"addr"`
	} `json:"http"`
	Db struct {
		User string `json:"user"`
		Pass string `json:"password"`
		Db   string `json:"database"`
		//Addr string `json:"address"`
	} `json:"database"`
	Debug struct {
		Enabled bool      `json:"enabled"`
		TimeStr string    `json:"time"`
		Tm      time.Time `json:"-"`
	} `json:"debug"`
}

var Loc, _ = time.LoadLocation("Europe/Amsterdam")
var Db *sql.DB
var cfg Config

func CreateDb() (*sql.DB, error) {
	var err error
	Db, err = sql.Open("mymysql", fmt.Sprintf("%s/%s/%s", cfg.Db.Db, cfg.Db.User, cfg.Db.Pass))

	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return Db, nil
}

// LoadConfig parses the configuration file with the given filename
// and creates one with empty values it if it doesn't exist.
func LoadConfig(filename string) error {
	data, err := ioutil.ReadFile(filename)

	// File with given filename doesn't exist. Create it.
	if err != nil {

		// Defaults
		cfg.Debug.TimeStr = "2006-01-02 15:04"

		data, err := json.MarshalIndent(&cfg, "", "\t")
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(filename, data, 0644)
		if err != nil {
			log.Println("ERROR while writing to config, err:", err, filename)
			return err
		}

		return nil
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Println("ERROR unmarshalling JSON in config, err:", err, filename)
		return err
	}

	cfg.Debug.Tm, err = time.ParseInLocation("2006-01-02 15:04", cfg.Debug.TimeStr, Loc)
	if err != nil {
		log.Println("ERROR in config file: Invalid Time, layout: 2006-01-02 15:04")
		return err
	}

	return nil
}

// Cfg returns a copy of the running configuration.
func Cfg() Config {
	return cfg
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

// Ip, returns (correct) Ip address for the given http request.
func Ip(r http.Request) string {
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	// Proxy stuff
	if ip == "127.0.0.1" {
		ip = r.Header.Get("X-FORWARDED-FOR")
	}
	return ip
}

// Log Small
func LogS(format string, args ...interface{}) {
	fmt.Printf(time.Now().Format("0102-15:04")+" "+format+"\n", args...)
}
