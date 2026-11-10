import { BalanceStrategy, ProxyProtocol } from '@sentinez/proto/sentinez/dmz/edge/v1/setting';

// Helpers of the edge setting (Server, Location, Upstream): the labels of
// its enums and the patterns its protobuf validation (buf.validate) holds.

/** The protocol of an upstream; the edge reads an unspecified one as HTTPS */
export function protocolOf(
  protocol: ProxyProtocol,
): ProxyProtocol.PROXY_PROTOCOL_HTTP | ProxyProtocol.PROXY_PROTOCOL_HTTPS {
  return protocol === ProxyProtocol.PROXY_PROTOCOL_HTTP
    ? ProxyProtocol.PROXY_PROTOCOL_HTTP
    : ProxyProtocol.PROXY_PROTOCOL_HTTPS;
}

/** The URL scheme of an upstream protocol */
export function schemeOf(protocol: ProxyProtocol): 'http' | 'https' {
  return protocol === ProxyProtocol.PROXY_PROTOCOL_HTTP ? 'http' : 'https';
}

/** The label of the protocols, in the order a select lists them */
export const PROTOCOL_LABEL = [
  { value: ProxyProtocol.PROXY_PROTOCOL_HTTPS, label: 'HTTPS' },
  { value: ProxyProtocol.PROXY_PROTOCOL_HTTP, label: 'HTTP' },
] as const;

/** Keys of the `ResourceSetting` messages that label a balance strategy */
type BalanceLabelKey = 'balanceRoundRobin' | 'balanceDefault';

/** The label key of a balance strategy; an unspecified one is the default */
export function balanceLabelKey(strategy: BalanceStrategy): BalanceLabelKey {
  return strategy === BalanceStrategy.BALANCE_STRATEGY_ROUND_ROBIN
    ? 'balanceRoundRobin'
    : 'balanceDefault';
}

/** The balance strategies, in the order a select lists them */
export const BALANCE_STRATEGIES = [
  BalanceStrategy.BALANCE_STRATEGY_ROUND_ROBIN,
  BalanceStrategy.BALANCE_STRATEGY_UNSPECIFIED,
] as const;

/** A "Name: value" header line; the value follows the first colon */
export type HeaderLine = { line: string; name: string; value: string };

/**
 * The "Name: value" lines of headers, blank ones left out: the format of
 * the headers of a location form and of the edge events.
 */
export function headerLinesOf(text: string): HeaderLine[] {
  return text
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean)
    .map((line) => {
      const at = line.indexOf(':');
      return at < 0
        ? { line, name: line, value: '' }
        : { line, name: line.slice(0, at).trim(), value: line.slice(at + 1).trim() };
    });
}

// Patterns from setting.proto (buf.validate): server.name,
// server.resource_domain, location.location / proxy_rewrite, upstream.server
export const NAME_RE = /^[a-zA-Z0-9-]+$/;
export const DOMAIN_RE = /^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,63}$/;
export const LOCATION_RE = /^\/.*$/;
export const UPSTREAM_RE =
  /^(((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3})|(([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}))$/;

/** A request header name: an RFC 7230 token, as the edge accepts it */
export const HEADER_NAME_RE = /^[!#$%&'*+\-.^_`|~0-9A-Za-z]+$/;
