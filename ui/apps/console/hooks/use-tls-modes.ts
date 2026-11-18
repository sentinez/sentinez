'use client';

import { useMemo } from 'react';
import { useTranslations } from 'next-intl';
import { TlsMode } from '@sentinez/proto/sentinez/dmz/edge/v1/setting';
import { TLS_MODES } from '@/lib/type/edge';

/** The TLS modes as select options, and the label of one */
export function useTlsModes() {
  const t = useTranslations('ResourceSetting');

  return useMemo(() => {
    const options = TLS_MODES.map((m) => ({
      value: m.value as TlsMode,
      label: t(m.label),
      description: t(m.help),
    }));
    const labelOf = (mode: TlsMode) =>
      (options.find((o) => o.value === mode) ?? options[0]!).label;
    return { options, labelOf };
  }, [t]);
}
