'use client';

import { useState } from 'react';
import { useTranslations } from 'next-intl';

import { Server } from '@sentinez/proto/sentinez/dmz/edge/v1/setting';

/** The edge setting of a resource as raw JSON, folded until asked for */
export function ResourceSettingJson({ server }: { server?: Server }) {
  const t = useTranslations('ResourceSetting');
  // Serialized only once unfolded, not on every render of the page
  const [open, setOpen] = useState(false);

  if (!server) {
    return null;
  }

  return (
    <details onToggle={(e) => setOpen(e.currentTarget.open)}>
      <summary className="w-fit cursor-pointer rounded-md text-sm text-muted-foreground hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring">
        {t('showJson')}
      </summary>
      {open && (
        <pre className="bg-muted text-foreground mt-2 max-h-96 overflow-auto rounded-md p-3 text-xs">
          {JSON.stringify(Server.toJSON(server), null, 2)}
        </pre>
      )}
    </details>
  );
}
