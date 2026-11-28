'use client';

import { useTheme } from 'next-themes';
import type { ComponentProps } from 'react';
import LetterGlitch from './letter-glitch';

// ThemedLetterGlitch a LetterGlitch in the colors of the theme: dark letters
// unless the theme is dark. Nothing renders until the theme is known (on the
// client), so the dark theme does not start light.
export function ThemedLetterGlitch(props: Omit<ComponentProps<typeof LetterGlitch>, 'lightMode'>) {
  const { resolvedTheme } = useTheme();
  if (!resolvedTheme) return null;
  return <LetterGlitch {...props} lightMode={resolvedTheme !== 'dark'} />;
}
