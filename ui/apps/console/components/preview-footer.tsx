export default function PreviewFooter() {
  return (
    <div className="flex w-full items-start bg-background text-foreground border-t">
      <div className="w-full border-b border-border bg-background px-6 py-3">
        <div className="flex justify-center w-full flex-wrap items-center gap-4">
          <div className="flex items-center gap-3">
            <span className="hidden text-xs text-muted-foreground sm:block items-center">
              &#169; 2025-2026 Sentinez Labs.
            </span>
          </div>
        </div>
      </div>
    </div>
  );
}
