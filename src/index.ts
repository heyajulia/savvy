import { EnergyPrices } from "./types.ts";
import addCharges from "./add-charges.ts";
import prepareQueryParameters from "./prepare-query-parameters.ts";

import "dotenv/load.ts";
import { DateTime } from "https://cdn.skypack.dev/luxon@3.3.0?dts";

async function main() {
  const prices = await getEnergyPrices();

  console.log(prices);
}

async function getEnergyPrices(): Promise<MyPrices> {
  const parameters = prepareQueryParameters();
  const response = await fetch(
    `https://api.energyzero.nl/v1/energyprices?${parameters}`,
  );
  const prices = await response.json() as EnergyPrices;

  prices.average = addCharges(prices.average);

  for (const price of prices.Prices) {
    price.price = addCharges(price.price);
  }

  const myPrices = {} as MyPrices;

  myPrices.prices = prices.Prices.map(({ price }) => price);
  myPrices.average = prices.average;
  myPrices.date = DateTime.fromISO(prices.tillDate, { locale: "nl" })
    .toLocaleString(DateTime.DATE_HUGE);

  return myPrices;
}

interface MyPrices {
  prices: number[];
  average: number;
  date: string;
}

main();
