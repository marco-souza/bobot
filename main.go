package main

import (
	"log"

	"github.com/marco-souza/bobot/internal/discord"
)

func main() {
	if err := discord.Run(); err != nil {
		log.Fatal(err)
	}
}
