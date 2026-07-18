package discord

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/marco-souza/bobot/internal/pi"
)

// piTimeout bounds a single pi turn so a hung agent can't wedge the
// gateway; Discord replies usually arrive within a minute or two.
const piTimeout = 5 * time.Minute

// typingInterval refreshes the typing indicator while a pi turn is in flight;
// a single ChannelTyping burst expires after ~10s, so we resend before that.
const typingInterval = 8 * time.Second

// triggerTyping starts repeated typing bursts on ch until the returned stop
// function is called. Failures are swallowed: a missing indicator should never
// block the reply path.
func triggerTyping(s *discordgo.Session, ch string) (stop func()) {
	done := make(chan struct{})
	tick := func() { if err := s.ChannelTyping(ch); err != nil { log.Printf("typing: %v", err) } }
	tick()
	go func() {
		t := time.NewTicker(typingInterval)
		defer t.Stop()
		for {
			select {
			case <-done:
				return
			case <-t.C:
				tick()
			}
		}
	}()
	return func() { select { case <-done: default: close(done) } }
}

// reply sends msg as a Discord reply tagged to m.
func reply(s *discordgo.Session, m *discordgo.MessageCreate, msg string) {
	if _, err := s.ChannelMessageSendReply(m.ChannelID, msg, m.Reference()); err != nil {
		fmt.Printf("send reply: %v", err)
	}
}

// answer routes m through a pi agent instance and posts the reply.
// Session is keyed by m.ChannelID so a channel/thread/DM keeps shared memory;
// the prompt is prefixed with the asker's name so pi sees who said what in a
// multi-user channel without us tracking sender state.
func answer(s *discordgo.Session, m *discordgo.MessageCreate) {
	stop := triggerTyping(s, m.ChannelID)
	defer stop()

	sessionID := m.ChannelID
	prompt := fmt.Sprintf("%s asked: %s", m.Author.Username, m.Content)
	out, err := pi.AskTimeout(sessionID, prompt, piTimeout)
	if err != nil {
		fmt.Printf("pi answer: %v", err)
		out = "⚠️ couldn't reach the agent"
	}
	if out == "" {
		out = "(no response)"
	}
	reply(s, m, out)
}

// mentionsBot reports whether m explicitly @mentions bobot.
func mentionsBot(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	for _, u := range m.Mentions {
		if isSelf(s, u.ID) {
			return true
		}
	}
	return false
}

// replyingToBot reports whether m is a reply to one of bobot's messages,
// even when the reply ping is suppressed.
func replyingToBot(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	if ref := m.Message.ReferencedMessage; ref != nil {
		return isSelf(s, ref.Author.ID)
	}
	return false
}

func isDirectMessage(m *discordgo.MessageCreate) bool {
	return m.GuildID == ""
}

func isSelf(s *discordgo.Session, id string) bool {
	if s.State == nil || s.State.User == nil {
		return false
	}
	return id == s.State.User.ID
}

func token() string {
	return "Bot " + os.Getenv("DISCORD_BOT_TOKEN")
}

const botInviteURL = "https://discord.com/oauth2/authorize?client_id=%s&scope=bot"

func botInvite(appID string) string {
	return fmt.Sprintf(botInviteURL, appID)
}
