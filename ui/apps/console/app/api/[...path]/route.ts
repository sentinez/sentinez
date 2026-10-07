import { NextRequest } from 'next/server';

// Server-only upstream API (not exposed to the browser).
const API_BASE_PATH = (process.env.API_BASE_PATH || 'http://localhost:8080').replace(/\/+$/, '');

export const dynamic = 'force-dynamic';

// Hop-by-hop headers (RFC 9110 §7.6.1) plus ones fetch/undici must set itself.
const HOP_BY_HOP = [
  'connection',
  'keep-alive',
  'proxy-authenticate',
  'proxy-authorization',
  'proxy-connection',
  'te',
  'trailer',
  'transfer-encoding',
  'upgrade',
];
const STRIP_REQUEST = [...HOP_BY_HOP, 'host', 'content-length', 'accept-encoding'];
// fetch transparently decompresses the body, so the original encoding/length no longer apply.
const STRIP_RESPONSE = [...HOP_BY_HOP, 'content-encoding', 'content-length'];

type RouteContext = { params: Promise<{ path: string[] }> };

async function proxy(req: NextRequest, { params }: RouteContext) {
  const { path } = await params;
  const target = `${API_BASE_PATH}/${path.map(encodeURIComponent).join('/')}${req.nextUrl.search}`;

  const headers = new Headers(req.headers);
  STRIP_REQUEST.forEach((h) => headers.delete(h));
  headers.set('x-forwarded-host', req.headers.get('host') ?? '');
  headers.set('x-forwarded-proto', req.nextUrl.protocol.replace(/:$/, ''));

  const hasBody = req.method !== 'GET' && req.method !== 'HEAD';

  let upstream: Response;
  try {
    upstream = await fetch(target, {
      method: req.method,
      headers,
      // Buffered so upstream gets a Content-Length; API payloads are small.
      body: hasBody ? await req.arrayBuffer() : undefined,
      redirect: 'manual',
      signal: req.signal,
      cache: 'no-store',
    });
  } catch (err) {
    if (req.signal.aborted) return new Response(null, { status: 499 });
    console.error(`[api proxy] ${req.method} ${target}:`, err);
    return Response.json({ msg: 'upstream unavailable' }, { status: 502 });
  }

  const resHeaders = new Headers(upstream.headers);
  STRIP_RESPONSE.forEach((h) => resHeaders.delete(h));

  return new Response(upstream.body, {
    status: upstream.status,
    statusText: upstream.statusText,
    headers: resHeaders,
  });
}

export {
  proxy as GET,
  proxy as POST,
  proxy as PUT,
  proxy as PATCH,
  proxy as DELETE,
  proxy as HEAD,
  proxy as OPTIONS,
};
