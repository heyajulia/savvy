package internal

import (
	"fmt"
	"strings"
)

func FormatCurrencyValue(value float64) string {
	return strings.Replace(fmt.Sprintf("€\u00a0%.2f", value), ".", ",", 1)
}
