package client

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/jroimartin/gocui"
	"io"
	"strings"
	"time"
)

type guiHandler func(g *gocui.Gui, v *gocui.View) error

var (
	connectionReady = make(chan struct{}, 1)
)

// loginHandler reads the nickname input by the user and writes it to the server
// afterwards it hides the login views and moves to the normal chat views
func loginHandler(ifc *ClientInterface, client *Client) guiHandler {
	return func(g *gocui.Gui, v *gocui.View) error {
		err := client.Connect(true)
		if err != nil {
			return err
		}

		v.Rewind() // ensure reading from beginning of view's buffer

		bs := make([]byte, 1024)
		n, err := v.Read(bs)
		if err != nil {
			return err
		}

		buf := bytes.NewBuffer(bs)
		scanner := bufio.NewScanner(buf)
		scanner.Scan()

		err = client.WriteMessage(scanner.Text())
		if err != nil {
			return err
		}

		resp, err := client.ReadMessage()
		if err != nil {
			return err
		}

		if string(resp) != "0" {
			nickname := strings.TrimSuffix(string(bs[:n]), "\n") // removing newline
			msg := fmt.Sprintf("username %q is already taken", nickname)
			err := ifc.DisplayError(msg)
			if err != nil {
				return err
			}
			return nil // make sure it's terminated also with the children
		}

		connectionReady <- struct{}{}

		// after login was successful, change to the relevant views
		nextViews := []string{ChatWindowViewName, UserBarViewName, MsgBoxViewName}

		for _, v := range nextViews {
			err := ifc.SetViewOnTop(v)
			if err != nil {
				return err
			}
		}

		err = ifc.SetCurrentView(MsgBoxViewName)
		if err != nil {
			return err
		}

		return nil
	}
}

// after connection was established, start ready runner
func startReadRunnerHandler(ifc *ClientInterface, client *Client) guiHandler {
	return func(g *gocui.Gui, v *gocui.View) error {
		go func() {
			for {
				select {
				case <-connectionReady:
					chatBoxView := ifc.registeredViews[ChatWindowViewName].view
					go client.ReadRunner(chatBoxView, ifc)
					return
				case <-time.After(time.Second * 10): //rework this to be ctx context
					return
				}
			}
		}()
		return nil

	}
}

func chatMsgHandler(client *Client) guiHandler {
	return func(g *gocui.Gui, v *gocui.View) error {
		v.Rewind() // ensure reading from beginning of view's buffer

		bs := make([]byte, 1024)
		n, err := v.Read(bs)
		if err != nil {
			if err != io.EOF {
				return err
			}

		}

		err = client.WriteMessage(string(bs[:n]))
		if err != nil {
			return err
		}

		v.Clear()
		v.SetCursor(0, 0)

		return nil
	}
}

func nextView(ifc *ClientInterface, setViewName string, viewNames ...string) guiHandler {
	return func(g *gocui.Gui, v *gocui.View) error {
		err := ifc.SetViewOnBottom(v.Name())
		if err != nil {
			return err
		}

		for _, vName := range viewNames {
			err = ifc.SetViewOnTop(vName)
			if err != nil {
				return err
			}

			if vName == setViewName {
				err = ifc.SetCurrentView(vName)
				if err != nil {
					return err
				}
			}

		}
		return nil
	}
}

func scrollView(ifc *ClientInterface, setViewName string, dy int) guiHandler {
	return func(g *gocui.Gui, v *gocui.View) error {
		ifc.SetCurrentView(UserBarViewName)
		vi := ifc.registeredViews[UserBarViewName].view

		if vi != nil {
			vi.Autoscroll = false
			ox, oy := vi.Origin()
			if err := vi.SetOrigin(ox, oy+dy); err != nil {
				return nil
			}
		}

		return nil

	}
}
