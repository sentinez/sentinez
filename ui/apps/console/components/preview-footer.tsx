export default function PreviewFooter() {
  return (
    <footer className="relative z-1 w-full shrink-0 border-t bg-background text-foreground">
      <div className="w-full px-6 py-3">
        <div className="flex justify-center w-full flex-wrap items-center gap-4">
          <div className="flex flex-col items-center gap-3 w-full">
            {/* <p className="hidden sm:text-[64px] lg:text-[256px] md:text-[128px] sm:block font-bold font-['Phudu',sans-serif]">
              CONSOLE
            </p> */}
            <div className="flex gap-2 justify-center w-2/3">
              <span className="hidden text-sm sm:block items-center font-['Phudu',sans-serif] font-semibold">
                SENTINÉZ LABS
              </span>
              <span className="hidden text-sm text-muted-foreground sm:block items-center">
                &#169; 2025-2026.
              </span>
            </div>
          </div>
        </div>
      </div>
    </footer>
  );
}
