package client

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/chat-app/app/server"
	"github.com/gorilla/websocket"
	"github.com/jroimartin/gocui"
)

var (
	clientConnOnce sync.Once
)

const (
	yellowPrintedText = "\u001B[33;1m%s\u001B[0m\n"
)

type Client struct {
	destination   string
	usersEndPoint string
	Conn          *websocket.Conn
}

func (c *Client) Connect(force bool) (err error) {
	_connect := func() error {
		dialer := websocket.Dialer{}
		c.Conn, _, err = dialer.Dial(c.destination, nil)
		if err != nil {
			return err
		}
		return nil
	}

	// to bypass sync.Once (for example when username is taken and login is retried)
	if force {
		return _connect()
	}

	clientConnOnce.Do(func() {
		err = _connect()
	})

	if err != nil {
		return err
	}
	return
}

func (c *Client) ReadMessage() ([]byte, error) {
	_, p, err := c.Conn.ReadMessage()
	return p, err
}

func (c *Client) WriteMessage(msg string) error {
	return c.Conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

func (c *Client) Close() error {
	return c.Conn.Close()
}

// parseActivityMsg parses message received when user joins or leaves the channel
func (c *Client) parseActivityMsg(msg string) string {
	regexStr := fmt.Sprintf(`has\s(%s|%s)\sthe\schannel`, server.UserJoinedEvent, server.UserLeftEvent)
	regexUserActivity := regexp.MustCompile(regexStr)
	out := regexUserActivity.FindString(msg)
	if out != "" {
		return msg
	}
	return out
}

// getUserList pulls the user list from the REST endpoint and unmarshals it into a map
func (c *Client) getUserList() (*map[string]struct{}, error) {
	resp, err := http.Get(c.usersEndPoint)
	if err != nil {
		return nil, err
	}

	msg, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}
	m := &map[string]struct{}{}

	err = json.Unmarshal(msg, m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (c *Client) ReadRunner(w io.Writer, ifc *ClientInterface) {
	for {
		b, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			return
		}

		msg := string(b)

		ifc.gui.Update(func(gui *gocui.Gui) error {

			// special command that is only triggered by the server
			if strings.HasPrefix(msg, "//") {

				parsedMsg := c.parseActivityMsg(msg)

				// we are dealing with a user joining or leaving the chat
				if parsedMsg != "" {
					_, err = fmt.Fprintf(w, fmt.Sprintf(yellowPrintedText, msg))
					if err != nil {
						return err
					}

					m, err := c.getUserList()

					if err != nil {
						return err
					}
					userBarView := ifc.registeredViews[UserBarViewName].view
					userBarView.Clear()
					userBarView.SetCursor(0, 0)

					for k, _ := range *m {
						fmt.Fprintln(userBarView, k)
					}
				}
				// regular broadcasted chat messages
			} else {
				_, err = fmt.Fprintf(w, fmt.Sprintf("%s", msg))
				if err != nil {
					log.Println("read runner encountered error: ", err)
					return err
				}
				return nil
			}
			return nil
		})
	}
}

func InitClient() (*Client, error) {
	client := &Client{
		destination:   "ws://localhost:8080/ws",
		usersEndPoint: "http://localhost:8080/users",
	}

	return client, nil
}
