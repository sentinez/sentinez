import { Avatar, Style } from '@dicebear/core';
import definition from '@dicebear/styles/lorelei.json';

const style = new Style(definition);

/**
 * DiceBear "Croodles Neutral" avatar as an SVG data URI. The same seed always
 * draws the same face, so a user keeps theirs across sessions and devices.
 */
export function avatarOf(seed: string): string {
  return new Avatar(style, { seed }).toDataUri();
}
