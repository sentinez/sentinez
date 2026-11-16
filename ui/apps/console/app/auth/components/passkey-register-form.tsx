'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useTranslations } from 'next-intl';
import { Fingerprint } from 'lucide-react';
import { PasskeyRegister } from '@/lib/api/iam/passkey';
import { toast } from '@/lib/toast';
import { handleStatusError } from '@/lib/error-handler';
import { AuthFooter, AuthForm, AuthHeader, AuthSubmit, EmailField } from './auth-form';

export function PasskeyRegisterForm() {
  const t = useTranslations('Auth');
  const te = useTranslations('Errors');
  const router = useRouter();
  const [email, setEmail] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  async function handleRegisterPasskey(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!email) {
      toast.warning(t('inputRequired'), t('enterEmail'));
      return;
    }

    try {
      setIsLoading(true);
      await PasskeyRegister({ emailOrUsername: email });
      toast.success(t('registerSuccess'), t('registerSuccessDescription'));
      router.push('/auth/passkey/login');
    } catch (err: any) {
      console.error('Passkey registration error:', err);
      if (err.name === 'NotAllowedError') {
        toast.error(t('registerCancelled'), t('registerCancelledDescription'));
      } else {
        handleStatusError(err, te);
      }
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <AuthForm>
      <AuthHeader title={t('passkeyRegisterTitle')} description={t('passkeyRegisterDescription')} />

      <form onSubmit={handleRegisterPasskey} className="flex flex-col gap-5">
        <EmailField
          label={t('email')}
          type="email"
          autoComplete="username"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          disabled={isLoading}
        />

        <AuthSubmit loading={isLoading} icon={Fingerprint}>
          {isLoading ? t('registering') : t('registerPasskey')}
        </AuthSubmit>
      </form>

      <AuthFooter text={t('hasAccount')} linkText={t('signIn')} href="/auth/passkey/login" />
    </AuthForm>
  );
}
