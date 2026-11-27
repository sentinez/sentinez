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

import { API_BASE_PATH } from '@/lib/api/base';
import { fallbackOf } from '@/lib/api/fallback';
import { paginate, toQuery } from '@/lib/api/pages';
import { filterSampleUsers, userStore } from '@/lib/faker/iam';

export interface ApiOptions {
  signal?: AbortSignal;
}

const withFallback = fallbackOf('iam');

// GET /iam/user?id=&email=
export async function GetUser(
  params: Partial<GetUserRequest>,
  options?: ApiOptions,
): Promise<GetUserResponse> {
  return withFallback<GetUserResponse>('GetUser', () => ({
    user: params.id ? userStore.find(params.id) : undefined,
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
    const users = filterSampleUsers(params);
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
    userStore.remove(params.id);
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
