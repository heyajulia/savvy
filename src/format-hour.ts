export default function formatHour(
  hour: number,
  formatAsTextualRange = false,
): string {
  const h = hour.toString().padStart(2, "0");

  return formatAsTextualRange ? `van ${h}:00 tot ${h}:59` : `${h}:00 – ${h}:59`;
}
