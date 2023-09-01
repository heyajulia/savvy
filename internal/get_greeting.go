package internal

import "time"

func GetGreeting() (hello string, goodbye string) {
	amsterdam, err := time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		panic(err)
	}

	now := time.Now().In(amsterdam)
	hour := now.Hour()

	if hour < 18 {
		hello = "Goedemiddag! ☀️"
		goodbye = "Fijne dag verder!"
	} else {
		hello = "Goedenavond! 🌙"
		goodbye = "Geniet van je avond!"
	}

	return
}
