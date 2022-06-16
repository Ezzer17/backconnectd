package main

import (
	"flag"
	"log"
	"os"

	"github.com/ezzer17/backconnectd/internal/config"
	"github.com/ezzer17/backconnectd/internal/server"
)

func parseFlags() string {
	var configPath string
	flag.StringVar(&configPath, "config", "./config.yml", "path to config file")
	flag.Parse()
	return configPath
}

func main() {
	var logger *log.Logger
	cfgPath := parseFlags()
	cfg, err := config.New(cfgPath)
	if err != nil {
		log.Fatal(err)
	}
	logfile, err := os.OpenFile(cfg.Logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Printf("Error opening logfile: %s", err)
		logger = log.Default()
	} else {
		defer logfile.Close()
		logger = log.New(logfile, "", log.LstdFlags)
	}
	srv := server.New(logger)
	go srv.BackconnectLoop(cfg.BackconnectAddr)
	srv.AdminLoop(cfg.AdminAddr)
}
