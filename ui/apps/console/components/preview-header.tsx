import { useFormatter, useTranslations } from 'next-intl';
import Language from './language';
import { Notifications } from './notifications';
import { BAR_HEIGHT } from '@/lib/layout-bars';
import { cn } from '@sentinez/ui/lib/utils';

type PreviewHeaderProps = {
  username: string;
  /** Slid up out of view: see useHideOnScroll */
  hidden?: boolean;
};

export default function PreviewHeader(props: PreviewHeaderProps) {
  const t = useTranslations('Header');
  const format = useFormatter();

  return (
    // stuck to the top of the window, slid up when hidden
    <div
      className={cn(
        'sticky top-0 z-20 flex w-full items-start bg-background text-foreground',
        'transition-transform duration-150 ease-out',
        props.hidden && '-translate-y-full',
      )}
      id="screenshot"
      data-theme-scope="preview"
    >
      <div
        className={cn(
          'flex w-full items-center border-b border-border bg-background px-3',
          BAR_HEIGHT,
        )}
      >
        <div className="flex w-full items-center justify-between gap-4">
          <div className="flex items-center gap-3">
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
              animate-gradient"
            >
              SENTINÉZ
            </span>
            <span className="hidden text-muted-foreground sm:inline">·</span>
            <p className="hidden text-xs text-muted-foreground sm:block">
              {t('welcome', { name: props.username })}
            </p>
          </div>

          <div className="flex min-w-0 items-center gap-5">
            {/* the date and, beside it, the notifications bell */}
            <div className="flex min-w-0 items-center gap-1">
              <span className="truncate text-xs text-muted-foreground transition-colors hover:text-foreground">
                {format.dateTime(new Date(), { dateStyle: 'full' })}
              </span>
              <Notifications />
            </div>

            <Language />
          </div>
        </div>
      </div>
    </div>
  );
}
