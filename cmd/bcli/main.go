package main

import (
	"flag"

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
	c, err := client.Connect(serverAddress)
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
