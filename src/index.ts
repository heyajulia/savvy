import { Prices, ZodApiResponse } from "./types.ts";
import addCharges from "./add-charges.ts";
import formatCurrencyValue from "./format-currency-value.ts";
import formatHour from "./format-hour.ts";
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

  const { prices, average, date } = await getEnergyPrices();
  const [, lowestPrice] = minBy(prices, ([, price]) => price)!;
  const [, highestPrice] = maxBy(prices, ([, price]) => price)!;
  const lowestPriceHours = prices.filter(([, price]) => price === lowestPrice)
    .map(([hour]) => formatHour(hour, true));
  const highestPriceHours = prices.filter(([, price]) => price === highestPrice)
    .map(([hour]) => formatHour(hour, true));

  const listFormatter = new Intl.ListFormat("nl", {
    style: "long",
    type: "conjunction",
  });

  const lowestHours = listFormatter.format(lowestPriceHours);
  const highestHours = listFormatter.format(highestPriceHours);

  const allPrices = prices.map(([hour, price]) =>
    `${getPriceEmoji(price, average)} ${formatHour(hour, false)}: ${
      formatCurrencyValue(price)
    } per kWh`
  ).join("\n");

  const message =
    dedent`Goedemiddag! ☀️ De energieprijzen van ${date} zijn bekend.
  
    Gemiddeld: ${formatCurrencyValue(average)} per kWh
    Hoog: ${formatCurrencyValue(highestPrice)} per kWh ${highestHours}.
    Laag: ${formatCurrencyValue(lowestPrice)} per kWh ${lowestHours}.
    
    Alle prijzen van morgen per uur:

    \`\`\`
    ${allPrices}\`\`\`

    Fijne dag verder!`;

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
