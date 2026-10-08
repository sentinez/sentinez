import type { Locale } from '@/i18n/config';
import type messages from '@/messages/en.json';

// Type-checks message keys and the locale for next-intl
declare module 'next-intl' {
  interface AppConfig {
    Locale: Locale;
    Messages: typeof messages;
  }
}

// Human-readable column name for the "Columns" visibility menu
declare module '@tanstack/react-table' {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  interface ColumnMeta<TData, TValue> {
    label?: string;
  }
}
