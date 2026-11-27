/**
 * Heights shared by the bars of the page column and the blocks at the top of
 * the domain sidebar, so their bottom borders line up: the team switcher over
 * `PreviewHeader` + `BreadcrumbBar`, the title and search over
 * `PageLayoutHeader`. The values are the `--console-*` variables of
 * `@sentinez/ui/custom.css`; full class names, so that Tailwind finds them.
 */

/** One bar of the page column: `PreviewHeader`, `BreadcrumbBar`. */
export const BAR_HEIGHT = 'h-(--console-bar-height)';

/** The two bars above `PageLayoutHeader`, together. */
export const TOP_BARS_HEIGHT = 'h-[calc(2*var(--console-bar-height))]';

/** The title and search block of the sidebar, over `PageLayoutHeader`. */
export const PAGE_HEADER_HEIGHT = 'h-(--console-page-header-height)';

/** `PageLayoutHeader` itself: it may grow on a narrow screen. */
export const PAGE_HEADER_MIN_HEIGHT = 'min-h-(--console-page-header-height)';
