package telegram

import (
	"fmt"
	"log/slog"
	"net/url"
	"strconv"

	"github.com/heyajulia/savvy/internal/telegram/chatid"
	"github.com/heyajulia/savvy/internal/telegram/option"
)

// Update represents a single Telegram update.
type Update struct {
	ID            int64
	Message       *Message
	CallbackQuery *CallbackQuery
}

func (u Update) IsMessage() bool {
	return u.Message != nil
}

func (u Update) IsCallbackQuery() bool {
	return u.CallbackQuery != nil
}

func (u Update) UserID() chatid.ChatID {
	if u.IsMessage() {
		return u.Message.From.ID
	}
	return u.CallbackQuery.From.ID
}

// Message represents a Telegram message.
type Message struct {
	ID   int64
	From User
	Text *string
}

// User represents a Telegram user.
type User struct {
	ID chatid.ChatID
}

// CallbackQuery represents an incoming callback query.
type CallbackQuery struct {
	ID      string
	From    User
	Message Message
	Data    string
}

// Context is passed to every handler and provides helpers for replying.
type Context struct {
	bot    *Bot
	update Update
}

// Update returns the raw update.
func (c *Context) Update() Update {
	return c.update
}

// UserID returns the chat ID of the user who triggered this update.
func (c *Context) UserID() chatid.ChatID {
	return c.update.UserID()
}

// Reply sends a text message to the same chat.
func (c *Context) Reply(text string, options ...option.Option) error {
	chatID := c.update.UserID()
	_, err := c.bot.doSendMessage(chatID, text, options...)
	return err
}

// ReplyHTML sends an HTML-formatted message to the same chat.
func (c *Context) ReplyHTML(text string, options ...option.Option) error {
	return c.Reply(text, append([]option.Option{option.ParseModeHTML}, options...)...)
}

// ReplyMarkdown sends a Markdown-formatted message to the same chat.
func (c *Context) ReplyMarkdown(text string, options ...option.Option) error {
	return c.Reply(text, append([]option.Option{option.ParseModeMarkdown}, options...)...)
}

// DeleteMessage deletes a message in the chat.
func (c *Context) DeleteMessage(messageID int64) error {
	return c.bot.deleteMessage(c.update.UserID(), messageID)
}

// AnswerCallbackQuery answers a callback query.
func (c *Context) AnswerCallbackQuery() error {
	if c.update.CallbackQuery == nil {
		return fmt.Errorf("telegram: not a callback query")
	}
	return c.bot.answerCallbackQuery(c.update.CallbackQuery.ID)
}

// SetMessageReaction sets a reaction on a message.
func (c *Context) SetMessageReaction(messageID int64) error {
	return c.bot.setMessageReaction(c.update.UserID(), messageID)
}

// Handler is a function that processes an update.
type Handler func(ctx *Context) error

// commandRoute maps command strings (e.g. "/start") to handlers.
type commandRoute struct {
	command string
	handler Handler
}

// callbackRoute maps callback data strings to handlers.
type callbackRoute struct {
	data    string
	handler Handler
}

// Bot is the high-level Telegram bot.
type Bot struct {
	token          string
	commands       []commandRoute
	callbacks      []callbackRoute
	unknownHandler Handler
}

// NewBot creates a new Bot with the given token.
func NewBot(token string) *Bot {
	return &Bot{
		token: token,
	}
}

// OnCommand registers a handler for a specific command (e.g. "/start").
func (b *Bot) OnCommand(command string, handler Handler) {
	b.commands = append(b.commands, commandRoute{command, handler})
}

// OnCallback registers a handler for a specific callback query data value.
func (b *Bot) OnCallback(data string, handler Handler) {
	b.callbacks = append(b.callbacks, callbackRoute{data, handler})
}

// OnUnknown registers a fallback handler for unrecognized commands and callbacks.
// If not set, unknown commands and callbacks are silently ignored.
func (b *Bot) OnUnknown(handler Handler) {
	b.unknownHandler = handler
}

