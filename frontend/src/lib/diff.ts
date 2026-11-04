/** Line-level comparison used to show what changed between two fetches of a page. Order is ignored, as on the server. */
export function diffLines(before: string, after: string): { added: string[]; removed: string[] } {
  const lines = (text: string) => text.split("\n").map(line => line.trim()).filter(Boolean);
  const remaining = new Map<string, number>();
  for (const line of lines(before)) remaining.set(line, (remaining.get(line) ?? 0) + 1);
  const added: string[] = [];
  for (const line of lines(after)) {
    const count = remaining.get(line) ?? 0;
    if (count > 0) remaining.set(line, count - 1);
    else added.push(line);
  }
  const removed: string[] = [];
  for (const [line, count] of remaining) for (let i = 0; i < count; i++) removed.push(line);
  return { added, removed };
}
