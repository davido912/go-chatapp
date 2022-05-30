package client

import (
	"github.com/jroimartin/gocui"
)

const (
	LoginViewName      = "login"
	LoginBoxViewName   = "loginbox"
	ChatWindowViewName = "chatwindow"
	UserBarViewName    = "userbar"
	MsgBoxViewName     = "msgbox"
)

// CustomView serves as a skeleton to all view structs
type CustomView struct {
	view     *gocui.View
	name     string       // logical object representation of the view
	title    string       // graphical title on the view
	editable bool         // whether a view can be fed input
	wrapText bool         // wraps text when it is longer than width of box
	frame    bool         // enable boarders around the view (graphic)
	x, y     int          // represents top screen axis
	x1, y1   int          // represents bottom screen axis
	initFunc func() error // func executed when view is initialised
}

func (cv *CustomView) Initialise(cl *ClientInterface) (err error) {
	cv.view, err = cl.gui.SetView(cv.name, cv.x, cv.y, cv.x1, cv.y1)
	if err != nil { // must include the assignments in this clause otherwise they'll be executed over and over again
		if err != gocui.ErrUnknownView {
			return err
		}

		if cv.title != "" {
			cv.view.Title = cv.title
		}

		if cv.editable {
			cv.view.Editable = cv.editable
		}

		if cv.wrapText {
			cv.view.Wrap = cv.wrapText
		}

		if cv.initFunc != nil {
			err = cv.initFunc()
			if err != nil {
				return err
			}
		}

		cl.registeredViews[cv.name] = cv
	}

	return nil
}

type KeyBinding struct {
	boundViewName string // viewname to apply the keybinding too
	key           interface{}
	modifier      gocui.Modifier // adds key cobinations
	handler       guiHandler     // function to execute
}

type ClientInterface struct {
	gui             *gocui.Gui
	UserClient      *Client
	registeredViews map[string]*CustomView
}

func NewClientInterface(userClient *Client) (*ClientInterface, error) {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return nil, err
	}

	ifc := &ClientInterface{
		registeredViews: make(map[string]*CustomView),
		UserClient:      userClient,
		gui:             g,
	}

	ifc.gui.SetManagerFunc(ifc.setLayout)

	keyBindings := getKeyBindings(ifc, ifc.UserClient)

	err = ifc.SetKeyBindings(keyBindings)

	if err != nil {
		return nil, err

	}
	return ifc, nil
}

func (cl *ClientInterface) Run() error {
	if err := cl.gui.MainLoop(); err != nil && err != gocui.ErrQuit {
		return err
	}
	return nil
}

func (cl *ClientInterface) setLayout(_ *gocui.Gui) error {
	maxX, maxY := cl.gui.Size()

	views := []*CustomView{
		{
			name:     ChatWindowViewName,
			x:        30,
			y:        0,
			x1:       maxX - 1,
			y1:       maxY - 4,
			title:    "Chat Room",
			editable: true,
			wrapText: true,
		},
		{
			name:     MsgBoxViewName,
			x:        0,
			y:        maxY - 3,
			x1:       maxX - 1,
			y1:       maxY - 1,
			title:    "Enter your message: ",
			editable: true,
		},
		{
			name:  UserBarViewName,
			x:     0,
			y:     0,
			x1:    29,
			y1:    maxY - 4,
			title: "Users Online",
		},
		{ // acts as background for the login box
			name:     LoginViewName,
			editable: true,
			x:        -1,
			y:        -1,
			x1:       maxX,
			y1:       maxY,
		},
		{
			name:     LoginBoxViewName,
			title:    "Choose your nickname: ",
			editable: true,
			x:        maxX/2 - 30,
			y:        maxY / 2,
			x1:       maxX/2 + 30,
			y1:       maxY/2 + 2,
			initFunc: func() error {
				return cl.SetCurrentView(LoginBoxViewName)
			},
		},
	}

	// registering the views in the GUI
	for _, v := range views {
		err := v.Initialise(cl)
		if err != nil {
			return err
		}
	}

	return nil
}

// SetKeyBindings assigns keybindings to the GUI
func (cl *ClientInterface) SetKeyBindings(kbs []*KeyBinding) error {
	for _, kb := range kbs {
		err := cl.gui.SetKeybinding(kb.boundViewName, kb.key, kb.modifier, kb.handler)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cl *ClientInterface) SetViewOnTop(viewname string) error {
	_, err := cl.gui.SetViewOnTop(viewname)
	return err
}

func (cl *ClientInterface) SetCurrentView(viewname string) error {
	_, err := cl.gui.SetCurrentView(viewname)
	return err
}

func (cl *ClientInterface) Close() {
	cl.gui.Close()
}

func getKeyBindings(ifc *ClientInterface, client *Client) []*KeyBinding {
	return []*KeyBinding{
		{
			boundViewName: "",
			key:           gocui.KeyCtrlC,
			modifier:      gocui.ModNone,
			handler: func(_ *gocui.Gui, _ *gocui.View) error {
				return gocui.ErrQuit
			},
		},
		{
			boundViewName: LoginBoxViewName,
			key:           gocui.KeyEnter,
			modifier:      gocui.ModNone,
			handler:       loginHandler(ifc, client),
		},
		{
			boundViewName: LoginBoxViewName,
			key:           gocui.KeyEnter,
			modifier:      gocui.ModNone,
			handler:       startReadRunnerHandler(ifc, client),
		},
		{
			boundViewName: MsgBoxViewName,
			key:           gocui.KeyEnter,
			modifier:      gocui.ModNone,
			handler:       chatMsgHandler(client),
		},
	}
}
