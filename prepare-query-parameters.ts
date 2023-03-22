import { DateTime } from "luxon";
import { URLSearchParams } from "node:url";

export default function prepareQueryParameters(): string {
  const fromDate = DateTime.utc().set({
    hour: 23,
    minute: 0,
    second: 0,
    millisecond: 0,
  });
  const tillDate = fromDate.plus({ days: 1 }).minus({ milliseconds: 1 });

  const params = new URLSearchParams({
    fromDate: fromDate.toISO(),
    tillDate: tillDate.toISO(),
    interval: "4",
    usageType: "1",
    inclBtw: "true",
  });

  return params.toString();
}
