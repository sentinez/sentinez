import { useFormatter, useTranslations } from 'next-intl';
import Language from './language';
import Theme from './theme';
import { Notifications } from './notifications';
import { NavSearch } from './nav-search';
import { Separator } from '@sentinez/ui/components/separator';
import { SidebarTrigger } from '@sentinez/ui/components/sidebar';
import { BAR_HEIGHT } from '@/lib/layout-bars';
import { cn } from '@sentinez/ui/lib/utils';

type PreviewHeaderProps = {
  username: string;
};

export default function PreviewHeader(props: PreviewHeaderProps) {
  const t = useTranslations('Header');
  const format = useFormatter();

  return (
    // stuck to the top of the window: its border stays in line with the team
    // block of the sidebar. The padding of PageLayoutHeader, and the sidebar
    // trigger centred in a slot as wide as its back button (w-9): the two
    // icons line up
    <div
      className={cn(
        'sticky top-0 z-20 flex w-full items-center gap-1 border-b bg-background px-1 text-foreground md:px-2',
        BAR_HEIGHT,
      )}
      id="screenshot"
      data-theme-scope="preview"
    >
      <div className="grid w-9 shrink-0 place-items-center">
        <SidebarTrigger className="size-8 cursor-pointer" />
      </div>
      <Separator orientation="vertical" className="data-vertical:h-4 data-vertical:self-center" />
      {/* laid out on its own width (@container), not the window's: the
          sidebars beside it take a part of the window. The brand and the
          icons never shrink; the date and the welcome show only when they
          fit, and truncate; the search is an icon until there is room for
          its box */}
      <div className="@container flex min-w-0 flex-1 items-center justify-between gap-3 px-2">
        <div className="flex min-w-0 items-center gap-3">
          <span
            className="[font-family:'Phudu',sans-serif]
              font-bold
              text-xs 
              bg-gradient-to-r
              from-cyan-400 
              via-blue-500
              to-purple-600
              bg-[length:200%_auto]
              bg-clip-text
              text-transparent
              animate-gradient
              shrink-0"
          >
            SENTINÉZ
          </span>
          <span className="hidden truncate text-xs text-muted-foreground transition-colors hover:text-foreground @3xl:inline">
            {format.dateTime(new Date(), { dateStyle: 'full' })}
          </span>
        </div>

        <div className="flex min-w-0 items-center gap-3">
          <p className="hidden truncate text-xs text-muted-foreground @5xl:block">
            {t('welcome', { name: props.username })}
          </p>
          {/* finds a page of the console navigation; beside it, the bell,
              the theme and the language */}
          <div className="flex shrink-0 items-center gap-1">
            <NavSearch />
            <Notifications />
            <Theme />
            <Language />
          </div>
        </div>
      </div>
    </div>
  );
}
