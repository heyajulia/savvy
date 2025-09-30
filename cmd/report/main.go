package main

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"log/slog"
	"math/rand/v2"
	"os"
	"strings"
	"time"

	"github.com/heyajulia/savvy/internal"
	"github.com/heyajulia/savvy/internal/bsky"
	"github.com/heyajulia/savvy/internal/config"
	"github.com/heyajulia/savvy/internal/cronitor"
	"github.com/heyajulia/savvy/internal/datetime"
	"github.com/heyajulia/savvy/internal/prices"
	"github.com/heyajulia/savvy/internal/ranges"
	"github.com/heyajulia/savvy/internal/stamp"
	"github.com/heyajulia/savvy/internal/telegram"
	"github.com/heyajulia/savvy/internal/telegram/chatid"
	"github.com/heyajulia/savvy/internal/telegram/option"
)

var (
	//go:embed templates
	templatesFS embed.FS

	templates = template.Must(template.ParseFS(templatesFS, "templates/*.tmpl"))
)

func main() {
	showVersion := flag.Bool("v", false, "print version and exit")
	kickstart := flag.Bool("kickstart", false, "send the energy report immediately and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(internal.About())
		os.Exit(0)
	}

	slog.SetDefault(internal.Logger())

	slog.Info("application info", slog.Group("app", slog.String("version", internal.Version), slog.String("commit", internal.Commit)))

	if !*kickstart {
		d := time.Duration(rand.N(50)) * time.Second

		slog.Info("waiting before sending energy report", slog.Duration("duration", d))

		time.Sleep(d)
	}

	config, err := config.Read()
	if err != nil {
		slog.Error("configuration error", slog.Any("err", err))
		os.Exit(1)
	}

	token := config.Telegram.Token
	chatID := config.Telegram.ChatID
	blueskyIdentifier := config.Bluesky.Identifier
	blueskyPassword := config.Bluesky.Password
	cronitorURL := config.Cronitor.URL

	slog.Info("posting energy report")

	if *kickstart {
		if err := post(token, chatID, blueskyIdentifier, blueskyPassword); err != nil {
			slog.Error("could not post", slog.Any("err", err))
			os.Exit(1)
		}

		os.Exit(0)
	}

	monitor := cronitor.New(cronitorURL)
	if err := monitor.Monitor(func() error {
		return post(token, chatID, blueskyIdentifier, blueskyPassword)
	}); err != nil {
		slog.Error("failed to post", slog.Any("err", err))
	}
}

func post(token string, chatID chatid.ChatID, blueskyIdentifier, blueskyPassword string) error {
	s := stamp.New(os.Getenv("STAMP_DIR"))

	exists, err := s.Exists()
	if err != nil {
		return fmt.Errorf("check stamp: %w", err)
	}

	if exists {
		slog.Info("report already sent today")
		return nil
	}

	data, err := getTemplateData()
	if err != nil {
		return fmt.Errorf("get template data: %w", err)
	}

	short, long, err := report(*data)
	if err != nil {
		return fmt.Errorf("get reports: %w", err)
	}

	url, err := postToTelegram(long, token, chatID)
	if err != nil {
		return fmt.Errorf("post report to telegram: %w", err)
	}

	if err := postToBluesky(short, blueskyIdentifier, blueskyPassword, url); err != nil {
		return fmt.Errorf("post report to bluesky: %w", err)
	}

	if err := s.Stamp(); err != nil {
		return fmt.Errorf("create stamp: %w", err)
	}

	if err := s.Prune(); err != nil {
		return fmt.Errorf("prune stamps: %w", err)
	}

	return nil
}

func postToBluesky(report, username, password, url string) error {
	client, err := bsky.Login(username, password)
	if err != nil {
		return fmt.Errorf("login to bluesky: %w", err)
	}

	if err := client.Post(report, url); err != nil {
		return fmt.Errorf("post to bluesky: %w", err)
	}

	return nil
}

func getTemplateData() (*templateData, error) {
	p, err := internal.GetEnergyPrices()
	if err != nil {
		return nil, fmt.Errorf("get energy prices: %w", err)
	}

	now := datetime.Now()
	hello, goodbye := internal.GetGreeting(now)
	tomorrow := datetime.Tomorrow(now)

	average := p.Average()
	hourlies := make([]hourly, 24)

	for hour, price := range p.All() {
		hourlies[hour] = hourly{
			Emoji:          internal.GetPriceEmoji(price, average),
			PaddedHour:     fmt.Sprintf("%02d", hour),
			FormattedPrice: prices.Format(price),
		}
	}

	data := templateData{
		Short:            false,
		Hello:            hello,
		Goodbye:          goodbye,
		TomorrowDate:     datetime.Format(tomorrow),
		AverageFormatted: prices.Format(average),
		AverageHours:     ranges.CollapseAndFormat(p.AverageHours()),
		HighFormatted:    prices.Format(p.High()),
		HighHours:        ranges.CollapseAndFormat(p.HighHours()),
		LowFormatted:     prices.Format(p.Low()),
		LowHours:         ranges.CollapseAndFormat(p.LowHours()),
		Hourly:           hourlies,
	}

	return &data, nil
}

func report(data templateData) (short, long string, err error) {
	var sb strings.Builder

	data.Short = true
	if err := templates.ExecuteTemplate(&sb, "report.tmpl", data); err != nil {
		return "", "", fmt.Errorf("render short report: %w", err)
	}

	short = sb.String()
	sb.Reset()

	data.Short = false
	if err := templates.ExecuteTemplate(&sb, "report.tmpl", data); err != nil {
		return "", "", fmt.Errorf("render long report: %w", err)
	}

	long = sb.String()

	return
}

func postToTelegram(report, token string, chatID chatid.ChatID) (string, error) {
	slog.Info("sending message", slog.String("chat_id", chatID.String()), slog.String("message", report))

	bot := telegram.NewClient(token)

	message, err := bot.SendMessage(chatID, report, option.ParseModeHTML)
	if err != nil {
		return "", fmt.Errorf("send message: %w", err)
	}

	messageID := int64(message.ID)
	idLogger := slog.With(slog.Int64("message_id", messageID))

	idLogger.Info("message sent")

	if err := bot.SetMessageReaction(chatID, messageID); err != nil {
		// Not being able to react to the message is not the end of the world.
		idLogger.Warn("could not react to message", slog.Any("err", err))
	} else {
		idLogger.Info("message reacted to")
	}

	// FIXME: Harcoded channel name.
	return fmt.Sprintf("https://t.me/energieprijzen/%d", messageID), nil
}
