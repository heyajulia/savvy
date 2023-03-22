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
  const text = generateSummary(prices);

  let response;

  try {
    response = await fetch(`https://api.telegram.org/bot${token}/sendMessage`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        chat_id,
        text,
        parse_mode: "MarkdownV2",
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
    throw new Error(response?.statusText);
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

function generateSummary(prices: EnergyPrices): string {
  const tomorrowDate = DateTime.fromISO(prices.tillDate).toLocaleString(
    DateTime.DATE_FULL,
    { locale: "nl-NL" },
  );
  const average = prices.average;
  const highest = maxBy(prices.Prices, ({ price }) => price)!;
  const lowest = minBy(prices.Prices, ({ price }) => price)!;

  const highestPrice = formatCurrencyValue(highest.price);
  const highestHour = DateTime.fromISO(highest.readingDate, { zone: "Etc/UTC" })
    .plus({ hour: 1 });
  const [highestHourStart, highestHourEnd] = getHourStartAndEnd(highestHour);

  const lowestPrice = formatCurrencyValue(lowest.price);
  const lowestHour = DateTime.fromISO(lowest.readingDate, { zone: "Etc/UTC" })
    .plus({ hour: 1 });
  const [lowestHourStart, lowestHourEnd] = getHourStartAndEnd(lowestHour);

  const allPrices = sortBy(
    prices.Prices,
    ({ readingDate }) => new Date(readingDate).getTime(),
  )
    .map(({ price }, index) => {
      const belowAverage = price < average ? "✅" : "❌";
      const hour = index.toString().padStart(2, "0");

      return `${belowAverage} ${hour}:00 – ${hour}:59: ${
        formatCurrencyValue(price)
      } per kWh`;
    })
    .join("\n");

  const text =
    dedent`Goedemiddag\! ☀️ De energieprijzen van morgen ${tomorrowDate} zijn bekend\.
  
    Gemiddeld: ${formatCurrencyValue(average)} per kWh
    Hoog: ${highestPrice} per kWh tussen \(o\.a\.\) ${highestHourStart} en ${highestHourEnd}
    Laag: ${lowestPrice} per kWh tussen \(o\.a\.\) ${lowestHourStart} en ${lowestHourEnd}
    
    Alle prijzen van morgen per uur:

    \`\`\`
    ${allPrices}\`\`\`

    Fijne dag verder\!`;

  return text;
}

main();
