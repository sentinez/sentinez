import type { ReactNode } from 'react';
import type { ColumnDef } from '@tanstack/react-table';

import type { EventKind } from '@sentinez/proto/sentinez/apps/analytic/v1/event';
import type { ApiOptions, Pages } from '@/lib/api/analytic';

// ListRequest the request of a page of events, the latest first
export type ListRequest = {
  resourceIds: string[];
  /** The requests of the events: those the filter names (requestIdsOf) */
  requestIds: string[];
  /** The field filters of the page, all of them, as filterParam makes them */
  filters: string[];
  page: Pages;
  from?: Date;
  to?: Date;
};

// FieldFilter the events holding value in the field of the name field
export type FieldFilter = { field: string; value: string };

// The API form of a filter, "field:value" (the field before the first colon)
export const filterParam = (f: FieldFilter) => `${f.field}:${f.value}`;

// SeriesDef a series of the histogram, the events of some keys
export interface SeriesDef {
  /** A CSS-safe id, the key of the chart config */
  id: string;
  label: string;
  color: string;
}

// FieldDef a field of the events: its top values, and a filter term
export interface FieldDef<T> {
  /** The name of the API (the proto field name), the key of a filter term */
  name: string;
  label: string;
  /** The values of an event a `name:value` term compares */
  get: (e: T) => string | string[];
  /** The label of a top value, the value itself by default */
  format?: (value: string) => string;
}

// RequestRef the request received by the edge an event is about
export type RequestRef = { id: string; time: number };

// DetailSection a titled list of name / value pairs of the detail sheet
export interface DetailSection {
  title: string;
  pairs: [string, ReactNode][];
}

/**
 * EventConfig what the explorer of a kind of events shows: how it lists
 * them, their columns, the series of their histogram, their fields and
 * their detail.
 */
export interface EventConfig<T> {
  kind: EventKind;
  title: string;
  subtitle: string;
  list: (req: ListRequest, options?: ApiOptions) => Promise<T[]>;
  idOf: (e: T) => string;
  /** When the event happened, in millis */
  timeOf: (e: T) => number;
  /** The columns after the time one */
  columns: ColumnDef<T>[];
  series: SeriesDef[];
  /** The series of a key of the histogram (see GetEventHistogramRequest) */
  seriesOf: (key: number) => string;
  /** The fields of the field list, in this order */
  fields: FieldDef<T>[];
  /** The texts a free-text filter term is searched in */
  text: (e: T) => string[];
  /** The request an event is about, linked to the ingress page */
  requestOf?: (e: T) => RequestRef | undefined;
  /**
   * The event of an id among those of a resource, null when none: the page
   * opens the one of its `request` param (see REQUEST_PARAM)
   */
  /** The event of a request id, looked up within from-to when given */
  byId?: (
    req: { id: string; resourceId: string; from?: Date; to?: Date },
    options?: ApiOptions,
  ) => Promise<T | null>;
  detail: {
    title: (e: T) => ReactNode;
    /** The rows after the time one */
    rows: (e: T) => [string, ReactNode][];
    sections?: (e: T) => DetailSection[];
    toJSON: (e: T) => unknown;
  };
}
