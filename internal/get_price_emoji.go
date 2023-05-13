package internal

func GetPriceEmoji(price float64, average float64) string {
	if price == 0 {
		return "🆓"
	}

	if price < 0 {
		return "💶"
	}

	if price <= average {
		return "✅"
	}

	return "❌"
}
