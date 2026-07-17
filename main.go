package main

import (
	"fmt"

	"github.com/marco-souza/bobot/internal/discord"
)

func main() {
	fmt.Println("bobot is here:", discord.Discord.InstallUrl())
}
