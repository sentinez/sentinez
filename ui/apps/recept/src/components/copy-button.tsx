import { Button } from '@sentinez/ui/components/button';
import { CheckIcon, CopyIcon } from 'lucide-react';

/**
 * Copies `text`, or the text of the element `from` selects, then shows a
 * check for a moment: no React in the browser, the script of Layout.astro
 * handles every [data-copy] button.
 */
export function CopyButton({ text, from, label = 'Copy' }: { text?: string; from?: string; label?: string }) {
  return (
    <Button
      type="button"
      variant="ghost"
      size="icon-xs"
      data-copy={text ?? ''}
      data-copy-from={from}
      aria-label={label}
      title={label}
    >
      <CopyIcon className="copy-idle" />
      <CheckIcon className="copy-done" />
    </Button>
  );
}
