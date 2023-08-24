package internal

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"time"

	"github.com/heyajulia/energieprijzen/internal/fp"
)

var ErrStatus = errors.New("status code is not 200")

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

func GetEnergyPrices() (*EnergyPrices, error) {
	r, err := getEnergyPrices()
	if err != nil {
		return nil, err
	}

	var prices []Price

	for i, price := range r.Prices {
		hour := i
		prices = append(prices, Price{hour, price.Price})
	}

	var e EnergyPrices

	average := r.Average

	ps := fp.Pluck[Price, float64]("Price", prices)

	// The brand-new "min" and "max" built-ins don't seem to work with floats.
	//
	// The docs say: "If T is a floating-point type and any of the arguments are NaNs, min will return NaN," but trying
	// to use "min" or "max" with floating-point numbers simply fails to compile[0] with the following error:
	//
	// "invalid argument: []float32{…} (value of type []float32) cannot be ordered"
	//
	// Maybe I'm missing something, but thankfully, it's easy enough to roll our own version that is blissfully unaware
	// of the subtleties of floating-point arithmetic. 🤠
	//
	// [0]: https://go.dev/play/p/O0WiBgCEzK5
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

func getEnergyPrices() (*energyPrices, error) {
	baseURL := "https://api.energyzero.nl/v1/energyprices"
	queryParams := PrepareQueryParameters()

	requestURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	requestURL.RawQuery = queryParams

	request, err := http.NewRequest("GET", requestURL.String(), nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, ErrStatus
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	log.Printf("status code %d, body: %#v\n", response.StatusCode, string(body))

	var e energyPrices

	err = json.Unmarshal(body, &e)
	if err != nil {
		return nil, err
	}

	return &e, nil
}

func min(prices []float64) float64 {
	return fp.Reduce(math.Min, math.Inf(1), prices)
}

func max(prices []float64) float64 {
	return fp.Reduce(math.Max, math.Inf(-1), prices)
}
