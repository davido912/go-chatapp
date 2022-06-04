package server

import (
	"errors"
	"reflect"
	"testing"

	"github.com/gorilla/websocket"
)

func TestRegister(t *testing.T) {

	tcs := []struct {
		description         string
		inputUser           *User
		inputLoggedUsers    map[Username]*websocket.Conn
		expectedError       error
		expectedLoggedUsers map[Username]*websocket.Conn
	}{
		{
			description:      "successful user registration",
			inputUser:        &User{Username("user1"), nil},
			inputLoggedUsers: map[Username]*websocket.Conn{},
			expectedError:    nil,
			expectedLoggedUsers: map[Username]*websocket.Conn{
				"user1": nil,
			},
		},
		{
			description: "username taken",
			inputUser:   &User{Username("user1"), nil},
			inputLoggedUsers: map[Username]*websocket.Conn{
				"user1": nil,
			},
			expectedError: UserNameTakenError{"user1"},
			expectedLoggedUsers: map[Username]*websocket.Conn{
				"user1": nil,
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.description, func(t *testing.T) {
			h := NewHub()
			h.loggedUsers = tc.inputLoggedUsers

			var err error

			go func() {
				<-h.registrationErrorChan
			}()

			err = h.register(tc.inputUser)

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("expected err to be %q but got %q\n", tc.expectedError, err)
			}

			if !reflect.DeepEqual(h.loggedUsers, tc.expectedLoggedUsers) {
				t.Errorf("expected logged users to be equal")
			}

		})

	}
}

func TestDeregister(t *testing.T) {

	tcs := []struct {
		description         string
		inputUser           *User
		inputLoggedUsers    map[Username]*websocket.Conn
		expectedLoggedUsers map[Username]*websocket.Conn
	}{
		{
			description: "successful user deregistration",
			inputUser:   &User{Username("user1"), nil},
			inputLoggedUsers: map[Username]*websocket.Conn{
				"user1": nil,
			},
			expectedLoggedUsers: map[Username]*websocket.Conn{},
		},
		{
			description: "user not found",
			inputUser:   &User{Username("user2"), nil},
			inputLoggedUsers: map[Username]*websocket.Conn{
				"user1": nil,
			},
			expectedLoggedUsers: map[Username]*websocket.Conn{
				"user1": nil,
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.description, func(t *testing.T) {
			h := NewHub()
			h.loggedUsers = tc.inputLoggedUsers

			h.deregister(tc.inputUser)

			if !reflect.DeepEqual(h.loggedUsers, tc.expectedLoggedUsers) {
				t.Errorf("expected logged users to be equal")
			}

		})

	}
}
