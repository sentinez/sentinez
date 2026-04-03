import { toast } from '@/lib/toast';

/**
 * Handle API status errors and show notifications.
 * This function is safe to call on both client and server.
 */
export const handleStatusError = (error: any) => {
  if (typeof window === 'undefined') return;

  const status = error.response?.status;
  const message = error.response?.data?.message || error.message;

  if (status === 404) {
    toast.error('Not Found', 'The requested resource could not be located.');
  } else if (status >= 500) {
    toast.error('Server Error', 'An internal server error occurred. Please try again later.');
  } else {
    toast.error('Error', message || 'An unexpected error occurred.');
  }
};
