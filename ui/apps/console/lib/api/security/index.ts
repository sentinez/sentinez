import { SecRuleset, SecRateLimit, SecRule } from '@sentinez/proto/sentinez/apps/security/v1/model';
import {
  CreateSecRulesetRequest,
  CreateSecRulesetResponse,
  CreateSecRateLimitRequest,
  CreateSecRateLimitResponse,
  CreateSecRuleRequest,
  CreateSecRuleResponse,
  DeleteSecRulesetRequest,
  DeleteSecRulesetResponse,
  DeleteSecRateLimitRequest,
  DeleteSecRateLimitResponse,
  DeleteSecRuleRequest,
  DeleteSecRuleResponse,
  GetSecRulesetRequest,
  GetSecRulesetResponse,
  GetSecRateLimitRequest,
  GetSecRateLimitResponse,
  GetCRSRuleInfoRequest,
  GetCRSRuleInfoResponse,
  GetConditionSchemaResponse,
  GetSecRuleRequest,
  GetSecRuleResponse,
  ListSecRulesetsRequest,
  ListSecRulesetsResponse,
  ListSecRateLimitsRequest,
  ListSecRateLimitsResponse,
  ListCRSRuleInfosRequest,
  ListCRSRuleInfosResponse,
  ListCRSRuleTagsRequest,
  ListCRSRuleTagsResponse,
  ListSecRulesRequest,
  ListSecRulesResponse,
  StatusResponse,
  UpdateSecRulesetRequest,
  UpdateSecRulesetResponse,
  UpdateSecRateLimitRequest,
  UpdateSecRateLimitResponse,
  UpdateSecRuleRequest,
  UpdateSecRuleResponse,
} from '@sentinez/proto/sentinez/apps/security/v1/security';
import axios from 'axios';

import { API_BASE_PATH } from '@/lib/api/base';
import { fallbackOf } from '@/lib/api/fallback';
import { type Pages, paginate, toQuery } from '@/lib/api/pages';
import {
  SAMPLE_CONDITION_SCHEMA,
  SAMPLE_RULE_INFOS,
  filterCRSRuleInfos,
  filterRulesets,
  rateLimitStore,
  ruleStore,
  rulesetStore,
  sampleCRSRuleTags,
} from '@/lib/faker/security';
import { byIds, byResources } from '@/lib/faker/store';
export interface ApiOptions {
  signal?: AbortSignal;
}

export type { Pages };

const withFallback = fallbackOf('security');

