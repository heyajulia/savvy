package main

import "github.com/heyajulia/energieprijzen/internal/telegram/chatid"

type configuration struct {
	Telegram telegramConfiguration `json:"telegram"`
	Bluesky  blueskyConfiguration  `json:"bluesky"`
	Cronitor cronitorConfiguration `json:"cronitor,omitempty"`
}

type telegramConfiguration struct {
	Token  string        `json:"token"`
	ChatID chatid.ChatID `json:"chat_id"`
}

type blueskyConfiguration struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type cronitorConfiguration struct {
	TelemetryURL string `json:"telemetry_url"`
}
