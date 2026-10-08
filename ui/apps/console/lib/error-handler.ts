import type { useTranslations } from 'next-intl';
import { toast } from '@/lib/toast';

export type ErrorsTranslator = ReturnType<typeof useTranslations<'Errors'>>;

/**
 * Handle API status errors and show notifications.
 * This function is safe to call on both client and server.
 * `t` comes from `useTranslations('Errors')` in the caller.
 */
export const handleStatusError = (error: any, t: ErrorsTranslator) => {
  if (typeof window === 'undefined') return;

  const status = error.response?.status;
  const message = error.response?.data?.message || error.message;

  if (status === 404) {
    toast.error(t('notFound'), t('notFoundDescription'));
  } else if (status >= 500) {
    toast.error(t('server'), t('serverDescription'));
  } else {
    toast.error(t('generic'), message || t('unexpected'));
  }
};
