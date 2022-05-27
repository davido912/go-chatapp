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

// TODO, feels inefficient to iterate twice over the logged users list
func (h *hub) updateOnlineUsers(user *User, userEvent string) error {
	msg := fmt.Sprintf(`// %s has %s the channel`, user.Username, userEvent)
	for _, v := range h.loggedUsers {
		err := v.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *hub) Run(ctx context.Context) {
	for {
		select {
		case user := <-h.registerChan:
			h.register(user)
			err := h.updateOnlineUsers(user, UserJoinedEvent)
			if err != nil {
				log.Printf("encountered error when updating users: %s", err)
			}
		case <-ctx.Done():
			log.Println("closing hub")
			return
		case user := <-h.deregisterChan:
			h.deregister(user)
			err := h.updateOnlineUsers(user, UserLeftEvent)
			if err != nil {
				log.Printf("encountered error when updating users: %s", err)
			}
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
