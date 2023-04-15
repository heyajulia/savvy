import { Prices, ZodApiResponse } from "./types.ts";
import addCharges from "./add-charges.ts";
import formatCurrencyValue from "./format-currency-value.ts";
import formatRanges, {
  formatHourEnd,
  formatHourStart,
  groupIntoRanges,
} from "./pretty-range.ts";
import getGreeting from "./get-greeting.ts";
import getPriceEmoji from "./get-price-emoji.ts";
import prepareQueryParameters from "./prepare-query-parameters.ts";

import "dotenv/load.ts";
import { DateTime } from "luxon";
import { maxBy } from "collections/max_by.ts";
import { minBy } from "collections/min_by.ts";
import dedent from "dedent";

async function main() {
  const chatId = Deno.env.get("CHAT_ID");
  const token = Deno.env.get("TOKEN");

  if (!chatId || !token) {
    throw new Error("Missing environment variables");
  }

  const [greeting, farewell] = getGreeting();

  const { prices, average, date } = await getEnergyPrices();
  const [, lowestPrice] = minBy(prices, ([, price]) => price)!;
  const [, highestPrice] = maxBy(prices, ([, price]) => price)!;

  const lowestPriceHours = prices.filter(([, price]) => price === lowestPrice)
    .map(([hour]) => hour);
  const highestPriceHours = prices.filter(([, price]) => price === highestPrice)
    .map(([hour]) => hour);

  const lowestPriceRanges = formatRanges(groupIntoRanges(lowestPriceHours));
  const highestPriceRanges = formatRanges(groupIntoRanges(highestPriceHours));

  const allPrices = prices.map(([hour, price]) =>
    `${getPriceEmoji(price, average)} ${formatHourStart(hour)} – ${
      formatHourEnd(hour)
    }: ${formatCurrencyValue(price)} per kWh`
  ).join("\n");

  const message = dedent`${greeting} De energieprijzen van ${date} zijn bekend.
  
    Gemiddeld: ${formatCurrencyValue(average)} per kWh
    Hoog: ${formatCurrencyValue(highestPrice)} per kWh ${highestPriceRanges}.
    Laag: ${formatCurrencyValue(lowestPrice)} per kWh ${lowestPriceRanges}.
    
    Alle prijzen van morgen per uur:

    \`\`\`
    ${allPrices}\`\`\`

    ${farewell}`;

  await fetch(`https://api.telegram.org/bot${token}/sendMessage`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      chat_id: chatId,
      text: message,
      parse_mode: "Markdown",
    }),
  });
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
    prices: prices.map(({ price }, hour) => [hour, addCharges(price)]),
    average: addCharges(average),
    date: DateTime.fromISO(tillDate, { locale: "nl" }).toLocaleString(
      DateTime.DATE_HUGE,
    ),
  };
}

main();
