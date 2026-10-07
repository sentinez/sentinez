import axios from 'axios';

import { API_BASE_PATH } from '@/lib/api/base';

export interface ApiOptions {
  signal?: AbortSignal;
}

export async function getResource(id: string, options?: ApiOptions) {
  try {
    const endpoint = `${API_BASE_PATH}/tenant/resource`;
    const resp = await axios.get(endpoint, {
      params: { id },
      signal: options?.signal,
    });
    return resp.data.resource ?? null;
  } catch (err: any) {
    if (axios.isCancel(err)) throw err;
    console.error('Error fetching resource:', err.message);
    return null;
  }
}

export async function getResourceByDomain(domain: string, options?: ApiOptions) {
  try {
    const endpoint = `${API_BASE_PATH}/tenant/resource/${domain}`;
    const resp = await axios.get(endpoint, { signal: options?.signal });
    return resp.data.resource ?? null;
  } catch (err: any) {
    if (axios.isCancel(err)) throw err;
    // console.error('Error fetching resource by domain:', err.message);
    return null;
  }
}

export async function listResources(
  params?: {
    resourceDomain?: string;
    resourceName?: string;
    status?: string;
    plan?: string;
  },
  options?: ApiOptions,
) {
  try {
    const endpoint = `${API_BASE_PATH}/tenant/resources`;
    const resp = await axios.get(endpoint, {
      params,
      signal: options?.signal,
    });
    return resp.data ?? null;
  } catch (err: any) {
    if (axios.isCancel(err)) throw err;
    console.error('Error listing resources:', err.message);
    return null;
  }
}
