import { EnergyPrices, Price } from "./types.ts";
import getHourStartAndEnd from "./get-hour-start-and-end.ts";

import { DateTime } from "luxon";
import { maxBy } from "collections/max_by.ts";
import { minBy } from "collections/min_by.ts";

export function findPrices(
  prices: EnergyPrices,
  fn: (
    array: readonly Price[],
    selector: (el: Price) => number,
  ) => Price | undefined,
): [string, string][] {
  const target = fn(prices.Prices, ({ price }) => price)!;
  const targets = prices.Prices.filter(({ price }) => price === target.price);

  return targets.map(({ readingDate }) => {
    const hour = DateTime.fromISO(readingDate, { zone: "Etc/UTC" });

    return getHourStartAndEnd(hour);
  });
}

export function findHighestPrices(prices: EnergyPrices): [string, string][] {
  return findPrices(prices, maxBy);
}

export function findLowestPrices(prices: EnergyPrices): [string, string][] {
  return findPrices(prices, minBy);
}
