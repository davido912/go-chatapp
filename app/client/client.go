package client

import (
	"encoding/json"
	"fmt"
	"github.com/chat-app/app/server"
	"github.com/gorilla/websocket"
	"github.com/jroimartin/gocui"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
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

	m := &map[string]struct{}{
		"Savasdfid_0":            struct{}{},
		"Savasdfid_1":            struct{}{},
		"Savasdfid_2":            struct{}{},
		"Savasdfid_3":            struct{}{},
		"Savasdfid_4":            struct{}{},
		"Savasdfid_5":            struct{}{},
		"Savasdfid_6":            struct{}{},
		"Savasdfid_7":            struct{}{},
		"Savasdfid_8":            struct{}{},
		"Savasdfid_9":            struct{}{},
		"Savasdfid_10":           struct{}{},
		"Savasdfid_11":           struct{}{},
		"Savasdfid_12":           struct{}{},
		"Savasdfid_13":           struct{}{},
		"Savasdfid_14":           struct{}{},
		"Savasdfid_15":           struct{}{},
		"Savasdfid_16":           struct{}{},
		"Savasdfid_17":           struct{}{},
		"Savasdfid_18":           struct{}{},
		"Savasdfid_19":           struct{}{},
		"Savasdfid_20":           struct{}{},
		"Savasdfid_21":           struct{}{},
		"Savasdfid_22":           struct{}{},
		"Savasdfid_23":           struct{}{},
		"Savasdfid_24":           struct{}{},
		"Savasdfid_25":           struct{}{},
		"Savasdfid_26":           struct{}{},
		"Savasdfid_27":           struct{}{},
		"Savasdfid_28":           struct{}{},
		"Savasdfid_29":           struct{}{},
		"Savasdfid_30":           struct{}{},
		"Savasdfid_31":           struct{}{},
		"Savasdfid_32":           struct{}{},
		"Savasdfid_33":           struct{}{},
		"Savasdfid_34":           struct{}{},
		"Savasdfid_35":           struct{}{},
		"Savasdfid_36":           struct{}{},
		"Savasdfid_37":           struct{}{},
		"Savasdfid_38":           struct{}{},
		"Savasdfid_39":           struct{}{},
		"Savasdfid_40":           struct{}{},
		"Savasdfid_41":           struct{}{},
		"Savasdfid_42":           struct{}{},
		"Savasdfid_43":           struct{}{},
		"Savasdfid_44":           struct{}{},
		"Savasdfid_45":           struct{}{},
		"Savasdfid_46":           struct{}{},
		"Savasdfid_47":           struct{}{},
		"Savasdfid_48":           struct{}{},
		"Savasdfid_49":           struct{}{},
		"Savasdfid_880":          struct{}{},
		"Savasdfid_881":          struct{}{},
		"Savasdfid_8888":         struct{}{},
		"Savasdfid_88882":        struct{}{},
		"Savasdfid_88882111":     struct{}{},
		"Savasdfid_885":          struct{}{},
		"Savasdfid_886":          struct{}{},
		"Savasdfid_887":          struct{}{},
		"Savasdfid_888":          struct{}{},
		"Savasdfid_889":          struct{}{},
		"Savasdfid_8820":         struct{}{},
		"Savasdfid_8821":         struct{}{},
		"Savasdfid_88288":        struct{}{},
		"Savasdfid_882882":       struct{}{},
		"Savasdfid_882882111":    struct{}{},
		"Savasdfid_8825":         struct{}{},
		"Savasdfid_8826":         struct{}{},
		"Savasdfid_8827":         struct{}{},
		"Savasdfid_8828":         struct{}{},
		"Savasdfid_8829":         struct{}{},
		"Savasdfid_8821110":      struct{}{},
		"Savasdfid_8821111":      struct{}{},
		"Savasdfid_88211188":     struct{}{},
		"Savasdfid_882111882":    struct{}{},
		"Savasdfid_882111882111": struct{}{},
		"Savasdfid_8821115":      struct{}{},
		"Savasdfid_8821116":      struct{}{},
		"Savasdfid_8821117":      struct{}{},
		"Savasdfid_8821118":      struct{}{},
		"Savasdfid_8821119":      struct{}{},
	}

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
