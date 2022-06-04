package server

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

const (
	UserJoinedEvent = "joined"
	UserLeftEvent   = "left"
)

type UserNameTakenError struct{ username string }

func (e UserNameTakenError) Error() string {
	return fmt.Sprintf("username %q is taken", e.username)
}

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
	var err error

	log.Println("registering user: ", user.Username)
	if _, ok := h.loggedUsers[user.Username]; ok {
		err = UserNameTakenError{string(user.Username)}
		h.registrationErrorChan <- err
		return err
	}

	h.loggedUsers[user.Username] = user.Conn
	h.registrationErrorChan <- err

	return err
}

func (h *hub) deregister(user *User) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.loggedUsers[user.Username]; ok {
		log.Println("deregistering user: ", user.Username)
		delete(h.loggedUsers, user.Username)
	} else {
		log.Printf("user %q not found", user.Username)
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
// 2022/05/31 16:43:23 write tcp 127.0.0.1:8080->127.0.0.1:60675: use of closed network connection
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
