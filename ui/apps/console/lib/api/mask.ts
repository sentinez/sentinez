/** The path of an update mask that names every field */
export const MASK_ALL = '*';

/**
 * The fields of an update_mask (google.protobuf.FieldMask). Its JSON form
 * names fields in lowerCamelCase, the keys of the ts-proto messages:
 * `timeWindow`, not `time_window`.
 */
export type UpdateMask = string[];

function isSet(value: unknown): boolean {
  if (Array.isArray(value)) return value.length > 0;
  return value !== undefined && value !== null && value !== '' && value !== 0 && value !== false;
}

/**
 * stored with the fields of patch that mask names, as the API applies an
 * update (AIP-134): no mask names the fields set in patch, `*` all of
 * them; a named field unset in patch is cleared. For the sample stores.
 */
export function applyMask<T extends object>(
  stored: T,
  patch: Partial<T> | undefined,
  mask: UpdateMask | undefined,
  empty: T,
): T {
  const from: Partial<T> = patch ?? {};
  const keys = (
    !mask?.length
      ? Object.keys(from).filter((k) => isSet(from[k as keyof T]))
      : mask.includes(MASK_ALL)
        ? Object.keys(empty)
        : mask
  ) as (keyof T)[];

  const next = { ...stored };
  for (const key of keys) {
    next[key] = (isSet(from[key]) ? from[key] : empty[key]) as T[keyof T];
  }
  return next;
}
