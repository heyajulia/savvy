package internal

func GetGreeting() (hello string, goodbye string) {
	hour := AmsterdamTime().Hour()

	if hour < 18 {
		hello = "Goedemiddag! ☀️"
		goodbye = "Fijne dag verder!"
	} else {
		hello = "Goedenavond! 🌙"
		goodbye = "Geniet van je avond!"
	}

	return
}
