package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"os"
	"slices"
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
	"github.com/urfave/cli/v3"
)

func reportCommand() *cli.Command {
	return &cli.Command{
		Name:  "report",
		Usage: "Generate and send the daily energy price report",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "kickstart",
				Usage: "Send report immediately without random delay",
			},
		},
		Action: runReport,
	}
}

func runReport(ctx context.Context, c *cli.Command) error {
	kickstart := c.Bool("kickstart")

	slog.SetDefault(internal.Logger())

	slog.Info("application info", slog.Group("app", slog.String("version", internal.Version), slog.String("commit", internal.Commit)))

	if !kickstart {
		d := time.Duration(rand.N(50)) * time.Second

		slog.Info("waiting before sending energy report", slog.Duration("duration", d))

		time.Sleep(d)
	}

	cfg, err := config.Read[config.Report]()
	if err != nil {
		slog.Error("configuration error", slog.Any("err", err))
		os.Exit(1)
	}

	token := cfg.Telegram.Token
	chatID := cfg.Telegram.ChatID
	channelName := cfg.Telegram.ChannelName
	blueskyIdentifier := cfg.Bluesky.Identifier
	blueskyPassword := cfg.Bluesky.Password
	cronitorURL := cfg.Cronitor.URL
	stampDir := cfg.StampDir

	slog.Info("posting energy report")

	if kickstart {
		if err := post(token, chatID, channelName, blueskyIdentifier, blueskyPassword, stampDir); err != nil {
			slog.Error("could not post", slog.Any("err", err))
			os.Exit(1)
		}

		return nil
	}

	monitor := cronitor.New(cronitorURL)
	if err := monitor.Monitor(func() error {
		return post(token, chatID, channelName, blueskyIdentifier, blueskyPassword, stampDir)
	}); err != nil {
		slog.Error("failed to post", slog.Any("err", err))
	}

	return nil
}

func post(token string, chatID chatid.ChatID, channelName, blueskyIdentifier, blueskyPassword, stampDir string) error {
	s := stamp.New(stampDir)

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

	url, err := postToTelegram(long, token, chatID, channelName)
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

type templateData struct {
	Short            bool
	Hello            string
	Goodbye          string
	TomorrowDate     string
	AverageFormatted string
	AverageHours     string
	HighFormatted    string
	HighHours        string
	LowFormatted     string
	LowHours         string
	Hourly           []hourly
}

type hourly struct {
	Emoji          string
	PaddedHour     string
	FormattedPrice string
}

func getTemplateData() (*templateData, error) {
	p, err := internal.GetEnergyPrices()
	if err != nil {
		return nil, fmt.Errorf("get energy prices: %w", err)
	}

	now := datetime.Now()
	hello, goodbye := internal.GetGreeting(now)
	tomorrow := datetime.Tomorrow(now)
	hourlyHours := hourNumbersForDay(tomorrow, p.Len())

	average := p.Average()
	hourlies := make([]hourly, 0, len(hourlyHours))

	for hour, price := range p.All() {
		if hour >= len(hourlyHours) {
			break
		}

		actualHour := hourlyHours[hour]

		hourlies = append(hourlies, hourly{
			Emoji:          internal.GetPriceEmoji(price, average),
			PaddedHour:     fmt.Sprintf("%02d", actualHour),
			FormattedPrice: prices.Format(price),
		})
	}

	data := templateData{
		Short:            false,
		Hello:            hello,
		Goodbye:          goodbye,
		TomorrowDate:     datetime.Format(tomorrow),
		AverageFormatted: prices.Format(average),
		AverageHours:     formatHourRanges(p.AverageHours(), hourlyHours),
		HighFormatted:    prices.Format(p.High()),
		HighHours:        formatHourRanges(p.HighHours(), hourlyHours),
		LowFormatted:     prices.Format(p.Low()),
		LowHours:         formatHourRanges(p.LowHours(), hourlyHours),
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

func formatHourRanges(indexes, hours []int) string {
	if len(indexes) == 0 || len(hours) == 0 {
		return ""
	}

	unique := make(map[int]struct{}, len(indexes))
	dedup := make([]int, 0, len(indexes))

	for _, idx := range indexes {
		if idx < 0 || idx >= len(hours) {
			continue
		}

		hour := hours[idx]

		if _, ok := unique[hour]; ok {
			continue
		}

		unique[hour] = struct{}{}
		dedup = append(dedup, hour)
	}

	if len(dedup) == 0 {
		return ""
	}

	slices.Sort(dedup)

	return ranges.CollapseAndFormat(dedup)
}

func hourNumbersForDay(day time.Time, count int) []int {
	if count <= 0 {
		return nil
	}

	loc := day.Location()
	startOfDay := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, loc)
	startOfDayUTC := startOfDay.UTC()

	hours := make([]int, 0, count)

	for i := range count {
		slot := startOfDayUTC.Add(time.Duration(i) * time.Hour).In(loc)
		hours = append(hours, slot.Hour())
	}

	return hours
}

func postToTelegram(report, token string, chatID chatid.ChatID, channelName string) (string, error) {
	slog.Info("sending message", slog.String("chat_id", chatID.String()), slog.String("message", report))

	bot := telegram.NewClient(token)

	message, err := bot.SendMessage(chatID, report, option.ParseModeHTML)
	if err != nil {
		return "", fmt.Errorf("send message: %w", err)
	}

	messageID := int64(message.ID)
	idLogger := slog.With(slog.Int64("message_id", messageID))

	idLogger.Info("message sent")

	// Not being able to react to the message is not the end of the world.
	if err := bot.SetMessageReaction(chatID, messageID); err != nil {
		idLogger.Warn("could not react to message", slog.Any("err", err))
	} else {
		idLogger.Info("message reacted to")
	}

	return fmt.Sprintf("https://t.me/%s/%d", channelName, messageID), nil
}
