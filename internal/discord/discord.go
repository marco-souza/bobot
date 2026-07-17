package discord

import (
	"fmt"
	"os"
)

type discord struct {
	AppID     string
	PublicKey string
	BotToken  string
}

var Discord *discord

func init() {
	Discord = &discord{
		AppID:     os.Getenv("DISCORD_APP_ID"),
		PublicKey: os.Getenv("DISCORD_PUBLIC_KEY"),
		BotToken:  os.Getenv("DISCORD_BOT_TOKEN"),
	}
}

func (d *discord) InstallUrl() string {
	return fmt.Sprintf(
		"https://discord.com/oauth2/authorize?client_id=%s",
		d.AppID,
	)
}
