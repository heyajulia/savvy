package chatid

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Verify interface compliance.
var (
	_ json.Unmarshaler = (*ChatID)(nil)
	_ fmt.Stringer     = (*ChatID)(nil)
)

type ChatID struct {
	id       *int64
	username *string
}

func (c *ChatID) UnmarshalJSON(data []byte) error {
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

	return fmt.Errorf("chatid: invalid id: %s", data)
}

func (c *ChatID) String() string {
	if c.id != nil {
		return strconv.FormatInt(*c.id, 10)
	}

	return *c.username
}
