package date

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

func Amsterdam() time.Time {
	return time.Now().In(amsterdam)
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

func Tomorrow() string {
	today := Amsterdam()
	tomorrow := today.AddDate(0, 0, 1)

	return replacer.Replace(tomorrow.Format("Monday 2 January 2006"))
}
