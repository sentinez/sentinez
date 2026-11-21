'use client';

import { type ComponentProps, useState } from 'react';
import Link from 'next/link';

/**
 * A Link that prefetches once the user shows intent (pointer or keyboard
 * focus), not when it scrolls into view. Meant for links in table rows: each
 * row is its own url, so a viewport prefetch would send one request per row
 * while the page loads. The "Hover-triggered prefetch" pattern of the Next
 * docs: `prefetch={null}` hands the link back its default prefetch.
 */
export function HoverLink({
  onPointerEnter,
  onFocus,
  ...props
}: Omit<ComponentProps<typeof Link>, 'prefetch'>) {
  const [active, setActive] = useState(false);
  return (
    <Link
      {...props}
      prefetch={active ? null : false}
      onPointerEnter={(e) => {
        setActive(true);
        onPointerEnter?.(e);
      }}
      onFocus={(e) => {
        setActive(true);
        onFocus?.(e);
      }}
    />
  );
}
