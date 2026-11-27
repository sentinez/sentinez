import type { ReactNode } from 'react';

/**
 * A titled part of a form page: a heading and its hint over the fields,
 * ruled off from what comes before it (the first section of a form is not).
 */
export function FormSection({
  title,
  hint,
  children,
}: {
  title: string;
  hint: string;
  children: ReactNode;
}) {
  return (
    <section className="mt-2 grid gap-4 border-t pt-6 first:mt-0 first:border-t-0 first:pt-0">
      <div className="grid gap-1">
        <h2 className="text-base font-semibold">{title}</h2>
        <p className="text-sm text-muted-foreground">{hint}</p>
      </div>
      {children}
    </section>
  );
}
