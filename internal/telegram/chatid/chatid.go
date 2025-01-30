package chatid

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type ChatIDKind int

const (
	KindID ChatIDKind = iota
	KindUsername
)

// Verify interface compliance.
var (
	_ json.Unmarshaler = (*ChatID)(nil)
	_ fmt.Stringer     = (*ChatID)(nil)
)

type ChatID struct {
	kind     ChatIDKind
	id       *int64
	username *string
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

func (c *ChatID) Kind() ChatIDKind {
	return c.kind
}
