package ws

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/pygojs/server/util"
)

type server struct {
	// Conn (client), by page the conn is viewing.
	conns map[*conn]string

	// BroadCast, also contains the page that conns need to be viewing to get the msg.
	bc chan OutMsg

	add chan *conn
	del chan *conn

	// Client askes server to change conns current page.
	chPage chan chPage
}

// See server.chPage
type chPage struct {
	c    *conn  // So server knows which conn.
	page string // New page string
}

type Handler struct {
	S *server
}

var Wss *server

// upgrader is used for... well basicly upgrading a regular HTTP request into a websocket connection.
// See ws.ServeHTTP .
var upgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// This makes the server allow every origin's connection.
		// Don't worry, it should only be for devving.
		return true
	},
}

// NewServer returns a running instance of ws.server that listens to the given url (e.q. /ws).
func NewServer(url string) *server {
	ws := &server{
		conns:  make(map[*conn]string),
		bc:     make(chan OutMsg),
		add:    make(chan *conn),
		del:    make(chan *conn),
		chPage: make(chan chPage),
	}

	go ws.run()

	http.Handle(url, Handler{S: ws})

	// For testing
	/*c := time.Tick(5 * time.Second)
	go func() {
		for _ = range c {
			Wss.Broadcast("b-itb4-1c", OutMsg{})
		}
	}()*/

	return ws
}

// run checks channels and, when it needs to, does stuff to the server instance it is used on.
func (s *server) run() {
	for {
		select {
		case c := <-s.add:
			s.conns[c] = "home"
		case c := <-s.del:
			if _, ok := s.conns[c]; ok {
				c.close()
				delete(s.conns, c)
			} else {
				log.Println("Warning: ws.Run; <-s.del but can't find it the element in conns map?", c)
			}
			fmt.Printf(" WS active conns: %d\n", len(s.conns))
		case p := <-s.chPage:
			s.conns[p.c] = p.page
		case m := <-s.bc: // Broadcast
			var count int
			for c, p := range s.conns {
				// Ignore this conn if it's not on the correct page.
				if m.dest.page != "" && p != m.dest.page {
					fmt.Println(m.dest.page, p)
					continue
				}
				select {
				case c.send <- m:
					count++
				default:
					fmt.Println("Can't send msg")
					// Couldn't send msg, delete conn.
					c.close()
					delete(s.conns, c)
				}
			}
			if count > 0 {
				fmt.Printf(" WS broadcast; page: %s, effected conns: %d\n",
					m.dest.page, count)
			}
		}
	}
}

// Broadcast sends the given OutMsg as JSON to all conns on the given page.
func (s *server) Broadcast(page string, msg OutMsg) {
	msg.dest.page = page
	s.bc <- msg
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	util.LogS("%s ws", util.Ip(*r))

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("ERROR while upgrading to ws, err:", err)
		return
	}

	c := &conn{
		ws:   ws,
		send: make(chan OutMsg),
		s:    h.S,
	}
	c.s.add <- c

	defer func() {
		c.s.del <- c
	}()

	go c.writer()
	c.reader()
}
