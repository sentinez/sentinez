/** A line of a diff: kept, only in the old text, or only in the new one */
export type DiffLine = { kind: 'same' | 'removed' | 'added'; text: string };

/** Unchanged lines collapsed into one marker, as a DiffLine is shown */
export type DiffHunk = DiffLine | { kind: 'skipped'; count: number };

// The size of the table of the longest common subsequence past which the
// texts are shown as removed then added, not compared line by line
const MAX_CELLS = 1_000_000;

/**
 * The lines of before and after, along their longest common subsequence,
 * like `git diff`: the common head and tail as is, the dynamic programming
 * only on what lies between.
 */
export function lineDiff(before: string[], after: string[]): DiffLine[] {
  let head = 0;
  while (head < before.length && head < after.length && before[head] === after[head]) head++;
  let tail = 0;
  while (
    tail < before.length - head &&
    tail < after.length - head &&
    before[before.length - 1 - tail] === after[after.length - 1 - tail]
  )
    tail++;

  const same = (text: string): DiffLine => ({ kind: 'same', text });
  return [
    ...before.slice(0, head).map(same),
    ...middleDiff(before.slice(head, before.length - tail), after.slice(head, after.length - tail)),
    ...before.slice(before.length - tail).map(same),
  ];
}

// The diff of a and b, which differ at both ends
function middleDiff(a: string[], b: string[]): DiffLine[] {
  const removed = (text: string): DiffLine => ({ kind: 'removed', text });
  const added = (text: string): DiffLine => ({ kind: 'added', text });
  if ((a.length + 1) * (b.length + 1) > MAX_CELLS) {
    return [...a.map(removed), ...b.map(added)];
  }

  // n[i * w + j] is the length of the lcs of a[i:] and b[j:]
  const w = b.length + 1;
  const n = new Uint32Array((a.length + 1) * w);
  for (let i = a.length - 1; i >= 0; i--) {
    for (let j = b.length - 1; j >= 0; j--) {
      n[i * w + j] =
        a[i] === b[j]
          ? n[(i + 1) * w + j + 1]! + 1
          : Math.max(n[(i + 1) * w + j]!, n[i * w + j + 1]!);
    }
  }

  const lines: DiffLine[] = [];
  let i = 0;
  let j = 0;
  while (i < a.length && j < b.length) {
    if (a[i] === b[j]) {
      lines.push({ kind: 'same', text: a[i]! });
      i++;
      j++;
    } else if (n[(i + 1) * w + j]! >= n[i * w + j + 1]!) {
      lines.push(removed(a[i++]!));
    } else {
      lines.push(added(b[j++]!));
    }
  }
  return [...lines, ...a.slice(i).map(removed), ...b.slice(j).map(added)];
}

/** lines with each run of unchanged ones longer than 2 × context collapsed, context kept around the changes */
export function collapse(lines: DiffLine[], context = 2): DiffHunk[] {
  const hunks: DiffHunk[] = [];
  let i = 0;
  while (i < lines.length) {
    if (lines[i]!.kind !== 'same') {
      hunks.push(lines[i++]!);
      continue;
    }
    let end = i;
    while (end < lines.length && lines[end]!.kind === 'same') end++;
    // a run at the start or the end keeps context on one side only
    const keepHead = i === 0 ? 0 : context;
    const keepTail = end === lines.length ? 0 : context;
    const count = end - i - keepHead - keepTail;
    if (count > 0) {
      hunks.push(...lines.slice(i, i + keepHead));
      hunks.push({ kind: 'skipped', count });
      hunks.push(...lines.slice(end - keepTail, end));
    } else {
      hunks.push(...lines.slice(i, end));
    }
    i = end;
  }
  return hunks;
}
