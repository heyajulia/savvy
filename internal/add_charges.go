package internal

// Purchase cost in Euros per kWh, including VAT.
const purchaseCost float64 = 0.0212

// Energy tax in Euros per kWh, including VAT.
const energyTax float64 = 0.15

func AddCharges(price float64) float64 {
	return price + purchaseCost + energyTax
}
