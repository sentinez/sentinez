import { Avatar, Style } from '@dicebear/core';
import definition from '@dicebear/styles/lorelei.json';

const style = new Style(definition);

const INK = '1b1b3a';

/**
 * DiceBear "Bold Pop" preset for Lorelei: eight backgrounds and eight hair
 * colors on a pale face with dark ink linework. The seed picks the pair.
 */
const BOLD_POP = {
  backgroundColor: [
    'ff5d8f',
    'ffb703',
    '43aa8b',
    '4d96ff',
    'b57bff',
    'ff8c42',
    '2ec4b6',
    'f15bb5',
  ],
  skinColor: ['fffdfb'],
  hairColor: [
    INK,
    'd90429',
    '0466c8',
    '007f5f',
    '7b2cbf',
    'f77f00',
    '006d77',
    '9d0208',
  ],
  eyebrowsColor: [INK],
  eyesColor: [INK],
  noseColor: [INK],
  mouthColor: [INK],
  frecklesColor: [INK],
  glassesColor: [INK],
  earringsColor: [INK],
  hairAccessoriesColor: [INK],
};

/**
 * DiceBear "Pastel Wall" preset for Lorelei: five soft backgrounds, the
 * drawing itself untouched. The seed picks which one an avatar stands on.
 */
const PASTEL_WALL = {
  backgroundColor: ['ffe3ea', 'e3edff', 'e2f5e9', 'fdf1d4', 'efe6ff'],
};

const PRESETS = {
  'bold-pop': BOLD_POP,
  'pastel-wall': PASTEL_WALL,
};

export type AvatarPreset = keyof typeof PRESETS;

/**
 * DiceBear "Lorelei" avatar as an SVG data URI. The same seed always draws
 * the same face, so a user keeps theirs across sessions and devices.
 */
export function avatarOf(
  seed: string,
  preset: AvatarPreset = 'bold-pop',
): string {
  return new Avatar(style, { ...PRESETS[preset], seed }).toDataUri();
}
