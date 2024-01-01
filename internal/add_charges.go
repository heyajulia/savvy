package internal

// Returns the price in Euros per kWh, including VAT.
func AddCharges(price float64) float64 {
	const (
		purchaseCost = 0.0484
		energyTax = 0.1312
	)

	return price + purchaseCost + energyTax
}
