package main

import (
	"flag"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/gdamore/tcell/v2"

	"github.com/ezzer17/backconnectd/internal/client"
	"github.com/ezzer17/backconnectd/internal/menu"
)

func parseFlags() string {
	var serverAddress string
	flag.StringVar(&serverAddress, "server", "127.0.0.1:2222", "Address of backconnectd server")
	flag.Parse()
	return serverAddress
}

func main() {
	serverAddress := parseFlags()
	conn, err := grpc.Dial(serverAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	c := client.New(conn)
	go c.Subscribe()
	if err != nil {
		panic(err)
	}
	screen, err := tcell.NewScreen()
	if err != nil {
		panic(err)
	}
	m := menu.New(screen, c)
	if err := m.Init(); err != nil {
		panic(err)
	}
	m.RunTillExit()
}
