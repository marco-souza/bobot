package config

import (
	"os"
	"strconv"
)

type discord struct {
	AppID     string
	PublicKey string
	BotToken  string
}

type Config struct {
	Port    int
	Host    string
	Discord discord
}

var Env *Config

func init() {
	host := os.Getenv("HOST")
	if host == "" {
		host = "http://localhost"
	}

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		port = 8080
	}

	Env = &Config{
		Host: host,
		Port: port,
		Discord: discord{
			AppID:     os.Getenv("DISCORD_API_KEY"),
			PublicKey: os.Getenv("DISCORD_PUBLIC_KEY"),
			BotToken:  os.Getenv("DISCORD_BOT_TOKEN"),
		},
	}
}
