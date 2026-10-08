/** Wire format of `v1Pages` (sentinez.types.v1.Pages) */
export interface Pages {
  /** 1-based page number */
  index?: number;
  size?: number;
  /** Ask the server to count the total rows */
  total?: boolean;
}

/** Query params for a `page` field: nested names use dotted keys */
export function pageQuery(page?: Pages): Record<string, unknown> {
  const query: Record<string, unknown> = {};
  if (page?.index !== undefined) query['page.index'] = page.index;
  if (page?.size !== undefined) query['page.size'] = page.size;
  if (page?.total !== undefined) query['page.total'] = page.total;
  return query;
}

/**
 * Pages a sample list the way the backend does (core/storage/dbx/postgres
 * Paging): 1-based index, and index or size 0 returns everything.
 */
export function paginate<T>(list: T[], page?: Pages): T[] {
  const index = page?.index ?? 0;
  const size = page?.size ?? 0;
  if (index <= 0 || size <= 0) return list;
  return list.slice((index - 1) * size, index * size);
}
