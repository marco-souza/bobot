package main

import (
	"fmt"

	"github.com/marco-souza/bobot/config"
)

func main() {
	authUrl := fmt.Sprintf(
		"https://discord.com/oauth2/authorize?client_id=%s",
		config.Discord.AppID,
	)

	fmt.Println("bobot is here:", authUrl)
}