// GET /security/secrules
export async function listSecRules(
  req: Partial<ListSecRulesRequest> = {},
  options?: ApiOptions,
): Promise<ListSecRulesResponse> {
  return withFallback('listSecRules', () => {
    const list = byResources(byIds(ruleStore.all(), req.ids), req.resourceIds);
    return { secrules: paginate(list, req.page), total: list.length };
  })(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/security/secrules`, {
      params: toQuery(ListSecRulesRequest.toJSON(ListSecRulesRequest.fromPartial(req))),
      // repeat the `ids` / `resourceIds` keys (collectionFormat: multi)
      paramsSerializer: { indexes: null },
      signal: options?.signal,
    });
    return ListSecRulesResponse.fromJSON(resp.data ?? {});
  });
}

// GET /security/secrule/{id}
export async function getSecRule(
  req: GetSecRuleRequest,
  options?: ApiOptions,
): Promise<GetSecRuleResponse> {
  return withFallback<GetSecRuleResponse>('getSecRule', () => ({
    secrule: ruleStore.find(req.id) ?? ruleStore.all()[0],
  }))(async () => {
    const endpoint = `${API_BASE_PATH}/security/secrule/${encodeURIComponent(req.id)}`;
    const resp = await axios.get(endpoint, { signal: options?.signal });
    return GetSecRuleResponse.fromJSON(resp.data ?? {});
  });
}

// POST /security/secrule
export async function createSecRule(
  req: CreateSecRuleRequest,
  options?: ApiOptions,
): Promise<CreateSecRuleResponse> {
  return withFallback('createSecRule', () => {
    const id = `sample-rule-${Date.now()}`;
    ruleStore.add(SecRule.fromPartial({ ...req.secrule, id }));
    return { id };
  })(async () => {
    const resp = await axios.post(
      `${API_BASE_PATH}/security/secrule`,
      CreateSecRuleRequest.toJSON(req),
      { signal: options?.signal },
    );
    return CreateSecRuleResponse.fromJSON(resp.data ?? {});
  });
}

// PUT /security/secrule/{id}
export async function updateSecRule(
  req: UpdateSecRuleRequest,
  options?: ApiOptions,
): Promise<UpdateSecRuleResponse> {
  return withFallback<UpdateSecRuleResponse>('updateSecRule', () => {
    // The owner of a row never changes
    const owner = ruleStore.find(req.id)?.owner;
    const next = SecRule.fromPartial({ ...req.secrule, id: req.id, owner });
    ruleStore.replace(req.id, next);
    return { secrule: next };
  })(async () => {
    const resp = await axios.put(
      `${API_BASE_PATH}/security/secrule/${encodeURIComponent(req.id)}`,
      UpdateSecRuleRequest.toJSON(req),
      { signal: options?.signal },
    );
    return UpdateSecRuleResponse.fromJSON(resp.data ?? {});
  });
}

// DELETE /security/secrule/{id}
export async function deleteSecRule(
  req: DeleteSecRuleRequest,
  options?: ApiOptions,
): Promise<DeleteSecRuleResponse> {
  return withFallback<DeleteSecRuleResponse>('deleteSecRule', () => {
    ruleStore.remove(req.id);
    return {};
  })(async () => {
    const endpoint = `${API_BASE_PATH}/security/secrule/${encodeURIComponent(req.id)}`;
    const resp = await axios.delete(endpoint, { signal: options?.signal });
    return DeleteSecRuleResponse.fromJSON(resp.data ?? {});
  });
}

// GET /security/secratelimits
export async function listSecRateLimits(
  req: Partial<ListSecRateLimitsRequest> = {},
  options?: ApiOptions,
): Promise<ListSecRateLimitsResponse> {
  return withFallback('listSecRateLimits', () => {
    const list = byResources(byIds(rateLimitStore.all(), req.ids), req.resourceIds);
    return { secratelimits: paginate(list, req.page), total: list.length };
  })(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/security/secratelimits`, {
      params: toQuery(ListSecRateLimitsRequest.toJSON(ListSecRateLimitsRequest.fromPartial(req))),
      // repeat the `ids` / `resourceIds` keys (collectionFormat: multi)
      paramsSerializer: { indexes: null },
      signal: options?.signal,
    });
    return ListSecRateLimitsResponse.fromJSON(resp.data ?? {});
  });
}

// GET /security/secratelimit/{id}
export async function getSecRateLimit(
  req: GetSecRateLimitRequest,
  options?: ApiOptions,
): Promise<GetSecRateLimitResponse> {
  return withFallback<GetSecRateLimitResponse>('getSecRateLimit', () => ({
    secratelimit: rateLimitStore.find(req.id) ?? rateLimitStore.all()[0],
  }))(async () => {
    const endpoint = `${API_BASE_PATH}/security/secratelimit/${encodeURIComponent(req.id)}`;
    const resp = await axios.get(endpoint, { signal: options?.signal });
    return GetSecRateLimitResponse.fromJSON(resp.data ?? {});
  });
}

// POST /security/secratelimit
export async function createSecRateLimit(
  req: CreateSecRateLimitRequest,
  options?: ApiOptions,
): Promise<CreateSecRateLimitResponse> {
  return withFallback('createSecRateLimit', () => {
    const id = `sample-ratelimit-${Date.now()}`;
    rateLimitStore.add(SecRateLimit.fromPartial({ ...req.secratelimit, id }));
    return { id };
  })(async () => {
    const resp = await axios.post(
      `${API_BASE_PATH}/security/secratelimit`,
      CreateSecRateLimitRequest.toJSON(req),
      { signal: options?.signal },
    );
    return CreateSecRateLimitResponse.fromJSON(resp.data ?? {});
  });
}

