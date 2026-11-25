import type { Locale } from '@/i18n/config';
import type messages from '@/messages/en.json';

// Type-checks message keys and the locale for next-intl
declare module 'next-intl' {
  interface AppConfig {
    Locale: Locale;
    Messages: typeof messages;
  }
}

// The column options of the console tables
declare module '@tanstack/react-table' {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  interface ColumnMeta<TData, TValue> {
    /** Human-readable column name for the "Columns" visibility menu */
    label?: string;
    /** The skeleton of a cell while the page loads (h-5 w-full by default):
     * a column of taller cells (a chart) gives its own, so the loading rows
     * are as tall as the rows that come */
    skeletonClassName?: string;
  }
}
