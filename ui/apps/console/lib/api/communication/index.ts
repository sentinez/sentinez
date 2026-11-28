import { Notification } from '@sentinez/proto/sentinez/apps/communication/v1/model';
import {
  ListNotificationRequest,
  ListNotificationResponse,
} from '@sentinez/proto/sentinez/apps/communication/v1/communication';
import axios from 'axios';

import { API_BASE_PATH, authorization } from '@/lib/api/base';
import { fallbackOf } from '@/lib/api/fallback';
import { paginate, toQuery } from '@/lib/api/pages';
import { notificationStore } from '@/lib/faker/communication';
import { byIds, byTenants } from '@/lib/faker/store';

const withFallback = fallbackOf('communication');

export interface ApiOptions {
  signal?: AbortSignal;
}

// GET /communication/notifications
export async function listNotifications(
  req: Partial<ListNotificationRequest> = {},
  options?: ApiOptions,
): Promise<ListNotificationResponse> {
  return withFallback('listNotifications', () => {
    const list = byTenants(byIds(notificationStore.all(), req.ids), req.tenantIds);
    return { notifications: paginate(list, req.page), total: list.length };
  })(async () => {
    const resp = await axios.get(`${API_BASE_PATH}/communication/notifications`, {
      params: toQuery(ListNotificationRequest.toJSON(ListNotificationRequest.fromPartial(req))),
      // repeat the `ids` / `tenantIds` keys (collectionFormat: multi)
      paramsSerializer: { indexes: null },
      signal: options?.signal,
    });
    return ListNotificationResponse.fromJSON(resp.data ?? {});
  });
}

// GET /communication/notifications/stream: the server-sent events of the
// apiserver (internal/defaults.SSEPath), served beside the REST routes
const STREAM_PATH = `${API_BASE_PATH}/communication/notifications/stream`;

/** A notification stream refused or cut by the server: its HTTP status */
export class StreamError extends Error {
  constructor(readonly status: number) {
    super(`notification stream: HTTP ${status}`);
  }
}

export interface StreamOptions extends ApiOptions {
  /** The tenants of the notifications, among those of the user; all without */
  tenantIds?: string[];
  /** Called once the stream is open (authorized) */
  onOpen?: () => void;
}

/**
 * Streams the notifications of the tenants of the signed-in user (of
 * options.tenantIds among them when set), published from now on, to
 * onNotification: resolves when the server ends the stream, rejects with a
 * StreamError when it refuses it. Abort with options.signal.
 *
 * fetch, not EventSource: the stream needs the Authorization header.
 */
export async function streamNotifications(
  onNotification: (n: Notification) => void,
  options?: StreamOptions,
): Promise<void> {
  const headers: HeadersInit = { Accept: 'text/event-stream' };
  const auth = authorization();
  if (auth) headers.Authorization = auth;

  const query = new URLSearchParams(options?.tenantIds?.map((id) => ['tenantIds', id]));
  const resp = await fetch(query.size ? `${STREAM_PATH}?${query}` : STREAM_PATH, {
    headers,
    signal: options?.signal,
    cache: 'no-store',
  });
  if (!resp.ok || !resp.body) throw new StreamError(resp.status);
  options?.onOpen?.();

  for await (const event of serverSentEvents(resp.body)) {
    if (event.event !== 'notification') continue;
    onNotification(Notification.fromJSON(JSON.parse(event.data)));
  }
}

interface ServerSentEvent {
  event: string;
  data: string;
}

/**
 * The events of a text/event-stream body: blocks of `field: value` lines
 * ended by a blank line; comments (`: ping`) and blocks without data are
 * skipped. The apiserver ends its lines with LF only.
 */
async function* serverSentEvents(body: ReadableStream<Uint8Array>) {
  const reader = body.getReader();
  const decoder = new TextDecoder();
  let buffer = '';
  try {
    for (;;) {
      const { value, done } = await reader.read();
      if (done) return;
      // stream: a character split between two chunks is kept for the next
      buffer += decoder.decode(value, { stream: true });
      let end: number;
      while ((end = buffer.indexOf('\n\n')) >= 0) {
        const event = parseEvent(buffer.slice(0, end));
        buffer = buffer.slice(end + 2);
        if (event) yield event;
      }
    }
  } finally {
    reader.releaseLock();
  }
}

function parseEvent(block: string): ServerSentEvent | undefined {
  let event = 'message';
  const data: string[] = [];
  for (const line of block.split('\n')) {
    if (line.startsWith(':')) continue;
    const colon = line.indexOf(':');
    const field = colon < 0 ? line : line.slice(0, colon);
    const value = colon < 0 ? '' : line.slice(colon + 1).replace(/^ /, '');
    if (field === 'event') event = value;
    else if (field === 'data') data.push(value);
  }
  return data.length ? { event, data: data.join('\n') } : undefined;
}
