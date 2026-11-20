'use client';

import Image from 'next/image';
import { motion } from 'motion/react';
import { ReactNode, useEffect, useState } from 'react';
import { Progress } from '@sentinez/ui/components/progress';
import { Field, FieldLabel } from '@sentinez/ui/components/field';
import { useTranslations } from 'next-intl';
import { Holiday, holidayById, holidayOn } from '@/components/holiday';

const _timeoutMillis = 900;
const _holidayTimeoutMillis = 1800;

/**
 * How long the splash stays. On a holiday it stays a bit longer: the
 * greeting fades in by 1s, then is left long enough to read.
 */
function timeoutOf(holiday?: Holiday) {
  return holiday ? _holidayTimeoutMillis : _timeoutMillis;
}

export default function LoadingEffect({ children }: { children: ReactNode }) {
  const [loading, setLoading] = useState(true);
  const [holiday, setHoliday] = useState<Holiday>();

  // The clock is read after mount, so the server render and the hydration
  // agree; `?holiday=<id>` previews a holiday.
  useEffect(() => {
    const preview = new URLSearchParams(window.location.search).get('holiday');
    const today = holidayById(preview) ?? holidayOn(new Date());
    setHoliday(today);

    const timer = setTimeout(() => setLoading(false), timeoutOf(today));

    return () => clearTimeout(timer);
  }, []);

  if (loading) {
    return <IsLoading holiday={holiday} />;
  }

  return <>{children}</>;
}

export function IsLoading({ holiday }: { holiday?: Holiday }) {
  const timeout = timeoutOf(holiday);
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
          <motion.div
            initial={{ opacity: 0, scale: 0.8, x: 50 }}
            animate={{
              opacity: 1,
              scale: 1,
              x: 0,
            }}
            transition={{
              opacity: {
                duration: 0.2,
              },
              scale: {
                duration: 0.2,
              },
              x: {
                delay: 0.25,
                duration: 0.4,
                ease: [0.22, 1, 0.36, 1],
              },
            }}
          >
            <Image src="/assets/sz.png" alt="Sentinez" width={70} height={70} loading="eager" />
          </motion.div>

          <motion.span
            initial={{ opacity: 0, x: 0 }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, scale: 0 }}

            transition={{
              delay: 0.5,
              duration: 0.4,
            }}
            className="text-5xl font-bold text-[#0504aa]"
          >
            Sentinéz
          </motion.span>
        </div>
        {holiday && (
          <motion.div
            initial={{ opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.6, duration: 0.4 }}
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
