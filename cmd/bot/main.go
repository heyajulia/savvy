package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"strings"
	_ "time/tzdata"

	"github.com/heyajulia/energieprijzen/internal"
	"github.com/heyajulia/energieprijzen/internal/date"
)

type credentials struct {
	Telegram string `json:"telegram"`
}

var (
	dryRun        bool
	token, chatID string
)

func init() {
	flag.BoolVar(&dryRun, "d", false, "dry run")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)

	// No need for credentials or a chat ID if we're not sending a message.
	if dryRun {
		return
	}

	creds := readCredentials()

	token = creds.Telegram

	if id, ok := os.LookupEnv("ENERGIEPRIJZEN_BOT_CHAT_ID"); ok {
		chatID = id
	} else {
		log.Fatalln("ENERGIEPRIJZEN_BOT_CHAT_ID is not set")
	}
}

func readCredentials() credentials {
	p := os.ExpandEnv("$CREDENTIALS_DIRECTORY/token")
	if _, err := os.Stat(p); errors.Is(err, fs.ErrNotExist) {
		log.Fatalln("Credentials file does not exist. Are you running this using systemd?")
	}

	f, err := os.Open(p)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	var creds credentials

	err = json.NewDecoder(f).Decode(&creds)
	if err != nil {
		log.Fatalln(err)
	}

	return creds
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

	log.Println("Getting greeting...")

	hello, goodbye := internal.GetGreeting()

	log.Println("Executing template...")

	data := TemplateData{
		Hello:        hello,
		Goodbye:      goodbye,
		TomorrowDate: date.Tomorrow(),
		EnergyPrices: prices,
	}

	err = report(data).Render(context.Background(), &sb)
	if err != nil {
		log.Fatalln(err)
	}

	message := strings.ReplaceAll(sb.String(), "<br>", "\n")

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
