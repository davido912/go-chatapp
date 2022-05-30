package ui

import (
	"bufio"
	"bytes"
	"github.com/chat-app/app/client"
	"github.com/jroimartin/gocui"
)

type guiHandler func(g *gocui.Gui, v *gocui.View) error

// loginHandler reads the nickname input by the user and writes it to the server
// afterwards it hides the login views and moves to the normal chat views
func loginHandler(ifc *ClientInterface, client *client.Client) guiHandler {
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

func chatMsgHandler(client *client.Client) guiHandler {
	return func(g *gocui.Gui, v *gocui.View) error {
		v.Rewind() // ensure reading from beginning of view's buffer

		bs := make([]byte, 1024)
		n, err := v.Read(bs)
		if err != nil {
			return err
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
