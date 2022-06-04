package client

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jroimartin/gocui"
)

type guiHandler func(g *gocui.Gui, v *gocui.View) error

var (
	// used to indicate the ReadRunner that the client established successful connection
	connectionReady = make(chan struct{}, 1)
)

// loginHandler reads the nickname input by the user and writes it to the server
// afterwards it hides the login views and moves to the normal chat views
func loginHandler(ifc *ClientInterface, client *Client) guiHandler {
	return func(g *gocui.Gui, v *gocui.View) error {
		v.Rewind() // ensure reading from beginning of view's buffer

		bs := make([]byte, 1024)
		n, err := v.Read(bs)
		if err != nil {
			return err
		}

		nickname := strings.TrimSuffix(string(bs[:n]), "\n") // removing newline
		buf := bytes.NewBuffer([]byte(nickname))

		// TODO: rework constant url here
		resp, err := http.Post("http://localhost:8080/login", "text/plain", buf)
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			msg := fmt.Sprintf("username %q is already taken", nickname)
			err := ifc.DisplayError(msg)
			if err != nil {
				return err
			}
			return nil // make sure it's terminated also with the children
		}

		err = client.Connect(true)
		if err != nil {
			return err
		}
		if err := client.WriteMessage(nickname); err != nil {
			return err
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

func exitViewHandler(ifc *ClientInterface, client *Client) guiHandler {
	return func(g *gocui.Gui, v *gocui.View) error {

		exitView := ifc.registeredViews[ExitWindowViewName]
		err := ifc.SetViewOnTop(exitView.name)
		if err != nil {
			return err
		}

		err = ifc.SetCurrentView(exitView.name)
		if err != nil {
			return err
		}

		// in this way we keep track which view was set current before
		exitView.prevCurrentView = ifc.registeredViews[v.Name()]

		return nil

	}
}

func exitHandler(ifc *ClientInterface, client *Client) guiHandler {
	return func(g *gocui.Gui, v *gocui.View) error {

		exitView := ifc.registeredViews[ExitWindowViewName]
		
		bs := make([]byte, 1024)
		exitView.view.Rewind()
		n, err := exitView.view.Read(bs)

		defer func() {
			exitView.view.Clear()
			v.SetCursor(0, 0)
		}()

		if err != nil {
			if !errors.Is(err, io.EOF) {
				return err
			}

		}
		if strings.ToLower(string(bs[:n])) == "y\n" {
			return gocui.ErrQuit
		}

		ifc.SetViewOnBottom(exitView.name)
		ifc.SetCurrentView(exitView.prevCurrentView.name)

		return nil

	}
}
