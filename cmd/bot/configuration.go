package main

import (
	"encoding/json"
	"fmt"
	"strconv"
)

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

type chatID struct {
	id       *uint64
	username *string
}

var (
	// Verify interface compliance.
	_ json.Unmarshaler = (*chatID)(nil)
	_ fmt.Stringer     = (*chatID)(nil)
)

func (c *chatID) String() string {
	switch {
	case c.id != nil:
		return strconv.FormatUint(*c.id, 10)
	case c.username != nil:
		return *c.username
	default:
		return ""
	}
}

func (c *chatID) UnmarshalJSON(data []byte) error {
	// Try to unmarshal data as a number.
	var id uint64
	if err := json.Unmarshal(data, &id); err == nil {
		c.id = &id
		return nil
	}

	// Try to unmarshal data as a string.
	var username string
	if err := json.Unmarshal(data, &username); err == nil {
		c.username = &username
		return nil
	}

	// Return an error if data is neither a number nor a string.
	return fmt.Errorf("chatID: cannot unmarshal %v", string(data))
}
