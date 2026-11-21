// How long a sidebar-tab page waits before it renders, so its loading.tsx
// shows when switching tabs. Intentional: never remove it.
export const PAGE_DELAY_MILLIS = 400;

const sleep = (millis: number) => new Promise<void>((resolve) => setTimeout(resolve, millis));

// Resolves with `work`, but not before `millis`: a slow `work` makes the wait
// longer, a fast one still waits the full time. A failure rejects at once.
async function atLeast<T>(millis: number, work?: Promise<T>): Promise<T | undefined> {
  const [res] = await Promise.all([work, sleep(millis)]);
  return res;
}

/**
 * Waits PAGE_DELAY_MILLIS, or longer if `work` is slower (see atLeast).
 * Without `work` it is just the delay; `await` it first in a sidebar-tab page.
 */
export function pageDelay(): Promise<void>;
export function pageDelay<T>(work: Promise<T>): Promise<T>;
export function pageDelay<T>(work?: Promise<T>) {
  return atLeast(PAGE_DELAY_MILLIS, work);
}

// How long the splash (logo and progress bar) stays when the app loads.
export const SPLASH_MILLIS = 900;
// On a holiday it stays a bit longer: the greeting fades in by 0.85s, then is
// left long enough to read.
export const SPLASH_HOLIDAY_MILLIS = 1400;

export const splashMillis = (isHoliday: boolean) =>
  isHoliday ? SPLASH_HOLIDAY_MILLIS : SPLASH_MILLIS;

// A data fetch stays loading at least this long, so skeletons do not flash.
// The refresh button keeps its icon spinning as long, so both end together.
export const MIN_LOADING_MILLIS = 300;

/** Resolves with `work`, but not before MIN_LOADING_MILLIS (see atLeast). */
export function withMinLoading<T>(work: Promise<T>): Promise<T> {
  return atLeast(MIN_LOADING_MILLIS, work) as Promise<T>;
}
