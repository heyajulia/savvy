import { DateTime } from "luxon";

export default function prepareQueryParameters(): string {
  const fromDate = DateTime.utc().set({
    hour: 22,
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
