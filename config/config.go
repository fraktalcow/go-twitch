package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds application configuration sourced from environment variables.
type Config struct {
	Port         string
	ClientID     string
	ClientSecret string
	RedirectURI  string
	BotUsername  string
	UserToken    string
	AppToken     string
}

// Load reads environment variables (from .env if present) and returns Config.
func Load() Config {
	// Load .env but do not fail hard if missing to mirror previous behavior
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env not loaded: %v", err)
	}

	cfg := Config{
		Port:         getenvDefault("PORT", "3000"),
		ClientID:     os.Getenv("TWITCH_CLIENT_ID"),
		ClientSecret: os.Getenv("TWITCH_CLIENT_SECRET"),
		RedirectURI:  os.Getenv("TWITCH_REDIRECT_URI"),
		BotUsername:  os.Getenv("TWITCH_BOT_USERNAME"),
		UserToken:    os.Getenv("TWITCH_USER_ACCESS_TOKEN"),
		AppToken:     os.Getenv("TWITCH_APP_ACCESS_TOKEN"),
	}
	return cfg
}

func getenvDefault(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}
