import type { ComponentProps } from 'react';
import { cn } from '@sentinez/ui/lib/utils';

export type CodeBlockProps = Omit<ComponentProps<'pre'>, 'children'> & {
  /** The text to show as is, or a value to show as pretty-printed JSON. */
  value: unknown;
};

/**
 * CodeBlock shows text (a certificate, a log) or a JSON value, pretty-printed,
 * in a scrolling block. It takes the width of its parent and never more
 * (`min-w-0`), so a long line scrolls inside the block instead of widening
 * the card around it at any zoom level.
 */
export function CodeBlock({ value, className, ...props }: CodeBlockProps) {
  const text = typeof value === 'string' ? value : JSON.stringify(value, null, 2);

  return (
    <pre
      data-slot="code-block"
      className={cn(
        'w-full min-w-0 max-h-96 overflow-auto scrollbar-thin rounded-md bg-muted p-3',
        'font-mono text-xs leading-relaxed text-foreground',
        className,
      )}
      {...props}
    >
      {text}
    </pre>
  );
}
