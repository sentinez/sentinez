import { useFormatter, useTranslations } from 'next-intl';
import Language from './language';
import { BAR_HEIGHT } from '@/lib/layout-bars';
import { cn } from '@sentinez/ui/lib/utils';

type PreviewHeaderProps = {
  username: string;
};

export default function PreviewHeader(props: PreviewHeaderProps) {
  const t = useTranslations('Header');
  const format = useFormatter();

  return (
    <div
      className="flex w-full items-start bg-background text-foreground"
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
            <span className="truncate text-xs text-muted-foreground transition-colors hover:text-foreground">
              {format.dateTime(new Date(), { dateStyle: 'full' })}
            </span>

            <Language />
          </div>
        </div>
      </div>
    </div>
  );
}
