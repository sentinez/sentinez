'use client';

import { useEffect, useState } from 'react';

/**
 * Whether a bar at the top of the page is hidden: once the window is
 * scrolled down, and until it is scrolled up by any amount (or back to the
 * top), when it shows again at once. Read once a frame.
 */
export function useHideOnScroll(): boolean {
  const [hidden, setHidden] = useState(false);

  useEffect(() => {
    let last = window.scrollY;
    let frame = 0;

    const onScroll = () => {
      if (frame) return;
      frame = requestAnimationFrame(() => {
        frame = 0;
        const y = window.scrollY;
        const delta = y - last;
        last = y;
        if (delta < 0 || y <= 0) setHidden(false);
        else if (delta > 0) setHidden(true);
      });
    };

    window.addEventListener('scroll', onScroll, { passive: true });
    return () => {
      window.removeEventListener('scroll', onScroll);
      cancelAnimationFrame(frame);
    };
  }, []);

  return hidden;
}
