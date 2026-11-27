/**
 * Heights shared by the bars of the page column and the blocks at the top of
 * the domain sidebar, so their bottom borders line up: the team switcher over
 * `PreviewHeader`, the title over `PageLayoutHeader`. The values are the
 * `--console-*` variables of `@sentinez/ui/custom.css`; full class names, so
 * that Tailwind finds them.
 */

/** The top bar of the page column, `PreviewHeader`, and the team block. */
export const BAR_HEIGHT = 'h-(--console-bar-height)';

/** The title block of the sidebar, over `PageLayoutHeader`. */
export const PAGE_HEADER_HEIGHT = 'h-(--console-page-header-height)';

/**
 * `PageLayoutHeader` itself: exactly that high from `sm` (beside the
 * sidebar, its border in line with the sidebar's); it may grow on a phone.
 */
export const PAGE_LAYOUT_HEADER_HEIGHT =
  'min-h-(--console-page-header-height) sm:h-(--console-page-header-height)';
