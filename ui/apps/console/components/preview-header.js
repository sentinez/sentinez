import { ArrowRight } from 'lucide-react';
export default function PreviewHeader() {
    return (<div className="flex w-full items-start bg-background text-foreground" id="screenshot" data-theme-scope="preview">
      <div className="w-full border-b border-border bg-background px-3 py-1">
        <div className="flex w-full flex-wrap items-center justify-between gap-4">
          {/* Brand */}
          <div className="flex items-center gap-3">
            {/* <img
                        alt="shadcnblocks"
                        className="size-5 dark:invert"
                        src="https://deifkwefumgah.cloudfront.net/shadcnblocks/block/logos/shadcnblocks-logo.svg"
                    /> */}

            <span className="text-sm font-semibold text-foreground">Sentinéz</span>

            <span className="hidden text-muted-foreground sm:inline">·</span>

            <p className="hidden text-sm text-muted-foreground sm:block">
              Production-ready sections for your next launch.
            </p>
          </div>

          {/* Navigation */}
          <div className="flex items-center gap-5">
            <a href="#" className="text-sm text-muted-foreground transition-colors hover:text-foreground">
              Browse blocks
            </a>

            <a href="#" className="inline-flex items-center gap-1 text-sm font-medium text-primary hover:underline">
              Get Pro access
              <ArrowRight className="size-3.5" aria-hidden="true"/>
            </a>
          </div>
        </div>
      </div>
    </div>);
}
