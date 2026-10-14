// A fetch stays loading at least this long, so skeletons do not flash
export const MIN_LOADING_MILLIS = 500;

const sleep = (millis: number) => new Promise<void>((resolve) => setTimeout(resolve, millis));

/** Resolves with `promise`, but not before MIN_LOADING_MILLIS; a failure rejects at once. */
export async function withMinLoading<T>(promise: Promise<T>): Promise<T> {
  const [res] = await Promise.all([promise, sleep(MIN_LOADING_MILLIS)]);
  return res;
}
