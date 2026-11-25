// Counts in the short form of the charts and their tables: 999, 1K, 1.2K,
// 3.4M, 1B. Always the K / M / B suffixes (the en form), whatever the locale:
// vi would write "1,2 N" and "3,4 Tr".
const _compact = new Intl.NumberFormat('en', {
  notation: 'compact',
  maximumFractionDigits: 1,
});

/** A count in its short form, e.g. 1234 as "1.2K" */
export function compactNumber(n: number): string {
  return _compact.format(n);
}
