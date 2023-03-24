// Purchase cost in Euros per kWh, including VAT.
const purchaseCost = 0.0212;

// Energy tax in Euros per kWh, including VAT.
const energyTax = 0.15;

export default function addCharges(price: number): number {
  return price + purchaseCost + energyTax;
}
