import type { Owner } from '@sentinez/proto/sentinez/types/v1/model';

/**
 * The rows of a sample service: the create, update and delete calls change
 * them, so that the later list and get calls show it.
 */
export class SampleStore<T extends { id: string }> {
  private rows: T[];

  constructor(rows: readonly T[]) {
    this.rows = [...rows];
  }

  /** Every row, in order */
  all(): T[] {
    return this.rows;
  }

  find(id: string): T | undefined {
    return this.rows.find((r) => r.id === id);
  }

  add(...rows: T[]): void {
    this.rows = [...this.rows, ...rows];
  }

  /** Puts row in place of the one of id */
  replace(id: string, row: T): void {
    this.rows = this.rows.map((r) => (r.id === id ? row : r));
  }

  remove(id: string): void {
    this.rows = this.rows.filter((r) => r.id !== id);
  }
}

/** The rows of list of one of ids, every row without ids */
export const byIds = <T extends { id: string }>(list: T[], ids?: string[]) =>
  ids?.length ? list.filter((r) => ids.includes(r.id)) : list;

/** The rows of list owned by one of the resources of ids, every row without ids */
export const byResources = <T extends { owner?: Owner }>(list: T[], ids?: string[]) =>
  ids?.length ? list.filter((r) => ids.includes(r.owner?.resourceId ?? '')) : list;

/** The rows of list of one of the tenants of ids, every row without ids */
export const byTenants = <T extends { tenantId: string }>(list: T[], ids?: string[]) =>
  ids?.length ? list.filter((r) => ids.includes(r.tenantId)) : list;
