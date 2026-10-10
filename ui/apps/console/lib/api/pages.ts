import type { Pages } from '@sentinez/proto/sentinez/types/v1/model';

export type { Pages };

/**
 * Query params for a GET request, from a request message's proto JSON
 * (`XxxRequest.toJSON`): nested messages use dotted keys (`page.index`) and
 * repeated fields stay arrays (send them with `paramsSerializer: { indexes: null }`).
 */
export function toQuery(json: unknown, prefix = ''): Record<string, unknown> {
  const query: Record<string, unknown> = {};
  if (!json || typeof json !== 'object') return query;
  for (const [key, value] of Object.entries(json)) {
    const name = prefix ? `${prefix}.${key}` : key;
    if (value && typeof value === 'object' && !Array.isArray(value)) {
      Object.assign(query, toQuery(value, name));
    } else {
      query[name] = value;
    }
  }
  return query;
}

/**
 * Pages a sample list the way the backend does (core/storage/dbx/postgres
 * Paging): 1-based index, and index or size 0 returns everything.
 */
export function paginate<T>(list: T[], page?: Partial<Pages>): T[] {
  const index = page?.index ?? 0;
  const size = page?.size ?? 0;
  if (index <= 0 || size <= 0) return list;
  return list.slice((index - 1) * size, index * size);
}
