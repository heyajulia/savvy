package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"time"

	"github.com/heyajulia/energieprijzen/internal/fp"
)

type energyPrices struct {
	Prices []struct {
		Price       float64   `json:"price"`
		ReadingDate time.Time `json:"readingDate"`
	} `json:"Prices"`
	IntervalType int       `json:"intervalType"`
	Average      float64   `json:"average"`
	FromDate     time.Time `json:"fromDate"`
	TillDate     time.Time `json:"tillDate"`
}

type Price struct {
	Hour  int
	Price float64
}

type EnergyPrices struct {
	Prices       []Price
	Average      float64
	AverageHours []int
	High         float64
	HighHours    []int
	Low          float64
	LowHours     []int
}

func GetEnergyPrices(log *slog.Logger) (*EnergyPrices, error) {
	r, err := getEnergyPrices(log)
	if err != nil {
		return nil, fmt.Errorf("get energy prices: %w", err)
	}

	var prices []Price

	for i, price := range r.Prices {
		hour := i
		prices = append(prices, Price{hour, price.Price})
	}

	var e EnergyPrices

	average := r.Average

	ps := fp.Pluck[Price, float64]("Price", prices)

	low := min(ps)
	high := max(ps)

	priceIs := func(target float64) func(p Price) bool {
		return func(p Price) bool {
			return p.Price == target
		}
	}

	e.Prices = prices
	e.Average = average
	e.AverageHours = fp.Pluck[Price, int]("Hour", fp.Where(priceIs(average), prices))
	e.Low = low
	e.LowHours = fp.Pluck[Price, int]("Hour", fp.Where(priceIs(low), prices))
	e.High = high
	e.HighHours = fp.Pluck[Price, int]("Hour", fp.Where(priceIs(high), prices))

	return &e, nil
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

func min(prices []float64) float64 {
	return fp.Reduce(math.Min, math.Inf(1), prices)
}

func max(prices []float64) float64 {
	return fp.Reduce(math.Max, math.Inf(-1), prices)
}
