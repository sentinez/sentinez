import type { HTMLAttributes, JSX } from 'react';

/** Props of LetterGlitch (React Bits); see letter-glitch.jsx for the ranges. */
export type LetterGlitchProps = HTMLAttributes<HTMLDivElement> & {
  /** Up to 8 colors glitching letters light up in. */
  glitchColors?: string[];
  /** Color of the letters at rest. */
  baseColor?: string;
  characters?: string;
  fontFamily?: string;
  fontSize?: number;
  spacing?: number;
  density?: number;
  baseOpacity?: number;
  glitchAmount?: number;
  speed?: number;
  flow?: number;
  smooth?: boolean;
  interactive?: boolean;
  cursorRadius?: number;
  clickEffect?: 'tear' | 'burst' | 'none';
  edgeFade?: number;
  centerFade?: number;
  intro?: boolean;
  /** Dark ink defaults, for light backgrounds. */
  lightMode?: boolean;
  backgroundColor?: string;
  dpr?: number;
};

declare function LetterGlitch(props: LetterGlitchProps): JSX.Element;

export default LetterGlitch;
