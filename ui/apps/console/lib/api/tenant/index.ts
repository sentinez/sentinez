import axios from 'axios';

const API_BASE_PATH = process.env.SNTZ_BASE_PATH || 'http://localhost:8080';

export async function getResource(id: string) {
  try {
    const endpoint = `${API_BASE_PATH}/tenant/resource`;
    const resp = await axios.get(endpoint, { params: { id } });
    return resp.data.resource ?? null;
  } catch (err: any) {
    console.error('Error fetching resource:', err.message);
    return null;
  }
}

export async function getResourceByDomain(domain: string) {
  try {
    const endpoint = `${API_BASE_PATH}/tenant/resource`;
    const resp = await axios.get(endpoint, { params: { resourceDomain: domain } });
    return resp.data.resource ?? null;
  } catch (err: any) {
    console.error('Error fetching resource by domain:', err.message);
    return null;
  }
}

export async function listResources(params?: {
  resourceDomain?: string;
  resourceName?: string;
  status?: string;
  plan?: string;
}) {
  try {
    const endpoint = `${API_BASE_PATH}/tenant/resources`;
    const resp = await axios.get(endpoint, { params });
    return resp.data ?? null;
  } catch (err: any) {
    console.error('Error listing resources:', err.message);
    return null;
  }
}
