import { Prices, ZodApiResponse } from "./types.ts";
import addCharges from "./add-charges.ts";
import prepareQueryParameters from "./prepare-query-parameters.ts";

import "dotenv/load.ts";
import { DateTime } from "luxon";

async function main() {
  const prices = await getEnergyPrices();

  console.log(prices);
}

async function getEnergyPrices(): Promise<Prices> {
  const parameters = prepareQueryParameters();
  const response = await fetch(
    `https://api.energyzero.nl/v1/energyprices?${parameters}`,
  );
  const { Prices: prices, average, tillDate } = ZodApiResponse.parse(
    await response.json(),
  );

  return {
    prices: prices.map(({ price }) => addCharges(price)),
    average: addCharges(average),
    date: DateTime.fromISO(tillDate, { locale: "nl" }).toLocaleString(
      DateTime.DATE_HUGE,
    ),
  };
}

main();
