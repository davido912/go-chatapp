package server

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"sync"
)

const (
	UserJoinedEvent = "joined"
	UserLeftEvent   = "left"
)

type Message struct {
	Username
	Body []byte
}

type Username string

type User struct {
	Username
	Conn *websocket.Conn
}

type hub struct {
	loggedUsers    map[Username]*websocket.Conn // for now all users are registered in the map right away
	mu             sync.Mutex
	registerChan   chan *User
	deregisterChan chan *User
	broadcastChan  chan Message
}

func (h *hub) register(user *User) {
	h.mu.Lock()
	defer h.mu.Unlock()
	log.Println("registering user: ", user.Username)
	h.loggedUsers[user.Username] = user.Conn
}

func (h *hub) deregister(user *User) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.loggedUsers[user.Username]; ok {
		log.Println("deregistering user: ", user.Username)
		delete(h.loggedUsers, user.Username)
	}
}

//func (h *hub) sendUserList(conn *websocket.Conn) error {
//
//	userlist := "//userlist: "
//	usernames := make([]string, len(h.loggedUsers))
//
//	var i int
//	for k, _ := range h.loggedUsers {
//		usernames[i] = string(k)
//		i++
//	}
//	userlist += strings.Join(usernames, ";")
//	err := conn.WriteMessage(websocket.TextMessage, []byte(userlist))
//	if err != nil {
//		return err
//	}
//	return nil
//}

// send a message to all users in regards to a user joining / leaving TODO: rework to use database
func (h *hub) updateOnlineUsers(user *User, userEvent string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	msg := fmt.Sprintf(`// %s has %s the channel`, user.Username, userEvent)

	for _, conn := range h.loggedUsers {
		err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			return err
		}
	}
	return nil
}

// broadcasts a message to all currently connected users TODO: include error for when a user logs out while lock is still obtained
func (h *hub) broadcast(msg Message) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, v := range h.loggedUsers {
		sender := fmt.Sprintf("[%s] ", msg.Username)
		msgBody := append([]byte(sender), msg.Body...)
		err := v.WriteMessage(websocket.TextMessage, msgBody)
		if err != nil {
			log.Println("err ", err)
		}
	}
}

func (h *hub) Run(ctx context.Context) {
	go func() {
		for { // TODO: consider whether a select is required here, maybe later when more functionality is added
			select {
			case msg := <-h.broadcastChan:
				h.broadcast(msg)
			}
		}
	}()

	for {
		select {
		case user := <-h.registerChan:
			h.register(user)
			err := h.updateOnlineUsers(user, UserJoinedEvent)
			if err != nil {
				log.Printf("encountered error when updating users: %s", err)
			}

		case user := <-h.deregisterChan:
			h.deregister(user)
			err := h.updateOnlineUsers(user, UserLeftEvent)
			if err != nil {
				log.Printf("encountered error when updating users: %s", err)
			}

		case <-ctx.Done():
			log.Println("closing hub")
			return

		}
	}
}

func NewHub() *hub {
	return &hub{
		loggedUsers:    make(map[Username]*websocket.Conn),
		mu:             sync.Mutex{},
		registerChan:   make(chan *User),
		deregisterChan: make(chan *User),
		broadcastChan:  make(chan Message),
	}
}
