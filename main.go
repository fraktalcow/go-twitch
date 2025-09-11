package main

import (
	"log"

	"go-twitch/config"
	"go-twitch/server"
	"go-twitch/twitch"
)

func main() {
	cfg := config.Load()

	if err := twitch.InitAppAccessToken(); err != nil {
		log.Fatalf("Error initializing Twitch app access token: %v", err)
	}

	go twitch.BotCommands()

	app := server.New(cfg)
	log.Fatal(app.Listen(":" + cfg.Port))
}
