'use client';

import { ComponentProps, ComponentType, ReactNode } from 'react';
import Link from 'next/link';
import { useTranslations } from 'next-intl';
import { Loader2, Mail } from 'lucide-react';
import { cn } from '@sentinez/ui/lib/utils';
import { Button } from '@sentinez/ui/components/button';
import { Field, FieldLabel, FieldSeparator } from '@sentinez/ui/components/field';
import { Input } from '@sentinez/ui/components/input';

// Building blocks shared by the auth forms (login, sign up, passkey), so the
// pages look the same.

type Icon = ComponentType<{ className?: string }>;

export function AuthForm({ children }: { children: ReactNode }) {
  return <div className="flex flex-col gap-8">{children}</div>;
}

export function AuthHeader({ title, description }: { title: string; description: string }) {
  return (
    <div className="flex flex-col gap-2">
      <h1 className="text-2xl font-semibold tracking-tight">{title}</h1>
      <p className="text-muted-foreground text-sm text-balance">{description}</p>
    </div>
  );
}

type IconInputProps = ComponentProps<typeof Input> & {
  icon?: Icon;
  // trailing sits at the right end of the input (e.g. a show-password button)
  trailing?: ReactNode;
};

export function IconInput({ icon: Icon = Mail, trailing, className, ...props }: IconInputProps) {
  return (
    <div className="relative">
      <Icon className="text-muted-foreground pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2" />
      <Input className={cn('h-10 pl-9', trailing && 'pr-9', className)} {...props} />
      {trailing && (
        <div className="absolute inset-y-0 right-0 flex w-9 items-center justify-center">
          {trailing}
        </div>
      )}
    </div>
  );
}

export function AuthField({
  id,
  label,
  aside,
  children,
}: {
  id: string;
  label: string;
  aside?: ReactNode;
  children: ReactNode;
}) {
  return (
    <Field className="gap-2">
      <div className="flex items-center">
        <FieldLabel htmlFor={id}>{label}</FieldLabel>
        {aside && <div className="ml-auto">{aside}</div>}
      </div>
      {children}
    </Field>
  );
}

export function EmailField({
  label,
  ...props
}: { label: string } & Omit<IconInputProps, 'id' | 'name'>) {
  return (
    <AuthField id="email" label={label}>
      <IconInput id="email" name="email" placeholder="m@example.com" required {...props} />
    </AuthField>
  );
}

export function AuthSubmit({
  loading,
  icon: Icon,
  children,
}: {
  loading: boolean;
  icon?: Icon;
  children: ReactNode;
}) {
  return (
    <Button type="submit" className="h-10 w-full cursor-pointer" disabled={loading}>
      {loading ? <Loader2 className="animate-spin" /> : Icon && <Icon />}
      {children}
    </Button>
  );
}

/** An "or" separator followed by a link to the other way of signing in */
export function AuthAlternative({
  href,
  icon: Icon,
  children,
}: {
  href: string;
  icon: Icon;
  children: ReactNode;
}) {
  const t = useTranslations('Auth');

  return (
    <div className="flex flex-col gap-4">
      <FieldSeparator className="text-xs uppercase">{t('or')}</FieldSeparator>
      <Button variant="outline" className="h-10 w-full" asChild>
        <Link href={href}>
          <Icon />
          {children}
        </Link>
      </Button>
    </div>
  );
}

export function AuthFooter({
  text,
  linkText,
  href,
}: {
  text: string;
  linkText: string;
  href: string;
}) {
  return (
    <p className="text-muted-foreground text-center text-sm">
      {text}{' '}
      <Link href={href} className="text-foreground font-medium underline-offset-4 hover:underline">
        {linkText}
      </Link>
    </p>
  );
}
