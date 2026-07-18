package discord

import (
	"fmt"
	"log/slog"
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
	s.AddHandler(onInteractionCreate)
	s.Identify.Intents = discordgo.IntentsDirectMessages |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsMessageContent

	appID := os.Getenv("DISCORD_APP_ID")
	slog.Info("invite link ready", "url", botInvite(appID))

	if err := s.Open(); err != nil {
		return fmt.Errorf("open discord connection: %w", err)
	}
	defer s.Close()

	if err := registerCommands(s, appID); err != nil {
		return fmt.Errorf("register commands: %w", err)
	}

	slog.Info("bobot running; press Ctrl+C to exit")

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
		answer(s, m)

	case mentionsBot(s, m), replyingToBot(s, m):
		answer(s, m)
	}
}
