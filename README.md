# bobot

[![Go Reference](https://pkg.go.dev/badge/github.com/marco-souza/bobot.svg)](https://pkg.go.dev/github.com/marco-souza/bobot)
[![Go Report Card](https://goreportcard.com/badge/github.com/marco-souza/bobot)](https://goreportcard.com/report/github.com/marco-souza/bobot)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![quality](https://github.com/marco-souza/bobot/actions/workflows/quality.yml/badge.svg)](https://github.com/marco-souza/bobot/actions/workflows/quality.yml)

A Discord gateway to [pi](https://github.com/marco-souza/pi) coding-agent
instances — talk to your agents from a server channel or DM instead of a
terminal. Written in Go on top of [`discordgo`](https://github.com/bwmarrin/discordgo).

## Overview

bobot bridges Discord and pi: each conversation (a DM, or an `@mention` / reply
in a server channel) is routed to a pi agent instance, and the agent's response
is posted back as a Discord reply tagged to the original message.

- DM bobot → a private agent session.
- `@bobot` in a server channel, or reply to one of bobot's messages → an agent
  session scoped to that thread.

Only the trigger routing is implemented today; pi instance invocation is the
next milestone.

## Usage

1. Create a Discord application at
   <https://discord.com/developers/applications>, then in **Bot → Privileged
   Gateway Intents** enable **Message Content Intent**.
2. Copy `.env` values from your application:

   ```ini
   DISCORD_APP_ID=...
   DISCORD_BOT_TOKEN=...
   ```

3. Run with [mise](https://mise.jdx.dev):

   ```sh
   mise install
   go run .
   ```

4. Open the invite URL bobot prints on startup to add it to your server. For
   private channels, give the bot's role **View Channel** + **Send Messages**
   on that channel.

## Contributing

Issues and pull requests are welcome on the
[repository](https://github.com/marco-souza/bobot). Keep changes small and
focused; run `go vet ./...` and `go build ./...` before submitting.