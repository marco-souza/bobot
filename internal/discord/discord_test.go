package discord

import (
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
)

const botID = "bobot"

// newSession returns a Session with just enough State for isSelf-based
// predicates; no gateway connection.
func newSession(t *testing.T) *discordgo.Session {
	t.Helper()
	s := &discordgo.Session{State: &discordgo.State{}}
	s.State.User = &discordgo.User{ID: botID}
	return s
}

func TestIsDirectMessage(t *testing.T) {
	for _, tc := range []struct {
		name string
		m    *discordgo.MessageCreate
		want bool
	}{
		{"dm", &discordgo.MessageCreate{Message: &discordgo.Message{GuildID: ""}}, true},
		{"server", &discordgo.MessageCreate{Message: &discordgo.Message{GuildID: "guild"}}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := isDirectMessage(tc.m); got != tc.want {
				t.Fatalf("isDirectMessage=%v want %v", got, tc.want)
			}
		})
	}
}

func TestIsSelf(t *testing.T) {
	for _, tc := range []struct {
		name string
		s    *discordgo.Session
		id   string
		want bool
	}{
		{"nil state", &discordgo.Session{}, botID, false},
		{"nil user", &discordgo.Session{State: &discordgo.State{}}, botID, false},
		{"match", newSession(t), botID, true},
		{"no match", newSession(t), "other", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := isSelf(tc.s, tc.id); got != tc.want {
				t.Fatalf("isSelf=%v want %v", got, tc.want)
			}
		})
	}
}

func TestMentionsBot(t *testing.T) {
	msg := func(mentions ...*discordgo.User) *discordgo.MessageCreate {
		return &discordgo.MessageCreate{Message: &discordgo.Message{Mentions: mentions}}
	}
	other := &discordgo.User{ID: "other"}
	self := &discordgo.User{ID: botID}

	for _, tc := range []struct {
		name string
		m    *discordgo.MessageCreate
		want bool
	}{
		{"empty", msg(), false},
		{"only other", msg(other), false},
		{"self among others", msg(other, self), true},
		{"only self", msg(self), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := mentionsBot(newSession(t), tc.m); got != tc.want {
				t.Fatalf("mentionsBot=%v want %v", got, tc.want)
			}
		})
	}
}

func TestReplyingToBot(t *testing.T) {
	msg := func(ref *discordgo.Message) *discordgo.MessageCreate {
		return &discordgo.MessageCreate{Message: &discordgo.Message{ReferencedMessage: ref}}
	}
	selfMsg := &discordgo.Message{Author: &discordgo.User{ID: botID}}
	otherMsg := &discordgo.Message{Author: &discordgo.User{ID: "other"}}

	for _, tc := range []struct {
		name string
		m    *discordgo.MessageCreate
		want bool
	}{
		{"no reference", msg(nil), false},
		{"reply to self", msg(selfMsg), true},
		{"reply to other", msg(otherMsg), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := replyingToBot(newSession(t), tc.m); got != tc.want {
				t.Fatalf("replyingToBot=%v want %v", got, tc.want)
			}
		})
	}
}

func TestBotInvite(t *testing.T) {
	got := botInvite("123")
	want := "https://discord.com/oauth2/authorize?client_id=123&scope=bot"
	if got != want {
		t.Fatalf("botInvite=%q want %q", got, want)
	}
}

func TestTruncateForDiscord(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want string
	}{
		{
			"short passes through",
			"hello",
			"hello",
		},
		{
			"at limit passes through",
			strings.Repeat("x", discordMsgLimit),
			strings.Repeat("x", discordMsgLimit),
		},
		{
			"over limit capped with ellipsis",
			strings.Repeat("x", discordMsgLimit+50),
			strings.Repeat("x", discordMsgLimit-1) + "…",
		},
		{
			"multibyte rune-safe boundary",
			strings.Repeat("�", discordMsgLimit+10),
			strings.Repeat("�", discordMsgLimit-1) + "…",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := truncateForDiscord(tc.in)
			if got != tc.want {
				t.Fatalf("truncateForDiscord len_in=%d len_out=%d want_len=%d",
					len(tc.in), len(got), len(tc.want))
			}
			// hard guarantee: result never exceeds Discord's cap.
			if len([]rune(got)) > discordMsgLimit {
				t.Fatalf("result %d runes exceeds limit %d", len([]rune(got)), discordMsgLimit)
			}
		})
	}
}