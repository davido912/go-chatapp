package client

import (
	"bufio"
	"bytes"
	"github.com/jroimartin/gocui"
	"io"
)

type guiHandler func(g *gocui.Gui, v *gocui.View) error

var (
	connectionReady = make(chan struct{}, 1)
)

// loginHandler reads the nickname input by the user and writes it to the server
// afterwards it hides the login views and moves to the normal chat views
func loginHandler(ifc *ClientInterface, client *Client) guiHandler {
	return func(g *gocui.Gui, v *gocui.View) error {
		v.Rewind() // ensure reading from beginning of view's buffer

		bs := make([]byte, 1024)
		_, err := v.Read(bs)
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
		<-connectionReady
		chatBoxView := ifc.registeredViews[ChatWindowViewName].view
		go client.ReadRunner(chatBoxView, ifc)
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
