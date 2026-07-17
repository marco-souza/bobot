package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port int
	Host string
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
	}
}