// PUT /security/secratelimit/{id}
export async function updateSecRateLimit(
  req: UpdateSecRateLimitRequest,
  options?: ApiOptions,
): Promise<UpdateSecRateLimitResponse> {
  return withFallback<UpdateSecRateLimitResponse>('updateSecRateLimit', () => {
    const owner = rateLimitStore.find(req.id)?.owner;
    const next = SecRateLimit.fromPartial({ ...req.secratelimit, id: req.id, owner });
    rateLimitStore.replace(req.id, next);
    return { secratelimit: next };
  })(async () => {
    const resp = await axios.put(
      `${API_BASE_PATH}/security/secratelimit/${encodeURIComponent(req.id)}`,
      UpdateSecRateLimitRequest.toJSON(req),
      { signal: options?.signal },
    );
    return UpdateSecRateLimitResponse.fromJSON(resp.data ?? {});
  });
}

// DELETE /security/secratelimit/{id}
export async function deleteSecRateLimit(
  req: DeleteSecRateLimitRequest,
  options?: ApiOptions,
): Promise<DeleteSecRateLimitResponse> {
  return withFallback<DeleteSecRateLimitResponse>('deleteSecRateLimit', () => {
    rateLimitStore.remove(req.id);
    return {};
  })(async () => {
    const endpoint = `${API_BASE_PATH}/security/secratelimit/${encodeURIComponent(req.id)}`;
    const resp = await axios.delete(endpoint, { signal: options?.signal });
    return DeleteSecRateLimitResponse.fromJSON(resp.data ?? {});
  });
}

// GET /security/secrulesets
export async function listSecRulesets(
  req: Partial<ListSecRulesetsRequest> = {},
  options?: ApiOptions,
): Promise<ListSecRulesetsResponse> {
  return withFallback('listSecRulesets', () => {
    const list = filterRulesets(req);
    return { secrulesets: paginate(list, req.page), total: list.length };
  })(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/security/secrulesets`, {
      params: toQuery(ListSecRulesetsRequest.toJSON(ListSecRulesetsRequest.fromPartial(req))),
      // repeat the keys of repeated filters (collectionFormat: multi)
      paramsSerializer: { indexes: null },
      signal: options?.signal,
    });
    return ListSecRulesetsResponse.fromJSON(resp.data ?? {});
  });
}

// GET /security/secruleset/{id}
export async function getSecRuleset(
  req: GetSecRulesetRequest,
  options?: ApiOptions,
): Promise<GetSecRulesetResponse> {
  return withFallback<GetSecRulesetResponse>('getSecRuleset', () => ({
    secruleset: rulesetStore.find(req.id) ?? rulesetStore.all()[0],
  }))(async () => {
    const endpoint = `${API_BASE_PATH}/security/secruleset/${encodeURIComponent(req.id)}`;
    const resp = await axios.get(endpoint, { signal: options?.signal });
    return GetSecRulesetResponse.fromJSON(resp.data ?? {});
  });
}

// POST /security/secruleset
export async function createSecRuleset(
  req: CreateSecRulesetRequest,
  options?: ApiOptions,
): Promise<CreateSecRulesetResponse> {
  return withFallback('createSecRuleset', () => {
    const id = `sample-ruleset-${Date.now()}`;
    rulesetStore.add(SecRuleset.fromPartial({ ...req.secruleset, id }));
    return { id };
  })(async () => {
    const resp = await axios.post(
      `${API_BASE_PATH}/security/secruleset`,
      CreateSecRulesetRequest.toJSON(req),
      { signal: options?.signal },
    );
    return CreateSecRulesetResponse.fromJSON(resp.data ?? {});
  });
}

// PUT /security/secruleset/{id}
export async function updateSecRuleset(
  req: UpdateSecRulesetRequest,
  options?: ApiOptions,
): Promise<UpdateSecRulesetResponse> {
  return withFallback<UpdateSecRulesetResponse>('updateSecRuleset', () => {
    const owner = rulesetStore.find(req.id)?.owner;
    const next = SecRuleset.fromPartial({ ...req.secruleset, id: req.id, owner });
    rulesetStore.replace(req.id, next);
    return { secruleset: next };
  })(async () => {
    const resp = await axios.put(
      `${API_BASE_PATH}/security/secruleset/${encodeURIComponent(req.id)}`,
      UpdateSecRulesetRequest.toJSON(req),
      { signal: options?.signal },
    );
    return UpdateSecRulesetResponse.fromJSON(resp.data ?? {});
  });
}

// DELETE /security/secruleset/{id}
export async function deleteSecRuleset(
  req: DeleteSecRulesetRequest,
  options?: ApiOptions,
): Promise<DeleteSecRulesetResponse> {
  return withFallback<DeleteSecRulesetResponse>('deleteSecRuleset', () => {
    rulesetStore.remove(req.id);
    return {};
  })(async () => {
    const endpoint = `${API_BASE_PATH}/security/secruleset/${encodeURIComponent(req.id)}`;
    const resp = await axios.delete(endpoint, { signal: options?.signal });
    return DeleteSecRulesetResponse.fromJSON(resp.data ?? {});
  });
}

// GET /security/crs/rules
export async function listCRSRuleInfos(
  req: Partial<ListCRSRuleInfosRequest> = {},
  options?: ApiOptions,
): Promise<ListCRSRuleInfosResponse> {
  return withFallback('listCRSRuleInfos', () => {
    const list = filterCRSRuleInfos(req);
    return { rules: paginate(list, req.page), total: list.length };
  })(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/security/crs/rules`, {
      params: toQuery(ListCRSRuleInfosRequest.toJSON(ListCRSRuleInfosRequest.fromPartial(req))),
      // repeat the keys of repeated filters (collectionFormat: multi)
      paramsSerializer: { indexes: null },
      signal: options?.signal,
    });
    return ListCRSRuleInfosResponse.fromJSON(resp.data ?? {});
  });
}

