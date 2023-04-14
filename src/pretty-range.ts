export function sortAscending(values: number[]): number[] {
  return values.sort((a, b) => a - b);
}

export function groupIntoRanges(values: number[]): (number | number[])[] {
  const ranges = [];

  for (let i = 0; i < values.length; i++) {
    if (values[i + 1] === values[i] + 1) {
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
      ranges.push(values[i]);
    }
  }

  return ranges;
}

export function formatHour(hour: number, hourStart = true) {
  const h = hour.toString().padStart(2, "0");

  return hourStart ? `${h}:00` : `${h}:59`;
}

export default function formatRanges(ranges: (number | number[])[]): string {
  const formatted = ranges.map((range) => {
    if (Array.isArray(range)) {
      return `van ${formatHour(range[0])} tot ${formatHour(range[1], false)}`;
    } else {
      return `om ${formatHour(range)}`;
    }
  });

  return new Intl.ListFormat("nl", { style: "long", type: "conjunction" })
    .format(formatted);
}
