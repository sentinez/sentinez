import axios from 'axios';

import { SENTINEZ_ACCESS_TOKEN_KEY } from '@/lib/const';

// Client-side requests go through the Next.js proxy at app/api/[...path],
// which forwards them to the server-only API_BASE_PATH.
export const API_BASE_PATH = '/api';

// Every lib/api module imports this file, so registering here covers all API
// calls. The token is only sent to our own API, never to another origin.
axios.interceptors.request.use((config) => {
  if (typeof window === 'undefined') return config;
  if (!config.url?.startsWith(`${API_BASE_PATH}/`)) return config;
  if (config.headers.has('Authorization')) return config;

  const token = window.localStorage.getItem(SENTINEZ_ACCESS_TOKEN_KEY);
  if (token) {
    config.headers.set('Authorization', `Bearer ${token}`);
  }
  return config;
});
