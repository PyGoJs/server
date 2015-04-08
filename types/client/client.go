package client

import (
	"crypto/sha256"
	"fmt"
	"log"
	"sync"

	"github.com/pygojs/server/util"
)

type Client struct {
	Id         int
	secretHash string // Secret is given with a request as authentication.
	Fac        string
}

// Client.SecretHash - Client
var clients = map[string]*Client{}
var clientsM = &sync.Mutex{}

func Get(secret string) (Client, bool) {
	// A SHA256 version of the secret is stored in the database, and in Client.
	// Hash the given secret with SHA256 to get the Client associated with that secret.
	hash := sha256.New()
	hash.Write([]byte(secret))
	secretHash := hash.Sum(nil)

	clientsM.Lock()
	clp, ok := clients[fmt.Sprintf("%x", secretHash)]
	var cl Client
	if ok {
		cl = *clp
	}
	clientsM.Unlock()

	return cl, ok
}

func UpdateCache() error {
	clientsM.Lock()
	clients = map[string]*Client{}
	clientsM.Unlock()

	rows, err := util.Db.Query("SELECT id, secret, facility FROM client LIMIT 100;")
	if err != nil {
		log.Println("ERROR while updating client cache, err:", err)
		return err
	}

	for rows.Next() {
		// (Not sure whether or not to make cl a pointer)
		cl := &Client{}

		err = rows.Scan(&cl.Id, &cl.secretHash, &cl.Fac)
		if err != nil {
			log.Println("ERROR while formatting clients in UpdateCache, err:", err)
			return err
		}

		clientsM.Lock()
		clients[cl.secretHash] = cl
		clientsM.Unlock()
	}

	return nil
}
