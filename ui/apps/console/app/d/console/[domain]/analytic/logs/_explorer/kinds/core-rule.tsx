'use client';

import { useMemo } from 'react';
import { useTranslations } from 'next-intl';

import { Badge } from '@sentinez/ui/components/badge';
import { EventKind } from '@sentinez/proto/sentinez/apps/analytic/v1/event';
import { CoreRuleEvent } from '@sentinez/proto/sentinez/types/coreruleset/v1/event';
import { RuleBehavior } from '@sentinez/proto/sentinez/types/rule/v1/rule';
import { listCoreRuleEvents, type ApiOptions } from '@/lib/api/analytic';
import { STATUS_COLORS } from '../status';
import type { EventConfig, ListRequest } from '../types';
import {
  MonoCell,
  columnOf,
  ruleRequestColumns,
  ruleRequestFields,
  ruleRequestOf,
  ruleRequestRows,
} from './shared';

const list = async (req: ListRequest, options?: ApiOptions) =>
  (await listCoreRuleEvents(req, options)).events;

// The series of a behavior, the key of the histogram of the table
const seriesOf = (key: number) =>
  key === RuleBehavior.RULE_BEHAVIOR_DENY
    ? 'deny'
    : key === RuleBehavior.RULE_BEHAVIOR_ALLOW
      ? 'allow'
      : 'other';

const distinct = (values: string[]) => [...new Set(values)];

/** The explorer of the core ruleset (OWASP CRS) matches */
export function useCoreRuleConfig(): EventConfig<CoreRuleEvent> {
  const t = useTranslations('Logs');
  return useMemo(() => {
    const c = columnOf(t);
    const behaviorLabel = (b: RuleBehavior) => t(`series.${seriesOf(b)}`);
    const behaviorBadge = (e: CoreRuleEvent) => (
      <Badge variant={e.behavior === RuleBehavior.RULE_BEHAVIOR_DENY ? 'destructive' : 'secondary'}>
        {behaviorLabel(e.behavior)}
      </Badge>
    );
    return {
      kind: EventKind.EVENT_KIND_CORE_RULE,
      title: t('coreRule.title'),
      subtitle: t('coreRule.subtitle'),
      list,
      idOf: (e) => e.id,
      timeOf: (e) => e.matchedTime,
      columns: [
        {
          id: 'behavior',
          size: 100,
          header: c('behavior'),
          cell: ({ row }) => behaviorBadge(row.original),
        },
        {
          id: 'score',
          size: 70,
          header: () => <div className="text-right">{c('score')}</div>,
          cell: ({ row }) => (
            <div className="text-right font-mono text-xs tabular-nums">
              {row.original.totalScore}
            </div>
          ),
        },
        ...ruleRequestColumns<CoreRuleEvent>(c),
        {
          id: 'rules',
          size: 220,
          header: c('rules'),
          cell: ({ row }) => <MonoCell value={row.original.matchedRuleIds.join(', ')} />,
        },
      ],
      series: [
        { id: 'deny', label: t('series.deny'), color: STATUS_COLORS.s5xx },
        { id: 'allow', label: t('series.allow'), color: STATUS_COLORS.s2xx },
        { id: 'other', label: t('series.other'), color: STATUS_COLORS.other },
      ],
      seriesOf,
      fields: [
        {
          name: 'matched_rule_ids',
          label: c('rules'),
          get: (e) => e.matchedRuleIds.map(String),
        },
        { name: 'severities', label: c('severities'), get: (e) => e.severities },
        {
          name: 'behavior',
          label: c('behavior'),
          get: (e) => String(e.behavior),
          format: (v) => behaviorLabel(Number(v) as RuleBehavior),
        },
        { name: 'total_score', label: c('score'), get: (e) => String(e.totalScore) },
        ...ruleRequestFields<CoreRuleEvent>(c),
      ],
      requestOf: ruleRequestOf,
      text: (e) => [
        e.requestId,
        behaviorLabel(e.behavior),
        e.requestDomain,
        e.requestPath,
        e.requestQuery,
        e.requestIp,
        ...e.matchedRuleIds.map(String),
        ...e.severities,
        ...e.messages,
      ],
      detail: {
        title: (e) => (
          <>
            {behaviorBadge(e)}
            <span className="min-w-0 truncate font-mono text-sm">{e.requestPath}</span>
          </>
        ),
        rows: (e) => [
          [c('behavior'), behaviorLabel(e.behavior)],
          [c('score'), String(e.totalScore)],
          [c('rules'), e.matchedRuleIds.join(', ')],
          [c('severities'), distinct(e.severities).join(', ')],
          ...ruleRequestRows(c, e),
        ],
        // a message per matched rule, in the order of the rules
        sections: (e) => [
          {
            title: t('detail.messages'),
            pairs: e.messages.map((m, i): [string, string] => [
              [e.matchedRuleIds[i], e.severities[i]].filter(Boolean).join(' · ') || String(i + 1),
              m,
            ]),
          },
        ],
        toJSON: (e) => CoreRuleEvent.toJSON(e),
      },
    };
  }, [t]);
}
