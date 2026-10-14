'use client';

import { Avatar, AvatarFallback, AvatarImage } from '@sentinez/ui/components/avatar';
import { cn } from '@sentinez/ui/lib/utils';
import { avatarOf } from '@/lib/avatar';

// "Alice Nguyen" -> "AN"
const initials = (name: string) =>
  name
    .split(/\s+/)
    .slice(0, 2)
    .map((w) => w[0]?.toUpperCase() ?? '')
    .join('');

/**
 * Avatar of a member's user. `seed` is the full name, or the user id when the
 * name is unknown, like the user menu, so a user has the same face everywhere.
 */
export function MemberAvatar({
  name,
  seed,
  className,
}: {
  name: string;
  seed: string;
  className?: string;
}) {
  return (
    <Avatar className={cn('h-8 w-8 rounded-lg', className)}>
      <AvatarImage src={avatarOf(seed)} alt={name} />
      <AvatarFallback className="rounded-lg">{initials(name)}</AvatarFallback>
    </Avatar>
  );
}
