package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/heyajulia/energieprijzen/internal/prices"
)

var ErrPriceLength = errors.New("unexpected number of prices")

func GetEnergyPrices() (*prices.Prices, error) {
	u, err := url.Parse("https://api.energyzero.nl/v1/energyprices")
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}
	u.RawQuery = QueryParameters(time.Now()).Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var e struct {
		Prices []struct {
			Price float64 `json:"price"`
		} `json:"Prices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&e); err != nil {
		return nil, fmt.Errorf("decoding response body: %w", err)
	}

	if len(e.Prices) != 24 {
		return nil, ErrPriceLength
	}

	ps := make([]float64, len(e.Prices))

	for i, price := range e.Prices {
		ps[i] = price.Price
	}

	return prices.New(ps), nil
}
