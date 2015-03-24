package ws

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type server struct {
	conns map[*conn]bool

	add chan *conn
	del chan *conn
}

type Handler struct {
	S *server
}

var Wss *server

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // This makes the server allow every origin's connection.
	},
}

func CreateServer() *server {
	Wss = &server{
		conns: make(map[*conn]bool),
		add:   make(chan *conn),
		del:   make(chan *conn),
	}
	go Wss.run()

	http.Handle("/ws", Handler{S: Wss})

	// For testing
	/*c := time.Tick(5 * time.Second)
	go func() {
		for _ = range c {
			Wss.Broadcast("b-itb4-1c", OutMsg{})
		}
	}()*/

	return Wss
}

func (s *server) run() {
	for {
		select {
		case c := <-s.add:
			s.conns[c] = true
			fmt.Printf(" WS active conns: %d\n", len(s.conns))
		case c := <-s.del:
			if _, ok := s.conns[c]; ok {
				c.close()
				delete(s.conns, c)
			} else {
				log.Println("Warning: ws.Run; <-s.del but can't find it the element in conns map?", c)
			}
		}
	}
}

func (s *server) Broadcast(page string, msg OutMsg) {
	var count int
	for c, _ := range s.conns {
		if page == "" || c.page == page {
			c.send <- msg
			count++
		}
	}
	if count > 0 {
		fmt.Printf(" WS broadcast; page: %s, effected conns: %d\n", page, count)
	}
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("ERROR while upgrading to ws, err:", err)
		return
	}

	c := &conn{
		ws:     ws,
		send:   make(chan OutMsg),
		chPage: make(chan string),
		s:      h.S,
	}
	c.s.add <- c

	defer func() {
		c.s.del <- c
	}()

	go c.writer()
	c.reader()
}
