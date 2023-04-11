import { DateTime } from "luxon";

export default function getGreeting(): [string, string] {
  const { hour } = DateTime.local({ zone: "Europe/Amsterdam" });

  return hour < 18
    ? ["Goedemiddag! ☀️", "Fijne dag verder!"]
    : ["Goedenavond! 🌙", "Geniet van je avond!"];
}
