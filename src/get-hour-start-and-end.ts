import { DateTime } from "luxon";

export default function getHourStartAndEnd(dt: DateTime): [string, string] {
  const start = dt.toFormat("HH:mm");
  const end = dt.plus({ minutes: 59 }).toFormat("HH:mm");

  return [start, end];
}
