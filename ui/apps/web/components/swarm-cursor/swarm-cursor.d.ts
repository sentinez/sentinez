import type { HTMLAttributes, JSX } from 'react';

/** Props of SwarmCursor (React Bits); see swarm-cursor.jsx for the ranges. */
export type SwarmCursorProps = HTMLAttributes<HTMLDivElement> & {
  /** Number of particles, up to 2000. */
  count?: number;
  size?: number;
  shape?: 'dot' | 'streak' | 'square';
  color?: string;
  /** Color particles take on when they rush and when clicked. */
  accentColor?: string;
  spread?: number;
  speed?: number;
  swirl?: number;
  tail?: number;
  trail?: number;
  wander?: number;
  depth?: number;
  /** Elements the swarm pours around while hovered; '' turns it off. */
  targets?: string;
  idle?: 'roam' | 'gather' | 'hide';
  clickEffect?: 'burst' | 'pulse' | 'none';
  hideCursor?: boolean;
  intro?: boolean;
  enabled?: boolean;
  /** Near black particles, for light backgrounds. */
  lightMode?: boolean;
  dpr?: number;
};

declare function SwarmCursor(props: SwarmCursorProps): JSX.Element;

export default SwarmCursor;
