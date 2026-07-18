# bobot

[![Go Reference](https://pkg.go.dev/badge/github.com/marco-souza/bobot.svg)](https://pkg.go.dev/github.com/marco-souza/bobot)
[![Go Report Card](https://goreportcard.com/badge/github.com/marco-souza/bobot)](https://goreportcard.com/report/github.com/marco-souza/bobot)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![quality](https://github.com/marco-souza/bobot/actions/workflows/quality.yml/badge.svg)](https://github.com/marco-souza/bobot/actions/workflows/quality.yml)

A small Discord bot that replies to direct messages and `@mentions`, written
in Go on top of [`discordgo`](https://github.com/bwmarrin/discordgo).

## Overview

bobot listens for messages and responds when:

- a user sends it a private direct message; or
- a user `@mentions` it in a server channel; or
- a user replies to one of bobot's own messages.

Replies are sent as Discord reply references, so they tag the original message.

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