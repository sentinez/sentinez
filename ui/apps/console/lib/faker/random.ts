// A random number generator of [0, 1) seeded by seed (mulberry32 over an
// FNV-1a hash): the same seed gives the same numbers, so samples asked twice
// (the histogram, then the visitors) agree and do not change on re-render
export function sampleRandom(seed: string): () => number {
  let h = 0x811c9dc5;
  for (let i = 0; i < seed.length; i++) {
    h = Math.imul(h ^ seed.charCodeAt(i), 0x01000193);
  }
  return () => {
    h = (h + 0x6d2b79f5) | 0;
    let t = Math.imul(h ^ (h >>> 15), 1 | h);
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

// The daily rhythm of the samples at time (millis): -1 at 00:00 UTC, 1 at
// 12:00
export const sampleDayCurve = (time: number) =>
  Math.sin(((new Date(time).getUTCHours() - 6) / 24) * 2 * Math.PI);

// A random element of list, drawn with random
export const sampleOneOf = <T>(random: () => number, list: readonly T[]): T =>
  list[Math.floor(random() * list.length)]!;

// SampleLevel the shape of the traffic of a sample resource: its share of a
// quiet resource (the rest scaled by the resource), the mean, the swing of
// the day and the noise around it, and the factor of a burst (1 bucket in 12)
export type SampleLevel = {
  quiet: number;
  mean: number;
  day: number;
  noise: number;
  burst: number;
};

// The traffic level of a resource in the bucket at time, random but the same
// for the same resource and bucket: a busy or quiet resource (its own
// scale), busier by day, with noise and a burst now and then
export function sampleLevel(resourceId: string, time: Date, l: SampleLevel): number {
  const scale = sampleRandom(resourceId)();
  const random = sampleRandom(`${resourceId}|${time.getTime()}`);
  const burst = random() < 0.08 ? l.burst : 1;
  const day = sampleDayCurve(time.getTime());
  return (l.quiet + (1 - l.quiet) * scale) * (l.mean + l.day * day + l.noise * random()) * burst;
}
