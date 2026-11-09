'use client';

import { useState } from 'react';
import { useTranslations } from 'next-intl';

import { Button } from '@sentinez/ui/components/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@sentinez/ui/components/card';
import { Server } from '@sentinez/proto/sentinez/dmz/edge/v1/setting';

/** The edge setting of a resource as raw JSON, hidden until asked for */
export function ResourceSettingJson({ server }: { server?: Server }) {
  const [showJson, setShowJson] = useState(false);
  const t = useTranslations('Resource');

  return (
    <Card>
      <CardHeader className="flex flex-row items-start justify-between gap-4">
        <div className="grid gap-1.5">
          <CardTitle>{t('settings')}</CardTitle>
          <CardDescription>{t('settingsDescription')}</CardDescription>
        </div>
        {server && (
          <Button
            variant="outline"
            size="sm"
            aria-expanded={showJson}
            onClick={() => setShowJson((v) => !v)}
          >
            {showJson ? t('hideJson') : t('showJson')}
          </Button>
        )}
      </CardHeader>
      <CardContent>
        {!server ? (
          <div className="text-sm text-muted-foreground p-4 border rounded-md">
            {t('noSettings')}
          </div>
        ) : showJson ? (
          <pre className="bg-muted text-foreground p-4 rounded-md overflow-x-auto text-xs min-h-[300px]">
            {JSON.stringify(Server.toJSON(server), null, 2)}
          </pre>
        ) : (
          <div className="text-sm text-muted-foreground">{t('hiddenJson')}</div>
        )}
      </CardContent>
    </Card>
  );
}
