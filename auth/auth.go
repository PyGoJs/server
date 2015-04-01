// auth handles basic authentication for the api/website.
// I am aware that creating your own basic authentication system is considered bad practise.
// However, for a school project (what this is) I value me learning things greater than having a secure system.
// If this wasn't a school project, and it's main goal wasn't learning, I would have used an existing auth package.
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/pygojs/server/util"
)

type User struct {
	Id    int
	Login string
	Pass  string `json:"-"` // SHA256
}

// session holds an existing login.
type session struct {
	key        string // Session key
	u          *User
	loginTime  time.Time
	accessTime time.Time
}

// users, login by User
var users = map[string]*User{}
var usersM = &sync.Mutex{}

// sessions, key by session
var sessions = map[string]*session{}
var sessionsM = &sync.Mutex{}

// sesTimeout contains the maximum amount of minutes that a session remains valid
// when it hasn't been accessed for this amount of minutes.
const sesTimeout = 15

// Login creates a new session for a user if a user for the given login can be found
// and the password is correct. Returns the user and the new session key, or an error if
// the given credentials are incorrect.
func Login(login, pass string) (User, string, error) {
	var u User

	usersM.Lock()
	up, ok := users[login]
	if ok {
		u = *up
	}
	usersM.Unlock()

	if ok == false {
		return User{}, "", errors.New("login not found")
	}

	// Hash the given pass for comparing to the user we found in cache.
	// (I am aware that just storing a SHA256 version of a password is not considered
	//  too secure. Seeing as this is just a school project and authentication is not
	//  the main goal of the project, I don't feel to need to create a large auth package
	//  with salts and stuffs (nor to use an existing package that handles authentication).)
	hash := sha256.New()
	hash.Write([]byte(pass))
	passHash := hash.Sum(nil)

	if fmt.Sprintf("%x", passHash) != u.Pass {
		return User{}, "", errors.New("password incorrect")
	}

	// genKey doesn't read the values of User, just store the pointer
	key, err := initSession(up, time.Now())

	return u, key, err
}

// CheckKey returns the user for the session key if the given session key is valid.
// It will also set the access time for the session to the currect time.
// Returns error if the key is incorrect.
func CheckKey(key string) (User, error) {
	var s session
	var ok bool

	sessionsM.Lock()
	sp, ok := sessions[key]
	if ok {
		sp.accessTime = time.Now()
		s = *sp
	}
	sessionsM.Unlock()

	if ok == false {
		return User{}, errors.New("key incorrect")
	}

	// Check if this session is invalid because it is timed out.
	if time.Since(s.accessTime).Minutes() > sesTimeout {
		sessionsM.Lock()
		delete(sessions, s.key)
		sessionsM.Unlock()
		return User{}, errors.New("session timed out")
	}

	usersM.Lock()
	u := *s.u
	usersM.Unlock()

	return u, nil
}

// initSession creates a session for the given user with the given login time.
// The session key is returned.
func initSession(u *User, loginTime time.Time) (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Println("ERROR in auth.genKey, err:", err)
		return "", err
	}

	hash := sha256.New()
	hash.Write([]byte(strconv.Itoa(u.Id) + string(b)))
	key := fmt.Sprintf("%x", hash.Sum(nil)[:20])

	sessionsM.Lock()
	sessions[key] = &session{
		key:       key,
		u:         u,
		loginTime: loginTime,
	}
	sessionsM.Unlock()

	return key, nil
}

// updateCache resets and supplies the users map(/cache) with the users from the database.
func updateCache() error {
	var err error

	usersM.Lock()
	users = map[string]*User{}
	usersM.Unlock()

	rows, err := util.Db.Query("SELECT id, login, password FROM user;")
	if err != nil {
		log.Println("ERROR while updating users' cache, err:", err)
		return err
	}

	for rows.Next() {
		var u User

		err = rows.Scan(&u.Id, &u.Login, &u.Pass)
		if err != nil {
			log.Println("ERROR while formatting users in UpdateCache, err:", err)
			// Error will be returned at the end
		}

		usersM.Lock()
		users[u.Login] = &u
		usersM.Unlock()
	}

	return err
}

// Run takes care of auth. (Cleaning sessions map)
func Run() {
	// Fetch the data from the user table in the database.
	updateCache()

	// Ticker for cleaning the sessions map.
	clean := time.NewTicker(10 * time.Minute)

	for {
		select {
		// Remove 'garbage' from the sessions map (timed out sessions).
		case <-clean.C:
			var count int
			sessionsM.Lock()
			for _, s := range sessions {
				if time.Since(s.accessTime).Minutes() > sesTimeout {
					delete(sessions, s.key)
					count++
				}
			}
			sessionsM.Unlock()
			if count > 0 {
				fmt.Println(" Auth: Cleaned sessions, amount:", count)
			}
		}
	}
}

func CreateUser(login, pass string) (User, error) {
	return User{}, nil
}

func (u *User) ChangePass(newPass string) error {
	return nil
}
