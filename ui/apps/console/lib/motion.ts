import type { MotionProps } from 'motion/react';

// Props of the rotating language word under the title (auth page): the old
// word leaves upwards, the next comes up from below.
export const LANG_SWAP = {
  initial: { y: 25, opacity: 0 },
  animate: { y: 0, opacity: 1 },
  exit: { y: -25, opacity: 0 },
  transition: { duration: 0.4, ease: 'easeInOut' },
} satisfies MotionProps;

// Splash (components/loading-effect), in seconds: the logo slides in, then
// the name, then the holiday greeting. The beat must end before SPLASH_MILLIS
// (lib/delays), the greeting before the longer SPLASH_HOLIDAY_MILLIS, with a
// short hold so the logo is seen whole before the splash goes.
const FADE = 0.2;
const MOVE = 0.35;
const LOGO_AT = 0.1;
const NAME_AT = 0.3;
const GREETING_AT = 0.5;

export const SPLASH_LOGO = {
  initial: { opacity: 0, scale: 0.8, x: 50 },
  animate: { opacity: 1, scale: 1, x: 0 },
  transition: {
    opacity: { duration: FADE },
    scale: { duration: FADE },
    x: { delay: LOGO_AT, duration: MOVE, ease: [0.22, 1, 0.36, 1] },
  },
} satisfies MotionProps;

export const SPLASH_NAME = {
  initial: { opacity: 0 },
  animate: { opacity: 1 },
  exit: { opacity: 0, scale: 0 },
  transition: { delay: NAME_AT, duration: MOVE },
} satisfies MotionProps;

export const SPLASH_GREETING = {
  initial: { opacity: 0, y: 8 },
  animate: { opacity: 1, y: 0 },
  transition: { delay: GREETING_AT, duration: MOVE },
} satisfies MotionProps;