// GET /security/crs/tags
export async function listCRSRuleTags(
  req: Partial<ListCRSRuleTagsRequest> = {},
  options?: ApiOptions,
): Promise<ListCRSRuleTagsResponse> {
  return withFallback('listCRSRuleTags', () => {
    const list = sampleCRSRuleTags(req);
    return { tags: paginate(list, req.page), total: list.length };
  })(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/security/crs/tags`, {
      params: toQuery(ListCRSRuleTagsRequest.toJSON(ListCRSRuleTagsRequest.fromPartial(req))),
      signal: options?.signal,
    });
    return ListCRSRuleTagsResponse.fromJSON(resp.data ?? {});
  });
}

// GET /security/crs/rules/{id}
export async function getCRSRuleInfo(
  req: Partial<GetCRSRuleInfoRequest> & Pick<GetCRSRuleInfoRequest, 'id'>,
  options?: ApiOptions,
): Promise<GetCRSRuleInfoResponse> {
  return withFallback<GetCRSRuleInfoResponse>('getCRSRuleInfo', () => ({
    rule: SAMPLE_RULE_INFOS.find((r) => r.id === req.id),
  }))(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/security/crs/rules/${req.id}`, {
      // `id` is the path param: zero it so toJSON leaves it out of the query
      params: toQuery(
        GetCRSRuleInfoRequest.toJSON(GetCRSRuleInfoRequest.fromPartial({ ...req, id: 0 })),
      ),
      signal: options?.signal,
    });
    return GetCRSRuleInfoResponse.fromJSON(resp.data ?? {});
  });
}

// GET /security/condition-schema
// The conditions a rule expression accepts: the editor offers what it lists.
export async function getConditionSchema(
  options?: ApiOptions,
): Promise<GetConditionSchemaResponse> {
  return withFallback<GetConditionSchemaResponse>('getConditionSchema', () => ({
    schema: SAMPLE_CONDITION_SCHEMA,
  }))(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/security/condition-schema`, {
      signal: options?.signal,
    });
    return GetConditionSchemaResponse.fromJSON(resp.data ?? {});
  });
}

// GET /security/status
export async function getSecurityStatus(options?: ApiOptions): Promise<StatusResponse> {
  return withFallback<StatusResponse>('getSecurityStatus', () => ({
    msg: 'sample: security service OK',
  }))(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/security/status`, { signal: options?.signal });
    return StatusResponse.fromJSON(resp.data ?? {});
  });
}
