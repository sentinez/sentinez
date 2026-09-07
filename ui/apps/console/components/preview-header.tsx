import Language from './language';

export default function PreviewHeader() {
  return (
    <div
      className="flex w-full items-start bg-background text-foreground"
      id="screenshot"
      data-theme-scope="preview"
    >
      <div className="w-full border-b border-border bg-background px-3 py-1">
        <div className="flex w-full flex-wrap items-center justify-between gap-4">
          <div className="flex items-center gap-3">
            <span className="text-xs font-semibold text-foreground">Sentinéz</span>

            <span className="hidden text-muted-foreground sm:inline">·</span>

            <p className="hidden text-xs text-muted-foreground sm:block">Hi, Welcome back !!!</p>
          </div>

          <div className="flex items-center gap-5">
            <span className="text-xs text-muted-foreground transition-colors hover:text-foreground">
              {new Date().toDateString()}
            </span>

            <Language />
          </div>
        </div>
      </div>
    </div>
  );
}
