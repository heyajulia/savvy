package datetime

import (
	"strings"
	"time"
)

var amsterdam *time.Location

func init() {
	loc, err := time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		panic(err)
	}

	amsterdam = loc
}

var replacer = strings.NewReplacer(
	"Monday", "maandag",
	"Tuesday", "dinsdag",
	"Wednesday", "woensdag",
	"Thursday", "donderdag",
	"Friday", "vrijdag",
	"Saturday", "zaterdag",
	"Sunday", "zondag",
	"January", "januari",
	"February", "februari",
	"March", "maart",
	"April", "april",
	"May", "mei",
	"June", "juni",
	"July", "juli",
	"August", "augustus",
	"September", "september",
	"October", "oktober",
	"November", "november",
	"December", "december",
)

func Now() time.Time {
	return time.Now().In(amsterdam)
}

func Tomorrow(t time.Time) time.Time {
	return t.AddDate(0, 0, 1)
}

func Format(t time.Time) string {
	const layout = "Monday 2 January 2006"

	return replacer.Replace(t.Format(layout))
}

func FormatRFC3339Milli(t time.Time) string {
	// See https://go.dev/issue/36472 and issue #75 in this repo.
	return t.Round(time.Millisecond).Format(time.RFC3339Nano)
}
