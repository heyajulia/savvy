import { __, curry } from "ramda";

export function sortAscending(values: number[]): number[] {
  return values.sort((a, b) => a - b);
}

export function groupIntoRanges(values: number[]): (number | number[])[] {
  sortAscending(values);

  const ranges = [];

  for (let i = 0; i < values.length; i++) {
    const value = values[i];
    const nextValue = values[i + 1];

    if (nextValue === value + 1) {
      const start = values[i];

      while (true) {
        const value = values[i];
        const nextValue = values[i + 1];

        if (nextValue === value + 1) {
          i++;
        } else {
          ranges.push([start, value]);
          
          break;
        }
      }
    } else {
      ranges.push(value);
    }
  }

  return ranges;
}

function formatHour(hour: number, hourStart: boolean) {
  const h = hour.toString().padStart(2, "0");

  return hourStart ? `${h}:00` : `${h}:59`;
}

export const formatHourStart = curry(formatHour)(__, true);
export const formatHourEnd = curry(formatHour)(__, false);

export default function formatRanges(ranges: (number | number[])[]): string {
  const formatted = ranges.map((range) => {
    if (Array.isArray(range)) {
      const [start, end] = range;

      return `van ${formatHourStart(start)} tot ${formatHourEnd(end)}`;
    } else {
      return `van ${formatHourStart(range)} tot ${formatHourEnd(range)}`;
    }
  });

  return new Intl.ListFormat("nl", { style: "long", type: "conjunction" })
    .format(formatted);
}
