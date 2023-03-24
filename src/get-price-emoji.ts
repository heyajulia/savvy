export default function getPriceEmoji(price: number, average: number): string {
  return price < average ? "✅" : "❌";
}
