package main

import "strconv"

type configuration struct {
	Telegram telegram              `json:"telegram"`
	Cronitor cronitorConfiguration `json:"cronitor,omitempty"`
}

type telegram struct {
	Token  string  `json:"token"`
	ChatID *chatID `json:"chat_id"`
}

// This struct is named cronitorConfiguration to avoid clashing with the cronitor package.
type cronitorConfiguration struct {
	TelemetryURL string `json:"telemetry_url"`
}

// TODO: Make these fields private and implement "encoding/json".Unmarshaler.
type chatID struct {
	ID       *uint64
	Username *string
}

func (c *chatID) String() string {
	switch {
	case c.ID != nil:
		return strconv.FormatUint(*c.ID, 10)
	case c.Username != nil:
		return *c.Username
	default:
		return ""
	}
}
