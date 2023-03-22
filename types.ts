export interface EnergyPrices {
  Prices: Price[]
  intervalType: number
  average: number
  fromDate: string
  tillDate: string
}

export interface Price {
  price: number
  readingDate: string
}
