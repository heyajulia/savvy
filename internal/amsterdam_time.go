package internal

import "time"

func AmsterdamTime() time.Time {
	amsterdam, err := time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		panic(err)
	}

	return time.Now().In(amsterdam)
}
