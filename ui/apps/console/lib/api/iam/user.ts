import axios from 'axios';
import {
  DeleteUserRequest,
  DeleteUserResponse,
  GetUserRequest,
  GetUserResponse,
  ListAccountsRequest,
  ListAccountsResponse,
  ListUsersRequest,
  ListUsersResponse,
  UpdateUserRequest,
  UpdateUserResponse,
} from '@sentinez/proto/sentinez/apps/iam/v1/iam';
import { User } from '@sentinez/proto/sentinez/apps/iam/v1/model';
import { ControlPlane } from '@sentinez/proto/sentinez/types/v1/known';

import { API_BASE_PATH } from '@/lib/api/base';
import { paginate, toQuery } from '@/lib/api/pages';

// Set NEXT_PUBLIC_USE_SAMPLE=true to return sample responses when an API call fails
const USE_SAMPLE = process.env.NEXT_PUBLIC_USE_SAMPLE === 'true';

export interface ApiOptions {
  signal?: AbortSignal;
}

// Written as proto JSON (the wire format), parsed with User.fromJSON
export const SAMPLE_USERS: User[] = [
  {
    metadata: { createdAt: '2026-08-20T03:00:00Z', updatedAt: '2026-08-20T03:00:00Z' },
    id: 'senz.iam.users.0199a6f0-4c1e-7a2b-9d3e-400000000001',
    fullName: 'Alice Nguyen',
    email: 'alice@example.com',
    controlPlane: 'CONTROL_PLANE_PORTAL',
  },
  {
    metadata: { createdAt: '2026-08-25T09:30:00Z', updatedAt: '2026-09-10T11:00:00Z' },
    id: 'senz.iam.users.0199a6f0-4c1e-7a2b-9d3e-400000000002',
    fullName: 'Bob Tran',
    email: 'bob@example.com',
    controlPlane: 'CONTROL_PLANE_PORTAL',
  },
  {
    metadata: { createdAt: '2026-09-02T08:00:00Z', updatedAt: '2026-09-02T08:00:00Z' },
    id: 'senz.iam.users.0199a6f0-4c1e-7a2b-9d3e-400000000003',
    fullName: 'Carol Le',
    email: 'carol@example.com',
    controlPlane: 'CONTROL_PLANE_PORTAL',
  },
].map((u) => User.fromJSON(u));

// Extra users so member lists have several pages; the last few belong to no
// tenant, so they can be invited (try dung@example.com)
const SAMPLE_NAMES = [
  'Duc Pham',
  'Emma Hoang',
  'Felix Vu',
  'Giang Dang',
  'Hana Bui',
  'Ivan Do',
  'Julia Ho',
  'Khoa Ngo',
  'Linh Duong',
  'Minh Ly',
  'Nam Truong',
  'Oanh Mai',
  'Phuc Dinh',
  'Quynh Lam',
  'Dung Cao',
  'Rin Ta',
];

const generatedUsers: User[] = SAMPLE_NAMES.map((fullName, i) => {
  const n = String(i + 4).padStart(2, '0');
  const at = new Date(Date.UTC(2026, 8, 1 + i, 8));
  return User.fromPartial({
    metadata: { createdAt: at, updatedAt: at },
    id: `senz.iam.users.0199a6f0-4c1e-7a2b-9d3e-4000000000${n}`,
    fullName,
    // "Duc Pham" -> duc@example.com
    email: `${fullName.slice(0, fullName.indexOf(' ')).toLowerCase()}@example.com`,
    controlPlane: ControlPlane.CONTROL_PLANE_PORTAL,
  });
});

/** Every sample user: SAMPLE_USERS first, then the generated ones */
export const ALL_SAMPLE_USERS: User[] = [...SAMPLE_USERS, ...generatedUsers];

let sampleStore: User[] = [...ALL_SAMPLE_USERS];

function withFallback<T>(label: string, fallback: () => T) {
  return async (call: () => Promise<T>): Promise<T> => {
    try {
      return await call();
    } catch (err: any) {
      if (!USE_SAMPLE || axios.isCancel(err)) throw err;
      console.warn(`[iam] ${label} failed, using sample response:`, err?.message);
      return fallback();
    }
  };
}

// GET /iam/user?id=&email=
export async function GetUser(
  params: Partial<GetUserRequest>,
  options?: ApiOptions,
): Promise<GetUserResponse> {
  return withFallback<GetUserResponse>('GetUser', () => ({
    user: sampleStore.find((u) => !!params.id && u.id === params.id),
  }))(async () => {
    const endpoint = `${API_BASE_PATH}/iam/user`;
    const resp = await axios.get(endpoint, {
      params: toQuery(GetUserRequest.toJSON(GetUserRequest.fromPartial(params))),
      signal: options?.signal,
    });
    return GetUserResponse.fromJSON(resp.data ?? {});
  });
}

// PUT /iam/user: renames the signed-in user
export async function UpdateUser(
  req: UpdateUserRequest,
  options?: ApiOptions,
): Promise<UpdateUserResponse> {
  const endpoint = `${API_BASE_PATH}/iam/user`;
  const resp = await axios.put(endpoint, UpdateUserRequest.toJSON(req), {
    signal: options?.signal,
  });
  return UpdateUserResponse.fromJSON(resp.data ?? {});
}

// GET /iam/users
export async function ListUsers(
  params: Partial<ListUsersRequest> = {},
  options?: ApiOptions,
): Promise<ListUsersResponse> {
  return withFallback<ListUsersResponse>('ListUsers', () => {
    const users = sampleStore.filter(
      (u) =>
        (!params.ids?.length || params.ids.includes(u.id)) &&
        (!params.emails?.length || params.emails.includes(u.email)),
    );
    return { users: paginate(users, params.page), total: users.length };
  })(async () => {
    const endpoint = `${API_BASE_PATH}/iam/users`;
    const resp = await axios.get(endpoint, {
      params: toQuery(ListUsersRequest.toJSON(ListUsersRequest.fromPartial(params))),
      // repeat the `ids` / `emails` keys (collectionFormat: multi)
      paramsSerializer: { indexes: null },
      signal: options?.signal,
    });
    return ListUsersResponse.fromJSON(resp.data ?? {});
  });
}

// DELETE /iam/user/{id}
export async function DeleteUser(
  params: DeleteUserRequest,
  options?: ApiOptions,
): Promise<DeleteUserResponse> {
  return withFallback<DeleteUserResponse>('DeleteUser', () => {
    sampleStore = sampleStore.filter((u) => u.id !== params.id);
    return {};
  })(async () => {
    const endpoint = `${API_BASE_PATH}/iam/user/${encodeURIComponent(params.id)}`;
    const resp = await axios.delete(endpoint, { signal: options?.signal });
    return DeleteUserResponse.fromJSON(resp.data ?? {});
  });
}

// GET /iam/accounts
export async function ListAccounts(
  params: Partial<ListAccountsRequest> = {},
  options?: ApiOptions,
): Promise<ListAccountsResponse> {
  const endpoint = `${API_BASE_PATH}/iam/accounts`;
  const resp = await axios.get(endpoint, {
    params: toQuery(ListAccountsRequest.toJSON(ListAccountsRequest.fromPartial(params))),
    // repeat the keys of repeated filters (collectionFormat: multi)
    paramsSerializer: { indexes: null },
    signal: options?.signal,
  });
  return ListAccountsResponse.fromJSON(resp.data ?? {});
}
