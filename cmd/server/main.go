package main

import (
	"context"
	"github.com/chat-app/app/server"
	"log"
)

func main() {
	var err error

	//doneCh := make(chan os.Signal, 1)
	//signal.Notify(doneCh, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		log.Println("closing server")
		cancel()
		log.Fatal(err)
	}()

	err = server.NewServer(ctx)
	if err != nil {
		return
	}

}
