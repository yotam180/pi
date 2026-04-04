const input = require("fs").readFileSync("/dev/stdin", "utf-8");
const items: { name: string; score: number }[] = JSON.parse(input);

const sorted = items.sort((a, b) => b.score - a.score);
console.log("=== Leaderboard ===");
sorted.forEach((item, i) => {
  const medal = i === 0 ? "* " : "  ";
  console.log(`${medal}${i + 1}. ${item.name.padEnd(10)} ${item.score}`);
});
console.log(`=== ${sorted.length} entries ===`);
