package chatid

import (
	"encoding"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type Kind int

const (
	KindID Kind = iota
	KindUsername
)

// Verify interface compliance.
var (
	_ encoding.TextUnmarshaler = (*ChatID)(nil)
	_ fmt.Stringer             = (*ChatID)(nil)
	_ json.Unmarshaler         = (*ChatID)(nil)
)

type ChatID struct {
	kind     Kind
	id       *int64
	username *string
}

func (c *ChatID) UnmarshalText(data []byte) error {
	s := string(data)

	if id, err := strconv.ParseInt(s, 10, 64); err == nil {
		c.kind = KindID
		c.id = &id
		return nil
	}

	if len(s) >= 2 && strings.HasPrefix(s, "@") {
		c.kind = KindUsername
		c.username = &s
		return nil
	}

	return fmt.Errorf("chatid: invalid id: %v", data)
}

func (c *ChatID) UnmarshalJSON(data []byte) error {
	var id int64
	if err := json.Unmarshal(data, &id); err == nil {
		c.kind = KindID
		c.id = &id
		return nil
	}

	var username string
	if err := json.Unmarshal(data, &username); err == nil {
		c.kind = KindUsername
		c.username = &username
		return nil
	}

	return fmt.Errorf("chatid: invalid id: %s", data)
}

func (c *ChatID) String() string {
	if c.kind == KindID {
		return strconv.FormatInt(*c.id, 10)
	}

	return *c.username
}

func (c *ChatID) Kind() Kind {
	return c.kind
}
