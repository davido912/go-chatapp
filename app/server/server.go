package server

import (
	"context"
	"encoding/json"
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
	s.mux.HandleFunc("/users", s.usersHandler)
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

		if err := <-s.hub.registrationErrorChan; err != nil {
			return
		}

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

func (s *Server) usersHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	bs, err := json.Marshal(s.loggedUsers)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = w.Write(bs)
	if err != nil {
		log.Println(err)
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

	serv.registerHandlers()

	err := http.ListenAndServe("localhost:8080", serv.mux) // edit later

	if err != nil {
		return err
	}

	return nil
}
