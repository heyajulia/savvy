import * as R from "ramda";

export function groupIntoRanges(values: number[]): (number | number[])[] {
  const pipe = R.pipe(
    R.sort(R.ascend<number>(R.identity)),
    R.groupWith((a, b) => R.equals(R.inc(a), b)),
    R.map((group) => {
      const first = R.head(group)!;

      return R.ifElse(
        R.pipe(R.length, R.equals(1)),
        R.always(first),
        R.pipe(R.last, R.append(R.__, [first])),
      )(group);
    }),
  );

  return pipe(values);
}

function formatHour(hour: number, hourStart: boolean) {
  const h = hour.toString().padStart(2, "0");

  return hourStart ? `${h}:00` : `${h}:59`;
}

export const formatHourStart = R.curry(formatHour)(R.__, true);
export const formatHourEnd = R.curry(formatHour)(R.__, false);

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
