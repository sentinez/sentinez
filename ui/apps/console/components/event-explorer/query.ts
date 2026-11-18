import type { EventConfig } from './types';

// The value of a field without its whitespace, as a filter term holds it
export const filterValueOf = (value: string): string => value.replace(/\s+/g, '');

// A request id, as the edge makes them (an optional prefix, a ULID)
const REQUEST_ID = /^(?:[a-zA-Z0-9-]+-)?req_[0-9A-HJKMNP-TV-Z]{26}$/;

// The terms of a filter that are ids of the pattern, as terms of their own
export const idsMatching =
  (pattern: RegExp) =>
  (query: string): string[] =>
    query.split(/\s+/).filter((term) => pattern.test(term));

/**
 * The request ids a filter names, as terms of their own: the list asks the
 * API for their events, which may be on any page.
 */
export const requestIdsOf = idsMatching(REQUEST_ID);

// The terms of a filter: lowercase, split on whitespace
export const termsOf = (query: string): string[] =>
  query.trim().toLowerCase().split(/\s+/).filter(Boolean);

/**
 * Whether an event matches all the terms of a filter (termsOf):
 * `field:value` compares a field of the config to the value (case and
 * whitespace ignored; any element of a repeated one), any other term is
 * searched in its texts.
 */
export function matchesTerms<T>(
  e: T,
  terms: string[],
  config: Pick<EventConfig<T>, 'fields' | 'text'>,
): boolean {
  return terms.every((term) => {
    const [name, ...rest] = term.split(':');
    const field = config.fields.find((f) => f.name === name);
    if (field && rest.length) {
      const value = rest.join(':');
      const got = field.get(e);
      return (Array.isArray(got) ? got : [got]).some(
        (v) => filterValueOf(v).toLowerCase() === value,
      );
    }
    return config.text(e).some((s) => s.toLowerCase().includes(term));
  });
}
