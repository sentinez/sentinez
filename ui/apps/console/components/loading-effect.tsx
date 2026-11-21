'use client';

import Image from 'next/image';
import { motion } from 'motion/react';
import { ReactNode, useEffect, useState } from 'react';
import { Progress } from '@sentinez/ui/components/progress';
import { Field, FieldLabel } from '@sentinez/ui/components/field';
import { useTranslations } from 'next-intl';
import { Holiday, holidayById, holidayOn } from '@/components/holiday';
import { splashMillis } from '@/lib/delays';
import { SPLASH_GREETING, SPLASH_LOGO, SPLASH_NAME } from '@/lib/motion';

export default function LoadingEffect({ children }: { children: ReactNode }) {
  const [loading, setLoading] = useState(true);
  const [holiday, setHoliday] = useState<Holiday>();

  // The clock is read after mount, so the server render and the hydration
  // agree; `?holiday=<id>` previews a holiday.
  useEffect(() => {
    const preview = new URLSearchParams(window.location.search).get('holiday');
    const today = holidayById(preview) ?? holidayOn(new Date());
    setHoliday(today);

    const timer = setTimeout(() => setLoading(false), splashMillis(!!today));

    return () => clearTimeout(timer);
  }, []);

  if (loading) {
    return <IsLoading holiday={holiday} />;
  }

  return <>{children}</>;
}

export function IsLoading({ holiday }: { holiday?: Holiday }) {
  const timeout = splashMillis(!!holiday);
  const t = useTranslations('Loading');

  useEffect(() => {
    const start = performance.now();
    // setInterval (not requestAnimationFrame) so progress still advances
    // when the tab is hidden or the browser throttles animation frames.
    const id = setInterval(() => {
      const elapsed = performance.now() - start;
      const next = Math.min((elapsed / timeout) * 99, 99);

      setProgress(next);

      if (next >= 99) clearInterval(id);
    }, 50);

    return () => clearInterval(id);
  }, [timeout]);

  const [progress, setProgress] = useState(10);

  return (
    <div className="fixed inset-0 flex items-center justify-center bg-background z-50 font-['Phudu',sans-serif]">
      <div className="relative flex flex-col items-center gap-4">
        <div className="flex items-center">
          <motion.div {...SPLASH_LOGO}>
            <Image src="/assets/sz.png" alt="Sentinez" width={70} height={70} loading="eager" />
          </motion.div>

          <motion.span
            {...SPLASH_NAME}
            className="text-5xl font-bold text-[#0504aa]"
          >
            Sentinéz
          </motion.span>
        </div>
        {holiday && (
          <motion.div
            {...SPLASH_GREETING}
            className={`text-lg font-semibold ${holiday.color}`}
          >
            {t(`holiday.${holiday.id}`)}
          </motion.div>
        )}
        {/* <div className="font-medium font-mono text-foreground">The Sentinéz is loading...</div> */}
        <Field className="w-full max-w-sm">
          <FieldLabel htmlFor="progress-upload">
            <span className="font-medium font-sans text-foreground">{t('text')}</span>
            {/* <span className="ml-auto">{progress}%</span> */}
          </FieldLabel>
          <Progress className="h-1" value={progress} id="progress-upload" />
        </Field>
      </div>
    </div>
  );
}
