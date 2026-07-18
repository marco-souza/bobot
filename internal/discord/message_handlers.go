package discord

import (
	"fmt"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/marco-souza/bobot/internal/pi"
)

// piTimeout bounds a single pi turn so a hung agent can't wedge the
// gateway; Discord replies usually arrive within a minute or two.
const piTimeout = 5 * time.Minute

// reply sends msg as a Discord reply tagged to m.
func reply(s *discordgo.Session, m *discordgo.MessageCreate, msg string) {
	if _, err := s.ChannelMessageSendReply(m.ChannelID, msg, m.Reference()); err != nil {
		fmt.Printf("send reply: %v", err)
	}
}

// answer routes m through a pi agent instance and posts the reply.
func answer(s *discordgo.Session, m *discordgo.MessageCreate) {
	out, err := pi.AskTimeout(m.Content, piTimeout)
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
