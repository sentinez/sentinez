'use client';

import { useEffect, useState } from 'react';
import type { ConditionSchema } from '@sentinez/proto/sentinez/types/rule/v1/rule';
import { getConditionSchema } from '@/lib/api/security';

// The schema is the same for every rule and every user: fetched once per
// page load, shared by every editor. A failure is kept too: the editors
// then offer everything and the server still validates the rule, instead of
// each of them asking again.
let _schema: Promise<ConditionSchema | undefined> | undefined;

function loadSchema(): Promise<ConditionSchema | undefined> {
  _schema ??= getConditionSchema()
    .then((res) => res.schema)
    .catch(() => undefined);
  return _schema;
}

/**
 * The conditions a rule expression accepts, undefined while it loads or
 * when it could not be fetched: the server still validates the rule.
 */
export function useConditionSchema(): ConditionSchema | undefined {
  const [schema, setSchema] = useState<ConditionSchema>();

  useEffect(() => {
    let active = true;
    loadSchema().then((s) => {
      if (active) setSchema(s);
    });
    return () => {
      active = false;
    };
  }, []);

  return schema;
}
