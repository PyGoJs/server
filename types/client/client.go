package client

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"log"
)

type Client struct {
	Id         int
	secretHash string // Secret is given with a request as authentication.
	Fac        string
}

// Client.SecretHash - Client
var clients = map[string]Client{}

func Get(secret string) (Client, bool) {
	// A SHA256 version of the secret is stored in the database, and in Client.
	// Hash the given secret with SHA256 to get the Client associated with that secret.
	hash := sha256.New()
	hash.Write([]byte(secret))
	secretHash := hash.Sum(nil)

	cl, ok := clients[fmt.Sprintf("%x", secretHash)]
	return cl, ok
}

func UpdateCache(db *sql.DB) error {
	clients = map[string]Client{}

	rows, err := db.Query("SELECT id, secret, facility FROM client LIMIT 100;")
	if err != nil {
		log.Println("ERROR while updating client cache, err:", err)
		return err
	}

	for rows.Next() {
		// (Not sure whether or not to make cl a pointer)
		cl := Client{}

		err = rows.Scan(&cl.Id, &cl.secretHash, &cl.Fac)
		if err != nil {
			log.Println("ERROR while formatting clients in UpdateCache, err:", err)
			return err
		}

		clients[cl.secretHash] = cl
	}

	return nil
}
