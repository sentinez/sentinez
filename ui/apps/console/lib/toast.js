import { toast as sonnerToast } from 'sonner';
/**
 * Enhanced toast utility for the console application.
 * Wraps sonner toast with a consistent API and safe client-side execution.
 */
export const toast = {
    success: (title, description) => {
        return sonnerToast.success(title, { description });
    },
    error: (title, description) => {
        return sonnerToast.error(title, { description });
    },
    info: (title, description) => {
        return sonnerToast.info(title, { description });
    },
    warning: (title, description) => {
        return sonnerToast.warning(title, { description });
    },
    loading: (title, description) => {
        return sonnerToast.loading(title, { description });
    },
    dismiss: (id) => {
        return sonnerToast.dismiss(id);
    },
};
