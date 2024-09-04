package prices

import (
	"iter"
	"math"
	"slices"
)

// Prices represents a collection of energy prices.
//
// Though we understand there to be 24 prices (one for each hour of the day) in practice, Prices doesn't enforce this.
// Instead, it is up to the caller to ensure that the number of prices is correct. This might seem odd, but it makes
// Prices somewhat more flexible, and it makes testing easier.
type Prices struct {
	prices                            []float64
	average, high, low                float64
	averageHours, highHours, lowHours []int
}

func New(prices []float64) *Prices {
	for i, p := range prices {
		// TODO: Perhaps we could let the caller decide whether to round or add charges to make testing easier.
		prices[i] = round(addCharges(p))
	}

	p := new(Prices)
	p.prices = prices
	p.calculate()

	return p
}

func (p *Prices) calculate() {
	p.average = calculateAverage(p.prices)
	p.high = round(slices.Max(p.prices))
	p.low = round(slices.Min(p.prices))

	p.averageHours = wherePriceIs(p.average, p.prices)
	p.highHours = wherePriceIs(p.high, p.prices)
	p.lowHours = wherePriceIs(p.low, p.prices)
}

func (p *Prices) All() iter.Seq2[int, float64] {
	return slices.All(p.prices)
}

func (p *Prices) Average() float64 {
	return p.average
}

func (p *Prices) AverageHours() []int {
	return p.averageHours
}

func (p *Prices) High() float64 {
	return p.high
}

func (p *Prices) HighHours() []int {
	return p.highHours
}

func (p *Prices) Low() float64 {
	return p.low
}

func (p *Prices) LowHours() []int {
	return p.lowHours
}

func calculateAverage(prices []float64) float64 {
	sum := 0.0

	for _, v := range prices {
		sum += v
	}

	return round(sum / float64(len(prices)))
}

func wherePriceIs(price float64, prices []float64) []int {
	var hours []int

	for hour, p := range prices {
		if p == price {
			hours = append(hours, hour)
		}
	}

	return hours
}

func addCharges(price float64) float64 {
	const (
		purchaseCost = 0.0484
		energyTax    = 0.1312
	)

	return price + purchaseCost + energyTax
}

func round(price float64) float64 {
	return math.Round(price*100) / 100
}
