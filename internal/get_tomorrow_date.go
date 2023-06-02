package internal

import (
	"strings"
	"time"
)

func GetTomorrowDate() string {
	t := time.Now()
	tomorrow := t.AddDate(0, 0, 1)
	replacer := strings.NewReplacer(
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

	return replacer.Replace(tomorrow.Format("Monday 2 January 2006"))
}
