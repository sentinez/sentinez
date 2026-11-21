'use client';

import * as React from 'react';
import { MotionConfig } from 'motion/react';
import { ThemeProvider as NextThemesProvider } from 'next-themes';

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    // reducedMotion="user": follow the OS "reduce motion" setting
    <MotionConfig reducedMotion="user">
      <NextThemesProvider
        attribute="class"
        defaultTheme="light"
        enableSystem
        disableTransitionOnChange
        enableColorScheme
      >
        {children}
      </NextThemesProvider>
    </MotionConfig>
  );
}
