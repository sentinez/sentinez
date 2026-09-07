import axios from 'axios';
const API_BASE_PATH = process.env.SNTZ_BASE_PATH || 'http://localhost:8080';
export async function getResource(id, options) {
    try {
        const endpoint = `${API_BASE_PATH}/tenant/resource`;
        const resp = await axios.get(endpoint, {
            params: { id },
            signal: options?.signal,
        });
        return resp.data.resource ?? null;
    }
    catch (err) {
        if (axios.isCancel(err))
            throw err;
        console.error('Error fetching resource:', err.message);
        return null;
    }
}
export async function getResourceByDomain(domain, options) {
    try {
        const endpoint = `${API_BASE_PATH}/tenant/resource/${domain}`;
        const resp = await axios.get(endpoint, { signal: options?.signal });
        return resp.data.resource ?? null;
    }
    catch (err) {
        if (axios.isCancel(err))
            throw err;
        console.error('Error fetching resource by domain:', err.message);
        return null;
    }
}
export async function listResources(params, options) {
    try {
        const endpoint = `${API_BASE_PATH}/tenant/resources`;
        const resp = await axios.get(endpoint, {
            params,
            signal: options?.signal,
        });
        return resp.data ?? null;
    }
    catch (err) {
        if (axios.isCancel(err))
            throw err;
        console.error('Error listing resources:', err.message);
        return null;
    }
}
