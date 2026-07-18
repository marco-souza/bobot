package discord

import (
	"fmt"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
)

// reply sends msg as a Discord reply tagged to m.
func reply(s *discordgo.Session, m *discordgo.MessageCreate, msg string) {
	if _, err := s.ChannelMessageSendReply(m.ChannelID, msg, m.Reference()); err != nil {
		log.Printf("send reply: %v", err)
	}
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
