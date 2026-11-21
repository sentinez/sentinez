import type { MotionProps } from 'motion/react';

// Props of the rotating language word under the title: the old word leaves
// upwards, the next comes up from below.
export const LANG_SWAP = {
  initial: { y: 25, opacity: 0 },
  animate: { y: 0, opacity: 1 },
  exit: { y: -25, opacity: 0 },
  transition: { duration: 0.4, ease: 'easeInOut' },
} satisfies MotionProps;
