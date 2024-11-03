package main

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type configuration struct {
	Telegram telegram              `json:"telegram"`
	Bluesky  blueskyConfiguration  `json:"bluesky"`
	Cronitor cronitorConfiguration `json:"cronitor,omitempty"`
}

type telegram struct {
	Token  string  `json:"token"`
	ChatID *chatID `json:"chat_id"`
}

type blueskyConfiguration struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

// This struct is named cronitorConfiguration to avoid clashing with the cronitor package.
type cronitorConfiguration struct {
	TelemetryURL string `json:"telemetry_url"`
}

type chatID struct {
	id       *int64
	username *string
}

// Verify interface compliance.
var (
	_ json.Unmarshaler = (*chatID)(nil)
	_ fmt.Stringer     = (*chatID)(nil)
)

func (c *chatID) UnmarshalJSON(data []byte) error {
	var id int64
	if err := json.Unmarshal(data, &id); err == nil {
		c.id = &id
		return nil
	}

	var username string
	if err := json.Unmarshal(data, &username); err == nil {
		c.username = &username
		return nil
	}

	return fmt.Errorf("configuration: invalid chat_id: %s", data)
}

func (c *chatID) String() string {
	switch {
	case c.id != nil:
		return strconv.FormatInt(*c.id, 10)
	case c.username != nil:
		return *c.username
	default:
		return ""
	}
}
