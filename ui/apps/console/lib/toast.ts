import { toast as sonnerToast } from 'sonner';

/**
 * Enhanced toast utility for the console application.
 * Wraps sonner toast with a consistent API and safe client-side execution.
 */
export const toast = {
  success: (title: string, description?: string) => {
    return sonnerToast.success(title, { description });
  },
  error: (title: string, description?: string) => {
    return sonnerToast.error(title, { description });
  },
  info: (title: string, description?: string) => {
    return sonnerToast.info(title, { description });
  },
  warning: (title: string, description?: string) => {
    return sonnerToast.warning(title, { description });
  },
  loading: (title: string, description?: string) => {
    return sonnerToast.loading(title, { description });
  },
  dismiss: (id?: string | number) => {
    return sonnerToast.dismiss(id);
  },
};
