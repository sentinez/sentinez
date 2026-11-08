'use client';

import { useMemo } from 'react';
import { useTranslations } from 'next-intl';

import { Badge } from '@sentinez/ui/components/badge';
import { EventKind } from '@sentinez/proto/sentinez/apps/analytic/v1/event';
import { ActionType } from '@sentinez/proto/sentinez/types/rule/v1/rule';
import { SecRuleEvent } from '@sentinez/proto/sentinez/types/secrule/v1/event';
import { useSecurityOptions } from '@/hooks/use-security-options';
import { listRateLimitEvents, listSecRuleEvents, type ApiOptions } from '@/lib/api/analytic';
import { ACTION_TYPE_BADGE_VARIANT } from '@/lib/type/security';
import type { EventConfig, ListRequest } from '../types';
import { useDomain } from '@/hooks/use-domain';
import { STATUS_COLORS } from '../status';
import {
  RuleLinks,
  columnOf,
  ruleRequestColumns,
  ruleRequestFields,
  ruleRequestOf,
  ruleRequestRows,
} from './shared';

// SecRuleKind the two tables of secrulepb.SecRuleEvent
export type SecRuleKind = 'secRule' | 'rateLimit';

const KINDS = {
  secRule: {
    kind: EventKind.EVENT_KIND_SEC_RULE,
    list: async (req: ListRequest, options?: ApiOptions) =>
      (await listSecRuleEvents(req, options)).events,
    // the security page of the matched rules
    rulePage: 'sec-rule',
    rulesColumn: 'secRules',
    color: STATUS_COLORS.s5xx,
  },
  rateLimit: {
    kind: EventKind.EVENT_KIND_RATE_LIMIT,
    list: async (req: ListRequest, options?: ApiOptions) =>
      (await listRateLimitEvents(req, options)).events,
    rulePage: 'rate-limiter',
    rulesColumn: 'rateLimiters',
    color: STATUS_COLORS.s4xx,
  },
} as const;

/** The explorer of the SecRule matches, or of the rate limit rejections */
export function useSecRuleConfig(which: SecRuleKind): EventConfig<SecRuleEvent> {
  const t = useTranslations('Logs');
  const { actionTypeLabel } = useSecurityOptions();
  const domain = useDomain();
  return useMemo(() => {
    const c = columnOf(t);
    const spec = KINDS[which];
    const ruleHref = (id: string) =>
      `/d/console/${domain}/security/${spec.rulePage}/${encodeURIComponent(id)}`;
    const typeOf = (e: SecRuleEvent) => e.action?.type ?? ActionType.ACTION_TYPE_UNSPECIFIED;
    const actionBadge = (e: SecRuleEvent) => (
      <Badge variant={ACTION_TYPE_BADGE_VARIANT[typeOf(e)]}>{actionTypeLabel(typeOf(e))}</Badge>
    );
    return {
      kind: spec.kind,
      title: t(`${which}.title`),
      subtitle: t(`${which}.subtitle`),
      list: spec.list,
      idOf: (e) => e.id,
      timeOf: (e) => e.matchedTime,
      columns: [
        {
          id: 'action',
          size: 110,
          header: c('action'),
          cell: ({ row }) => actionBadge(row.original),
        },
        ...ruleRequestColumns<SecRuleEvent>(c),
        {
          id: 'rules',
          size: 260,
          header: c(spec.rulesColumn),
          cell: ({ row }) => <RuleLinks ids={row.original.matchedSecruleIds} href={ruleHref} />,
        },
      ],
      series: [{ id: 'events', label: t('series.events'), color: spec.color }],
      // the histogram of a rule table has no key
      seriesOf: () => 'events',
      fields: [
        {
          name: 'matched_secrule_ids',
          label: c(spec.rulesColumn),
          get: (e) => e.matchedSecruleIds,
        },
        {
          name: 'action',
          label: c('action'),
          get: (e) => String(typeOf(e)),
          format: (v) => actionTypeLabel(Number(v) as ActionType),
        },
        ...ruleRequestFields<SecRuleEvent>(c),
      ],
      requestOf: ruleRequestOf,
      text: (e) => [
        e.requestId,
        actionTypeLabel(typeOf(e)),
        e.requestDomain,
        e.requestPath,
        e.requestQuery,
        e.requestIp,
        ...e.matchedSecruleIds,
        ...e.messages,
      ],
      detail: {
        title: (e) => (
          <>
            {actionBadge(e)}
            <span className="min-w-0 truncate font-mono text-sm">{e.requestPath}</span>
          </>
        ),
        rows: (e) => [
          [c('action'), actionTypeLabel(typeOf(e))],
          [
            c(spec.rulesColumn),
            <RuleLinks key="rules" ids={e.matchedSecruleIds} href={ruleHref} />,
          ],
          ...ruleRequestRows(c, e),
        ],
        sections: (e) => [
          {
            title: t('detail.messages'),
            pairs: e.messages.map((m, i): [string, string] => [String(i + 1), m]),
          },
          { title: t('detail.tags'), pairs: Object.entries(e.tags) },
        ],
        toJSON: (e) => SecRuleEvent.toJSON(e),
      },
    };
  }, [t, which, actionTypeLabel, domain]);
}
