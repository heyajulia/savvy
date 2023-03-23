export default function getPriceEmoji(price: number, average: number) {
  if (price === 0) {
    return "🆓";
  }

  if (price < 0) {
    return "💶";
  }

  return price < average ? "✅" : "❌";
}
