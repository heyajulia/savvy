import {EnergyPrices} from './types'
import formatCurrencyValue from './format-currency-value'
import getHourStartAndEnd from './get-hour-start-and-end'
import prepareQueryParameters from './prepare-query-parameters'

import {DateTime} from 'luxon'
import * as dotenv from 'dotenv'
import dedent from 'dedent'
import fetch from 'node-fetch'
import maxBy from 'lodash.maxby'
import minBy from 'lodash.minby'
import sortBy from 'lodash.sortby'

async function main() {
  dotenv.config()

  const {TOKEN: token, CHAT_ID: chat_id} = process.env

  if (!token) {
    throw new Error('No token provided')
  }

  if (!chat_id) {
    throw new Error('No chat ID provided')
  }

  const prices = await getEnergyPrices()
  const text = generateSummary(prices)

  let response

  try {
    response = await fetch(`https://api.telegram.org/bot${token}/sendMessage`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        chat_id,
        text,
        parse_mode: 'MarkdownV2',
      }),
    })
  } catch (error) {
    console.error(error)
  }

  if (response?.ok) {
    const body = await response.json()

    if (body.ok) {
      console.log('Message sent successfully!')
    } else {
      throw new Error(body)
    }
  } else {
    throw new Error(response?.statusText)
  }
}

async function getEnergyPrices(): Promise<EnergyPrices> {
  const parameters = prepareQueryParameters()
  const response = await fetch(`https://api.energyzero.nl/v1/energyprices?${parameters}`)
  const json = await response.json()

  return json as EnergyPrices
}

function generateSummary(prices: EnergyPrices): string {
  const tomorrowDate = DateTime.fromISO(prices.tillDate).toLocaleString(DateTime.DATE_FULL, {locale: 'nl-NL'})
  const average = prices.average
  const highest = maxBy(prices.Prices, 'price')!
  const lowest = minBy(prices.Prices, 'price')!

  const highestPrice = formatCurrencyValue(highest.price)
  const highestHour = DateTime.fromISO(highest.readingDate, {zone: 'Etc/UTC'}).plus({hour: 1})
  const [highestHourStart, highestHourEnd] = getHourStartAndEnd(highestHour)

  const lowestPrice = formatCurrencyValue(lowest.price)
  const lowestHour = DateTime.fromISO(lowest.readingDate, {zone: 'Etc/UTC'}).plus({hour: 1})
  const [lowestHourStart, lowestHourEnd] = getHourStartAndEnd(lowestHour)

  const allPrices = sortBy(prices.Prices, ({readingDate}) => new Date(readingDate).getTime())
    .map(({price}, index) => {
      const belowAverage = price < average ? '✅' : '❌'
      const hour = index.toString().padStart(2, '0')

      return `${belowAverage} ${hour}:00 – ${hour}:59: ${formatCurrencyValue(price)} per kWh`
    })
    .join('\n')

  const text = dedent`Goedemiddag\! ☀️ De energieprijzen van morgen ${tomorrowDate} zijn bekend\.
  
    Gemiddeld: ${formatCurrencyValue(average)} per kWh
    Hoog: ${highestPrice} per kWh tussen \(o\.a\.\) ${highestHourStart} en ${highestHourEnd}
    Laag: ${lowestPrice} per kWh tussen \(o\.a\.\) ${lowestHourStart} en ${lowestHourEnd}
    
    Alle prijzen van morgen per uur:

    \`\`\`
    ${allPrices}
    \`\`\`

    Fijne dag verder\!`

  return text
}

main()
