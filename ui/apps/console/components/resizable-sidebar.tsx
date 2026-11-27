'use client';

import * as React from 'react';
import { SidebarProvider, SidebarRail, useSidebar } from '@sentinez/ui/components/sidebar';
import { cn } from '@sentinez/ui/lib/utils';

import { SIDEBAR_WIDTH_COOKIE } from '@/lib/layout-bars';

const SIDEBAR_WIDTH_MAX_AGE = 60 * 60 * 24 * 7;
// how far the pointer moves before a press on the rail is a drag, not a click
const RESIZE_THRESHOLD = 3;

type SidebarWidth = {
  width: number;
  setWidth: (width: number) => void;
  clampWidth: (width: number) => number;
};

const SidebarWidthContext = React.createContext<SidebarWidth | null>(null);

type ResizableSidebarProviderProps = React.ComponentProps<typeof SidebarProvider> & {
  /** The width in px to start with */
  defaultWidth: number;
  minWidth?: number;
  maxWidth?: number;
};

/**
 * A SidebarProvider whose sidebar is resized by dragging ResizableSidebarRail.
 * The width is kept here, out of the shadcn component (which `shadcn add`
 * overwrites), set as `--sidebar-width` and saved in SIDEBAR_WIDTH_COOKIE.
 */
export function ResizableSidebarProvider({
  defaultWidth,
  minWidth = 200,
  maxWidth = 480,
  className,
  style,
  ...props
}: ResizableSidebarProviderProps) {
  const clampWidth = React.useCallback(
    (value: number) => Math.min(maxWidth, Math.max(minWidth, value)),
    [minWidth, maxWidth],
  );
  const [width, setWidthState] = React.useState(() => clampWidth(defaultWidth));
  const setWidth = React.useCallback(
    (value: number) => {
      const clamped = clampWidth(value);
      setWidthState(clamped);
      document.cookie = `${SIDEBAR_WIDTH_COOKIE}=${clamped}; path=/; max-age=${SIDEBAR_WIDTH_MAX_AGE}`;
    },
    [clampWidth],
  );
  const value = React.useMemo(
    () => ({ width, setWidth, clampWidth }),
    [width, setWidth, clampWidth],
  );

  return (
    <SidebarWidthContext.Provider value={value}>
      <SidebarProvider
        style={{ '--sidebar-width': `${width}px`, ...style } as React.CSSProperties}
        className={cn(
          // data-resizing is set by ResizableSidebarRail while dragging: the
          // sidebar follows the pointer instead of easing after it
          'data-resizing:cursor-col-resize data-resizing:select-none',
          'data-resizing:**:data-[slot=sidebar-gap]:transition-none',
          'data-resizing:**:data-[slot=sidebar-container]:transition-none',
          className,
        )}
        {...props}
      />
    </SidebarWidthContext.Provider>
  );
}

type Drag = {
  wrapper: HTMLElement;
  startX: number;
  startWidth: number;
  direction: number;
  width: number;
  moved: boolean;
};

/**
 * The SidebarRail of a ResizableSidebarProvider: a click toggles the sidebar,
 * a drag resizes it while expanded. While dragging it writes `--sidebar-width`
 * straight to the wrapper, so nothing re-renders; the width is committed when
 * the drag ends.
 */
export function ResizableSidebarRail({
  className,
  ...props
}: React.ComponentProps<typeof SidebarRail>) {
  const { toggleSidebar, state, isMobile } = useSidebar();
  const sizing = React.useContext(SidebarWidthContext);
  const drag = React.useRef<Drag | null>(null);
  const resizable = !!sizing && state === 'expanded' && !isMobile;

  const handlePointerDown = (event: React.PointerEvent<HTMLButtonElement>) => {
    const target = event.currentTarget;
    const wrapper = target.closest<HTMLElement>('[data-slot=sidebar-wrapper]');
    if (!resizable || !wrapper || event.button !== 0) return;
    target.setPointerCapture(event.pointerId);
    const side = target.closest('[data-side]')?.getAttribute('data-side');
    drag.current = {
      wrapper,
      startX: event.clientX,
      startWidth: sizing.width,
      direction: side === 'right' ? -1 : 1,
      width: sizing.width,
      moved: false,
    };
  };

  const handlePointerMove = (event: React.PointerEvent<HTMLButtonElement>) => {
    const current = drag.current;
    if (!current || !sizing) return;
    const delta = (event.clientX - current.startX) * current.direction;
    if (!current.moved) {
      if (Math.abs(delta) < RESIZE_THRESHOLD) return;
      current.moved = true;
      current.wrapper.dataset.resizing = 'true';
    }
    current.width = sizing.clampWidth(current.startWidth + delta);
    current.wrapper.style.setProperty('--sidebar-width', `${current.width}px`);
  };

  const handlePointerEnd = () => {
    const current = drag.current;
    if (!current?.moved) {
      drag.current = null;
      return;
    }
    delete current.wrapper.dataset.resizing;
    sizing?.setWidth(current.width);
    // drag.current stays set so the click that follows is not a toggle
  };

  const handleClick = () => {
    if (drag.current?.moved) {
      drag.current = null;
      return;
    }
    toggleSidebar();
  };

  return (
    <SidebarRail
      onClick={handleClick}
      onPointerDown={handlePointerDown}
      onPointerMove={handlePointerMove}
      onPointerUp={handlePointerEnd}
      onPointerCancel={handlePointerEnd}
      title={resizable ? 'Drag to resize, click to toggle' : 'Toggle Sidebar'}
      className={cn('touch-none', resizable && 'cursor-col-resize!', className)}
      {...props}
    />
  );
}