// ProcessUpdates fetches and processes all pending updates, starting from offset.
// It returns the last processed update ID so callers can resume.
func (b *Bot) ProcessUpdates(offset int64) (int64, error) {
	rawUpdates, err := b.getUpdates(offset)
	if err != nil {
		return offset, err
	}

	for _, raw := range rawUpdates {
		update := Update{
			ID: raw.ID,
		}

		if raw.Message != nil {
			update.Message = &Message{
				ID:   raw.Message.ID,
				From: User{ID: raw.Message.From.ID},
				Text: raw.Message.Text,
			}
		}

		if raw.CallbackQuery != nil {
			update.CallbackQuery = &CallbackQuery{
				ID: raw.CallbackQuery.ID,
				From: User{
					ID: raw.CallbackQuery.From.ID,
				},
				Message: Message{
					ID: raw.CallbackQuery.Message.ID,
				},
				Data: raw.CallbackQuery.Data,
			}
		}

		ctx := &Context{bot: b, update: update}

		if err := b.dispatch(ctx); err != nil {
			slog.Error("handler error", slog.Any("err", err))
		}

		offset = update.ID
	}

	return offset, nil
}

func (b *Bot) dispatch(ctx *Context) error {
	switch {
	case ctx.update.IsMessage():
		return b.dispatchMessage(ctx)
	case ctx.update.IsCallbackQuery():
		return b.dispatchCallback(ctx)
	default:
		return nil
	}
}

func (b *Bot) dispatchMessage(ctx *Context) error {
	text := ctx.update.Message.Text
	if text == nil {
		return nil
	}

	for _, route := range b.commands {
		if *text == route.command {
			return route.handler(ctx)
		}
	}

	if b.unknownHandler != nil {
		return b.unknownHandler(ctx)
	}

	return nil
}

func (b *Bot) dispatchCallback(ctx *Context) error {
	data := ctx.update.CallbackQuery.Data

	for _, route := range b.callbacks {
		if data == route.data {
			return route.handler(ctx)
		}
	}

	if b.unknownHandler != nil {
		return b.unknownHandler(ctx)
	}

	return nil
}

// SendMessage sends a message to the given chat. It is a public method for
// non-interactive use (e.g. posting reports). For replying inside handlers,
// prefer Context.Reply.
func (b *Bot) SendMessage(chatID chatid.ChatID, text string, options ...option.Option) (*Message, error) {
	msg, err := b.doSendMessage(chatID, text, options...)
	if err != nil {
		return nil, err
	}

	return &Message{
		ID:   msg.ID,
		From: User{ID: msg.From.ID},
		Text: msg.Text,
	}, nil
}

// doSendMessage is the internal transport-level send used by both SendMessage
// and Context.Reply.
func (b *Bot) doSendMessage(chatID chatid.ChatID, text string, options ...option.Option) (*rawMessage, error) {
	parameters := url.Values{
		"chat_id": {chatID.String()},
		"text":    {text},
	}

	for _, o := range options {
		o(parameters)
	}

	msg, err := sendRequest[rawMessage](b.token, "sendMessage", parameters)
	if err != nil {
		return nil, fmt.Errorf("telegram: sendMessage: %w", err)
	}

	return msg, nil
}

// SetMessageReaction sets a reaction on a message in the given chat.
func (b *Bot) SetMessageReaction(chatID chatid.ChatID, messageID int64) error {
	return b.setMessageReaction(chatID, messageID)
}

func (b *Bot) deleteMessage(chatID chatid.ChatID, messageID int64) error {
	return b.fireOff("deleteMessage", url.Values{
		"chat_id":    {chatID.String()},
		"message_id": {strconv.FormatInt(messageID, 10)},
	})
}

func (b *Bot) answerCallbackQuery(id string) error {
	return b.fireOff("answerCallbackQuery", url.Values{
		"callback_query_id": {id},
	})
}

func (b *Bot) setMessageReaction(chatID chatid.ChatID, messageID int64) error {
	return b.fireOff("setMessageReaction", url.Values{
		"chat_id":    {chatID.String()},
		"message_id": {strconv.FormatInt(messageID, 10)},
		"is_big":     {"true"},
		"reaction":   {reaction},
	})
}

func (b *Bot) getUpdates(offset int64) ([]rawUpdate, error) {
	updates, err := sendRequest[[]rawUpdate](b.token, "getUpdates", url.Values{
		"offset":          {strconv.FormatInt(offset, 10)},
		"timeout":         {"60"},
		"allowed_updates": {allowedUpdates},
	})
	if err != nil {
		return nil, fmt.Errorf("telegram: getUpdates: %w", err)
	}

	return *updates, nil
}

func (b *Bot) fireOff(method string, parameters url.Values) error {
	if _, err := sendRequest[any](b.token, method, parameters); err != nil {
		return fmt.Errorf("telegram: %s: %w", method, err)
	}

	return nil
}
