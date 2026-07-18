package discord

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

// Run starts the bot and blocks until interrupted by SIGINT/SIGTERM.
func Run() error {
	s, err := discordgo.New(token())
	if err != nil {
		return fmt.Errorf("create discord session: %w", err)
	}

	s.AddHandler(onMessageCreate)
	s.Identify.Intents = discordgo.IntentsDirectMessages |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsMessageContent

	fmt.Printf("Invite bobot: %s\n", botInvite(os.Getenv("DISCORD_APP_ID")))

	if err := s.Open(); err != nil {
		return fmt.Errorf("open discord connection: %w", err)
	}
	defer s.Close()

	log.Println("bobot is running. Press Ctrl+C to exit.")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	return nil
}

// onMessageCreate routes an incoming message to the right reply path.
func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}

	switch {
	case isDirectMessage(m):
		reply(s, m, "Hey! What's up?")

	case mentionsBot(s, m), replyingToBot(s, m):
		reply(s, m, "You called?")
	}
}
