package internal

import "time"

func GetGreeting(t time.Time) (hello, goodbye string) {
	hour := t.Hour()

	if hour < 18 {
		hello = "Goedemiddag! ☀️"
		goodbye = "Fijne dag verder!"
	} else {
		hello = "Goedenavond! 🌙"
		goodbye = "Geniet van je avond!"
	}

	return
}
