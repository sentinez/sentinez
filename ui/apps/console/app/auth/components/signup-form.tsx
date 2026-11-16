'use client';

import { useState } from 'react';
import { useTranslations } from 'next-intl';
import { PasskeyRegister } from '@/lib/api/iam/passkey';
import { toast } from '@/lib/toast';
import { AuthFooter, AuthForm, AuthHeader, AuthSubmit, EmailField } from './auth-form';

export function SignUpForm() {
  const t = useTranslations('Auth');
  const [email, setEmail] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  async function handleSignUp(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!email) {
      toast.warning(t('inputRequired'), t('enterEmailFirst'));
      return;
    }

    try {
      setIsLoading(true);
      await PasskeyRegister({ emailOrUsername: email });
      toast.success(t('accountCreated'));
    } catch (err: any) {
      console.error(err);
      toast.error(t('signUpFailedTitle'), t('signUpFailed', { message: err.message }));
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <AuthForm>
      <AuthHeader title={t('signUpTitle')} description={t('signUpDescription')} />

      <form onSubmit={handleSignUp} className="flex flex-col gap-5">
        <EmailField
          label={t('email')}
          type="email"
          autoComplete="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          disabled={isLoading}
        />

        <AuthSubmit loading={isLoading}>
          {isLoading ? t('creatingAccount') : t('signUpButton')}
        </AuthSubmit>
      </form>

      <AuthFooter text={t('hasAccount')} linkText={t('signIn')} href="/auth" />
    </AuthForm>
  );
}
