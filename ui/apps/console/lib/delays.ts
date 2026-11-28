// How long the spinner covers the page when switching sidebar tabs
// (components/tab-loading.tsx). The page fetches underneath meanwhile.
// Intentional: never remove it.
export const TAB_LOADING_MILLIS = 400;

/** Resolves after millis, or as soon as signal (when given) aborts */
export function sleep(millis: number, signal?: AbortSignal): Promise<void> {
  return new Promise((resolve) => {
    const timer = setTimeout(resolve, millis);
    signal?.addEventListener(
      'abort',
      () => {
        clearTimeout(timer);
        resolve();
      },
      { once: true },
    );
  });
}

// Resolves with `work`, but not before `millis`: a slow `work` makes the wait
// longer, a fast one still waits the full time. A failure rejects at once.
async function atLeast<T>(millis: number, work: Promise<T>): Promise<T> {
  const [res] = await Promise.all([work, sleep(millis)]);
  return res;
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
  return atLeast(MIN_LOADING_MILLIS, work);
}
