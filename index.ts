import { EnergyPrices } from "./types.ts";
import formatCurrencyValue from "./format-currency-value.ts";
import getHourStartAndEnd from "./get-hour-start-and-end.ts";
import prepareQueryParameters from "./prepare-query-parameters.ts";

import "dotenv/load.ts";
import { DateTime } from "luxon";
import { maxBy } from "collections/max_by.ts";
import { minBy } from "collections/min_by.ts";
import { sortBy } from "collections/sort_by.ts";
import dedent from "dedent";

async function main() {
  const token = Deno.env.get("TOKEN");
  const chat_id = Deno.env.get("CHAT_ID");

  if (!token) {
    throw new Error("No token provided");
  }

  if (!chat_id) {
    throw new Error("No chat ID provided");
  }

  const prices = await getEnergyPrices();
  const message = generateMessage(prices);

  let response;

  try {
    response = await fetch(`https://api.telegram.org/bot${token}/sendMessage`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        chat_id,
        text: message,
        parse_mode: "Markdown",
      }),
    });
  } catch (error) {
    console.error(error);
  }

  if (response?.ok) {
    const body = await response.json();

    if (body.ok) {
      console.log("Message sent successfully!");
    } else {
      throw new Error(body);
    }
  } else {
    const text = await response?.text();
    throw new Error(`Request failed: ${response?.statusText}: ${text}`);
  }
}

async function getEnergyPrices(): Promise<EnergyPrices> {
  const parameters = prepareQueryParameters();
  const response = await fetch(
    `https://api.energyzero.nl/v1/energyprices?${parameters}`,
  );
  const json = await response.json();

  return json as EnergyPrices;
}

function findHighestPrices(prices: EnergyPrices): [string, string][] {
  const maximum = maxBy(prices.Prices, ({ price }) => price)!;
  const maxima = prices.Prices.filter(({ price }) => price === maximum.price);

  return maxima.map(({ readingDate }) => {
    const hour = DateTime.fromISO(readingDate, { zone: "Etc/UTC" });
    const [start, end] = getHourStartAndEnd(hour);

    return [start, end];
  });
}

function findLowestPrices(prices: EnergyPrices): [string, string][] {
  const minimum = minBy(prices.Prices, ({ price }) => price)!;
  const minima = prices.Prices.filter(({ price }) => price === minimum.price);

  return minima.map(({ readingDate }) => {
    const hour = DateTime.fromISO(readingDate, { zone: "Etc/UTC" });
    const [start, end] = getHourStartAndEnd(hour);

    return [start, end];
  });
}

function getPriceEmoji(price: number, average: number): string {
  if (price === 0) {
    return "🆓";
  }

  if (price <= 0) {
    return "💶";
  }

  if (price < average) {
    return "✅";
  }

  return "❌";
}

function generateMessage(prices: EnergyPrices): string {
  const tomorrowDate = DateTime.fromISO(prices.tillDate).toLocaleString(
    DateTime.DATE_FULL,
    { locale: "nl-NL" },
  );
  const average = prices.average;
  const highest = maxBy(prices.Prices, ({ price }) => price)!;
  const lowest = minBy(prices.Prices, ({ price }) => price)!;

  const lf = new Intl.ListFormat("nl-NL", {
    style: "long",
    type: "conjunction",
  });

  const highestPrice = formatCurrencyValue(highest.price);
  const highestHours = lf.format(
    findHighestPrices(prices).map(([start, end]) => `van ${start} tot ${end}`),
  );

  const lowestPrice = formatCurrencyValue(lowest.price);
  const lowestHours = lf.format(
    findLowestPrices(prices).map(([start, end]) => `van ${start} tot ${end}`),
  );

  const allPrices = sortBy(
    prices.Prices,
    ({ readingDate }) => new Date(readingDate).getTime(),
  )
    .map(({ price }, index) => {
      const priceEmoji = getPriceEmoji(price, average);
      const hour = index.toString().padStart(2, "0");

      return `${priceEmoji} ${hour}:00 – ${hour}:59: ${
        formatCurrencyValue(price)
      } per kWh`;
    })
    .join("\n");

  const text =
    dedent`Goedemiddag! ☀️ De energieprijzen van morgen ${tomorrowDate} zijn bekend.
  
    Gemiddeld: ${formatCurrencyValue(average)} per kWh
    Hoog: ${highestPrice} per kWh ${highestHours}.
    Laag: ${lowestPrice} per kWh ${lowestHours}.
    
    Alle prijzen van morgen per uur:

    \`\`\`
    ${allPrices}\`\`\`

    Fijne dag verder!`;

  return text;
}

main();
