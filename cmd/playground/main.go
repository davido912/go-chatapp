package main

import (
	"fmt"
	"github.com/chat-app/app/client"
	"log"
	"sync"
	"time"
)

const (
	sent = "// %s has %s the channel"
)

func runner(ch chan int) {
	for {
		select {
		case i := <-ch:
			fmt.Println(i)
		default:
			fmt.Println("bey")
			return
		}
	}
}

func insert(ch chan int) {
	var i int
	for {
		time.Sleep(time.Millisecond * 200)
		ch <- i
		i++
	}
}

var (
	l     sync.Mutex
	users = make(map[string]string)
)

func Move() {
	l.Lock()
	defer l.Unlock()

	users["flower"] = "susan"
}

type Name struct {
	name string
}

func (n *Name) Add() {
	fmt.Println(n)
	m[n.name] = n
}

var m map[string]*Name = make(map[string]*Name)

type hand func(name string)

func a(next hand) hand {
	fmt.Println("running func")

	return func(s string) {
		fmt.Println("this is the returned")
		next(s)

	}
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered in f ", r)
		}
	}()
	_client, err := client.InitClient()
	if err != nil {
		fmt.Println(err)
		return
	}

	g, err := client.NewClientInterface(_client)
	if err != nil {
		log.Fatalln(err)
	}

	defer g.Close()

	err = g.Run()
	if err != nil {
		log.Fatalln(err)
	}

}
