package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"text/template"

	"github.com/heyajulia/energieprijzen/internal"
	"github.com/heyajulia/energieprijzen/internal/ranges"
)

var (
	token, chatID string
)

func init() {
	ensureFlag := func(name string) {
		found := false

		flag.Visit(func(f *flag.Flag) {
			if f.Name == name {
				found = true
			}
		})

		if !found {
			flag.Usage()
			os.Exit(0)
		}
	}

	flag.StringVar(&token, "token", "", "Telegram bot token")
	flag.StringVar(&chatID, "chat-id", "", "Telegram chat ID")

	flag.Parse()

	ensureFlag("token")
	ensureFlag("chat-id")
}

func main() {
	prices, err := internal.GetEnergyPrices()
	if err != nil {
		panic(err)
	}

	var sb strings.Builder

	tmpl := template.Must(template.New("message").Funcs(template.FuncMap{
		"AddCharges":          internal.AddCharges,
		"CollapseAndFormat":   ranges.CollapseAndFormat,
		"FormatCurrencyValue": internal.FormatCurrencyValue,
		"GetPriceEmoji":       internal.GetPriceEmoji,
		"Pad":                 func(i int) string { return fmt.Sprintf("%02d", i) },
	}).Parse(message))

	g := internal.GetGreeting()
	hello := g.Hello
	goodbye := g.Goodbye

	err = tmpl.Execute(&sb, struct {
		Hello        string
		Goodbye      string
		TomorrowDate string
		*internal.EnergyPrices
	}{
		Hello:        hello,
		Goodbye:      goodbye,
		TomorrowDate: internal.GetTomorrowDate(),
		EnergyPrices: prices,
	})

	if err != nil {
		panic(err)
	}

	s := sb.String()

	resp, err := http.PostForm("https://api.telegram.org/bot"+token+"/sendMessage", url.Values{
		"chat_id":    {chatID},
		"text":       {s},
		"parse_mode": {"HTML"},
	})
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic(fmt.Errorf("unexpected status code: %d", resp.StatusCode))
	}
}

const message = `{{.Hello}}De energieprijzen van {{.TomorrowDate}} zijn bekend.

Gemiddeld: {{FormatCurrencyValue (AddCharges .Average)}} per kWh {{CollapseAndFormat .AverageHours}}
Hoog: {{FormatCurrencyValue (AddCharges .High)}} per kWh {{CollapseAndFormat .HighHours}}
Laag: {{FormatCurrencyValue (AddCharges .Low)}} per kWh {{CollapseAndFormat .LowHours}}

Alle prijzen van morgen per uur:

<code>` +

	"{{range .Prices}}" +
	"{{GetPriceEmoji (AddCharges .Price) (AddCharges $.Average)}} {{Pad .Hour}}:00 – {{Pad .Hour}}:59: {{FormatCurrencyValue (AddCharges .Price)}} per kWh\n" +
	"{{end}}</code>\n\n{{.Goodbye}}"
