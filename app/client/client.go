package client

import (
	"fmt"
	"github.com/chat-app/app/server"
	"github.com/gorilla/websocket"
	"github.com/jroimartin/gocui"
	"io"
	"log"
	"strings"
	"sync"
)

var (
	clientConnOnce sync.Once
)

type Client struct {
	destination string
	Conn        *websocket.Conn
	LoggedUsers map[string]bool
}

func (c *Client) Connect() (err error) {
	clientConnOnce.Do(func() {
		dialer := websocket.Dialer{}
		c.Conn, _, err = dialer.Dial(c.destination, nil)
		if err != nil {
			return
		}
	})
	if err != nil {
		return err
	}
	return
}

func (c *Client) readMessage() ([]byte, error) {
	_, p, err := c.Conn.ReadMessage()
	return p, err
}

func (c *Client) WriteMessage(msg string) error {
	return c.Conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

func (c *Client) Close() error {
	return c.Conn.Close()
}

func (c *Client) updateOnlineUsersLoc(username, activity string) {
	if activity == server.UserJoinedEvent {
		c.LoggedUsers[username] = true
		return
	}
	delete(c.LoggedUsers, username)
}

// TODO: refactor
func (c *Client) ReadRunner(w io.Writer, g *gocui.Gui) {

	for {
		b, err := c.readMessage()
		if err != nil {
			log.Println("read runner encountered error: ", err)
		}

		msg := string(b)
		// if activity msg indicating leaving or join, parse it, break it and update map
		// nice if we can just erase where the location is
		testview, _ := g.View("userbar")

		if strings.HasPrefix(msg, "//") {
			msgArr := strings.Split(msg, " ")
			c.updateOnlineUsersLoc(msgArr[1], msgArr[3])
			testview.Clear()
			for k, _ := range c.LoggedUsers {
				fmt.Fprintln(testview, k)
			}
		}

		g.Update(func(gui *gocui.Gui) error {
			_, err = fmt.Fprintf(w, fmt.Sprintf("%s\n", msg))
			if err != nil {
				log.Println("read runner encountered error: ", err)
				return err
			}
			return nil
		})
	}
}

func InitClient() (*Client, error) {
	client := &Client{
		destination: "ws://localhost:8080/ws",
		LoggedUsers: make(map[string]bool),
	}

	err := client.Connect()
	if err != nil {
		return nil, err
	}

	return client, nil
}
