'use client';

import Image from 'next/image';
import { motion } from 'motion/react';
import { ReactNode, useEffect, useState } from 'react';
import { Progress } from '@sentinez/ui/components/progress';
import { Field, FieldLabel } from '@sentinez/ui/components/field';
import { useTranslations } from 'next-intl';

export default function LoadingEffect({ children }: { children: ReactNode }) {
  const [loading, setLoading] = useState(true);
  useEffect(() => {
    const timer = setTimeout(() => setLoading(false), 1500);

    return () => clearTimeout(timer);
  }, []);

  if (loading) return <IsLoading timeout={1000} />;

  return <>{children}</>;
}

export function IsLoading({ timeout }: { timeout: number }) {
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

  const [progress, setProgress] = useState(0);
  return (
    <div className="fixed inset-0 flex items-center justify-center bg-background z-50 [font-family:'Phudu',sans-serif]">
      <div className="flex flex-col items-center gap-4">
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
                duration: 0.25,
              },
              scale: {
                duration: 0.25,
              },
              x: {
                delay: 0.3,
                duration: 0.45,
                ease: [0.22, 1, 0.36, 1],
              },
            }}
          >
            <Image src="/assets/sntz.png" alt="Sentinez" width={70} height={70} loading="eager" />
          </motion.div>

          <motion.span
            initial={{ opacity: 0, x: 0 }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, scale: 0 }}

            transition={{
              delay: 0.75,
              duration: 0.4,
            }}
            className="text-5xl font-bold text-[#0504aa]"
          >
            Sentinéz
          </motion.span>
        </div>
        {/* <div className="font-medium font-mono text-foreground">The Sentinéz is loading...</div> */}
        <Field className="w-full max-w-sm">
          <FieldLabel htmlFor="progress-upload">
            <span className="font-medium font-sans text-foreground">{t('text')}</span>
            {/* <span className="ml-auto">{progress}%</span> */}
          </FieldLabel>
          <Progress value={progress} id="progress-upload" />
        </Field>
      </div>
    </div>
  );
}
