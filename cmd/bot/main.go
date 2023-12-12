package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"strings"
	"text/template"
	_ "time/tzdata"

	"github.com/heyajulia/energieprijzen/internal"
	"github.com/heyajulia/energieprijzen/internal/date"
	"github.com/heyajulia/energieprijzen/internal/ranges"
)

var (
	dryRun        bool
	token, chatID string
)

func init() {
	flag.BoolVar(&dryRun, "d", false, "dry run")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)

	// No need for a token or a chat ID if we're not sending a message.
	if dryRun {
		return
	}

	path := os.ExpandEnv("$CREDENTIALS_DIRECTORY/token")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Fatalln("Credentials file does not exist. Are you running this using systemd?")
	}

	f, err := os.Open(path)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		log.Fatalln(err)
	}

	token = strings.TrimSpace(string(b))

	if c, ok := os.LookupEnv("ENERGIEPRIJZEN_BOT_CHAT_ID"); ok {
		chatID = c
	} else {
		log.Fatalln("ENERGIEPRIJZEN_BOT_CHAT_ID is not set")
	}
}

func main() {
	user, err := user.Current()
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Running as user %s [ID: %s]\n", user.Username, user.Uid)

	prices, err := internal.GetEnergyPrices()
	if err != nil {
		log.Fatalln(err)
	}

	var sb strings.Builder

	log.Println("Setting up template...")

	tmpl := template.Must(template.New("message").Funcs(template.FuncMap{
		"AddCharges":          internal.AddCharges,
		"CollapseAndFormat":   ranges.CollapseAndFormat,
		"FormatCurrencyValue": internal.FormatCurrencyValue,
		"GetPriceEmoji":       internal.GetPriceEmoji,
		"Pad":                 func(i int) string { return fmt.Sprintf("%02d", i) },
	}).Parse(message))

	log.Println("Getting greeting...")

	hello, goodbye := internal.GetGreeting()

	log.Println("Executing template...")

	err = tmpl.Execute(&sb, struct {
		Hello        string
		Goodbye      string
		TomorrowDate string
		*internal.EnergyPrices
	}{
		Hello:        hello,
		Goodbye:      goodbye,
		TomorrowDate: date.Tomorrow(),
		EnergyPrices: prices,
	})

	if err != nil {
		log.Fatalln(err)
	}

	message := sb.String()

	log.Printf("message: %#v\n", message)

	if dryRun {
		log.Println("Dry run, not sending message")

		return
	}

	log.Println("Sending message...")

	resp, err := http.PostForm("https://api.telegram.org/bot"+token+"/sendMessage", url.Values{
		"chat_id":    {chatID},
		"text":       {message},
		"parse_mode": {"HTML"},
	})
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("status code: %d, body: %#v\n", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Unexpected status code: %d\n", resp.StatusCode)
	}
}

const message = `{{.Hello}} De energieprijzen van {{.TomorrowDate}} zijn bekend.

Gemiddeld: {{FormatCurrencyValue (AddCharges .Average)}} per kWh {{CollapseAndFormat .AverageHours}}
Hoog: {{FormatCurrencyValue (AddCharges .High)}} per kWh {{CollapseAndFormat .HighHours}}.
Laag: {{FormatCurrencyValue (AddCharges .Low)}} per kWh {{CollapseAndFormat .LowHours}}.

Alle prijzen van morgen per uur:

<code>` +

	"{{range .Prices}}" +
	"{{GetPriceEmoji (AddCharges .Price) (AddCharges $.Average)}} {{Pad .Hour}}:00 – {{Pad .Hour}}:59: {{FormatCurrencyValue (AddCharges .Price)}} per kWh\n" +
	"{{end}}</code>\n{{.Goodbye}}"
