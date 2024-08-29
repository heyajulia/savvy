package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"math"
	"net/http"
	"net/url"
	"slices"
)

var ErrNoPrices = errors.New("prices not yet available")

type energyPrices struct {
	Prices []struct {
		Price float64 `json:"price"`
	} `json:"Prices"`
}

type Price struct {
	Hour  int
	Price float64
}

type EnergyPrices struct {
	prices map[int]float64
}

// Returns the price in Euros per kWh, including VAT.
func AddCharges(price float64) float64 {
	const (
		purchaseCost = 0.0484
		energyTax    = 0.1312
	)

	return price + purchaseCost + energyTax
}

// Yeah, this is pretty weird. Bear with me.

func round(f float64) float64 {
	return math.Round(f*100) / 100
}

func NewEnergyPrices(prices []Price) *EnergyPrices {
	m := make(map[int]float64, len(prices))

	for _, p := range prices {
		m[p.Hour] = round(AddCharges(p.Price))
	}

	return &EnergyPrices{m}
}

func (e *EnergyPrices) Get(hour int) (float64, bool) {
	price, ok := e.prices[hour]
	return price, ok
}

func (e *EnergyPrices) Average() float64 {
	sum := 0.0

	for v := range maps.Values(e.prices) {
		sum += v
	}

	return round(sum / float64(len(e.prices)))
}

func (e *EnergyPrices) AverageHours() []int {
	return e.wherePriceIs(e.Average())
}

func (e *EnergyPrices) High() float64 {
	return round(slices.Max(slices.Collect(maps.Values(e.prices))))
}

func (e *EnergyPrices) HighHours() []int {
	return e.wherePriceIs(e.High())
}

func (e *EnergyPrices) Low() float64 {
	return round(slices.Min(slices.Collect(maps.Values(e.prices))))
}

func (e *EnergyPrices) LowHours() []int {
	return e.wherePriceIs(e.Low())
}

func (e *EnergyPrices) wherePriceIs(price float64) []int {
	hours := []int{}

	for hour, p := range e.prices {
		if p == price {
			hours = append(hours, hour)
		}
	}

	return hours
}

func GetEnergyPrices(log *slog.Logger) (*EnergyPrices, error) {
	// TODO: Clean this whole thing up.

	r, err := getEnergyPrices(log)
	if err != nil {
		return nil, fmt.Errorf("get energy prices: %w", err)
	}

	if len(r.Prices) == 0 {
		return nil, ErrNoPrices
	}

	var prices []Price

	for i, price := range r.Prices {
		hour := i
		prices = append(prices, Price{hour, price.Price})
	}

	return NewEnergyPrices(prices), nil
}

func getEnergyPrices(log *slog.Logger) (*energyPrices, error) {
	baseURL := "https://api.energyzero.nl/v1/energyprices"
	queryParams := PrepareQueryParameters()

	requestURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}
	requestURL.RawQuery = queryParams

	response, err := http.Get(requestURL.String())
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	log.Info("got energy prices", slog.Group("response", slog.Int("status_code", response.StatusCode), slog.String("body", string(body))))

	var e energyPrices

	err = json.Unmarshal(body, &e)
	if err != nil {
		return nil, fmt.Errorf("decoding response body: %w", err)
	}

	return &e, nil
}
