package ws

import "github.com/gorilla/websocket"

type conn struct {
	// The websocket itself.
	ws   *websocket.Conn
	send chan OutMsg

	// Page the conn is currently viewing (class)
	page   string
	chPage chan string

	// Every conn has a pointer to the server it's connected to.
	s *server
}

func (c *conn) reader() {
	for {
		var msg inMsg
		err := c.ws.ReadJSON(&msg)
		if err != nil {
			//log.Println("ERROR while reading msg from ws, err:", err)
			break
		}

		if msg.Page != "" {
			c.chPage <- msg.Page
		} else {
			c.send <- OutMsg{Error: "unknown message send"}
		}
	}
	c.ws.Close()
}

func (c *conn) writer() {
forLoop:
	for /*msg := range c.send*/ {
		select {
		case msg := <-c.send:
			err := c.ws.WriteJSON(msg)
			if err != nil {
				//log.Println("ERROR writing msg in ws client, err:", err)
				break forLoop
			}
		case page := <-c.chPage:
			c.page = page
		}
	}
	c.ws.Close()
}

func (c *conn) close() {
	close(c.send)
	close(c.chPage)
}
