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

	UsernameTakenErrorText = "username %q is already taken"
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
	loggedUsers           map[Username]*websocket.Conn
	mu                    sync.Mutex
	registerChan          chan *User
	registrationErrorChan chan error
	deregisterChan        chan *User
	broadcastChan         chan Message
}

func (h *hub) register(user *User) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	log.Println("registering user: ", user.Username)
	if _, ok := h.loggedUsers[user.Username]; ok {
		err := user.Conn.WriteMessage(websocket.TextMessage, []byte("1")) // user already exists
		if err != nil {
			log.Println("hit error when registering user: ", err)
			return err
		}
		usernameTakenErr := fmt.Errorf(UsernameTakenErrorText, user.Username)
		h.registrationErrorChan <- usernameTakenErr

		return usernameTakenErr
	}

	err := user.Conn.WriteMessage(websocket.TextMessage, []byte("0"))
	if err != nil {
		log.Println("hit error when registering user: ", err)
		return err
	}
	h.loggedUsers[user.Username] = user.Conn
	h.registrationErrorChan <- nil

	return nil
}

func (h *hub) deregister(user *User) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.loggedUsers[user.Username]; ok {
		log.Println("deregistering user: ", user.Username)
		delete(h.loggedUsers, user.Username)
	} else {
		log.Printf("user %q is not found", user.Username)
	}
}

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

// broadcasts a message to all currently connected users TODO: include error for when a user logs out while lock is still obtained (check error that websocket is closed)
func (h *hub) broadcast(msg Message) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, v := range h.loggedUsers {
		sender := fmt.Sprintf("[%s] ", msg.Username)
		msgBody := append([]byte(sender), msg.Body...)
		err := v.WriteMessage(websocket.TextMessage, msgBody)
		if err != nil {
			log.Println(err)
		}
	}
}

func (h *hub) Run(ctx context.Context) {
	// goroutine to broadcast msgs in channel
	go func() {
		for {
			select {
			case msg := <-h.broadcastChan:
				h.broadcast(msg)
			}
		}
	}()

	for {
		select {
		case user := <-h.registerChan:
			err := h.register(user)
			if err != nil {
				log.Println(err)
			} else {

				err = h.updateOnlineUsers(user, UserJoinedEvent)
				if err != nil {
					log.Printf("encountered error when updating users: %s", err)
				}
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
		loggedUsers:           make(map[Username]*websocket.Conn),
		mu:                    sync.Mutex{},
		registerChan:          make(chan *User),
		registrationErrorChan: make(chan error),
		deregisterChan:        make(chan *User),
		broadcastChan:         make(chan Message),
	}
}
