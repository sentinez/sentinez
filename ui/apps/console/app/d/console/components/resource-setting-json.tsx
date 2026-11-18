'use client';

import { useState } from 'react';
import { useTranslations } from 'next-intl';

import { Resource as EdgeResource } from '@sentinez/proto/sentinez/dmz/edge/v1/setting';
import { CodeBlock } from '@/components/code-block';

/** The edge setting of a resource as raw JSON, folded until asked for */
export function ResourceSettingJson({ resource }: { resource?: EdgeResource }) {
  const t = useTranslations('ResourceSetting');
  // Serialized only once unfolded, not on every render of the page
  const [open, setOpen] = useState(false);

  if (!resource) {
    return null;
  }

  return (
    <details onToggle={(e) => setOpen(e.currentTarget.open)}>
      <summary className="w-fit cursor-pointer rounded-md text-sm text-muted-foreground hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring">
        {t('showJson')}
      </summary>
      {open && <CodeBlock className="mt-2" value={EdgeResource.toJSON(resource)} />}
    </details>
  );
}
