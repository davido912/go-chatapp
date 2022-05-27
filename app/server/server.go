package server

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strings"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type Server struct {
	mux *http.ServeMux
	*hub
}

func (s *Server) registerHandlers() {
	s.mux.HandleFunc("/ws", s.wsHandler)
	s.mux.HandleFunc("/", s.home)
}

// TODO: fix so that the msg isn't being sent to the user who sent it - rework,
func (s *Server) broadcastRunner(ctx context.Context, c chan Message) {
	for {
		select {

		case msg := <-c:
			func(msg Message) {
				s.mu.Lock()
				defer s.mu.Unlock()
				for k, v := range s.hub.loggedUsers {
					if k != msg.Username {
						sender := fmt.Sprintf("[%s] ", msg.Username)
						msgBody := append([]byte(sender), msg.Body...)
						err := v.WriteMessage(websocket.TextMessage, msgBody)
						if err != nil {
							log.Println("err ", err)
						}
					}
				}
			}(msg)

		case <-ctx.Done():
			log.Println("closing broadcast runner")
			return
		}
	}
}

func (s *Server) wsHandler(w http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		return
	}

	var user *User

	go func(conn *websocket.Conn) {
		defer func() {
			if err != nil {
				log.Printf("ws handler connection encountered error: %s\n \n err: %s \n", conn.LocalAddr(), err)
			}
			log.Println("closing ws connection")
			_ = conn.Close()
		}()

		// reading first message which is the username TODO: validation of validates
		_, p, err := conn.ReadMessage()

		if err != nil {
			return
		}

		user = &User{Username: Username(p), Conn: conn}
		s.hub.registerChan <- user

		defer func() {
			s.hub.deregisterChan <- user
		}()

		for {
			_, p, err := conn.ReadMessage()
			if err != nil {
				return
			}

			if len(p) > 0 {
				if strings.HasPrefix(string(p), "//") { // TODO: move all of these functions to validation

				} else {
					s.hub.broadcastChan <- Message{user.Username, p}
				}
			}

		}

	}(conn)
}

func (s *Server) home(w http.ResponseWriter, req *http.Request) {

	var out string

	for k, _ := range s.hub.loggedUsers {
		out += fmt.Sprintf("%q \n", k)
	}
	_, err := w.Write([]byte(out))
	if err != nil {
		log.Fatal(err)
		return
	}
}

func NewServer(ctx context.Context) error {

	h := NewHub()

	serv := Server{
		mux: http.NewServeMux(),
		hub: h,
	}

	go h.Run(ctx)
	go serv.broadcastRunner(ctx, h.broadcastChan)

	serv.registerHandlers()

	err := http.ListenAndServe("localhost:8080", serv.mux) // edit later

	if err != nil {
		return err
	}

	return nil
}
