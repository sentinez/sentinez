'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useTranslations } from 'next-intl';
import { Fingerprint, Lock } from 'lucide-react';
import { PasskeyLogin } from '@/lib/api/iam/passkey';
import { toast } from '@/lib/toast';
import { handleStatusError } from '@/lib/error-handler';
import { SENTINEZ_ACCESS_TOKEN_KEY } from '@/lib/const';
import { homePath, storeUser } from '@/lib/session';
import {
  AuthAlternative,
  AuthFooter,
  AuthForm,
  AuthHeader,
  AuthSubmit,
  EmailField,
} from './auth-form';

export function PasskeyLoginForm() {
  const t = useTranslations('Auth');
  const te = useTranslations('Errors');
  const router = useRouter();
  const [email, setEmail] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  async function handleLoginPasskey(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!email) {
      toast.warning(t('inputRequired'), t('enterEmail'));
      return;
    }

    try {
      setIsLoading(true);
      const resp = await PasskeyLogin({ emailOrUsername: email });
      toast.success(t('loginSuccess'), t('welcome'));

      if (resp?.accessToken) {
        localStorage.setItem(SENTINEZ_ACCESS_TOKEN_KEY, resp.accessToken);
        storeUser(resp.user, email);
      }

      router.push(homePath());
    } catch (err: any) {
      console.error('Passkey login error:', err);
      if (err.name === 'NotAllowedError') {
        toast.error(t('passkeyCancelled'), t('passkeyCancelledDescription'));
      } else {
        handleStatusError(err, te);
      }
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <AuthForm>
      <AuthHeader title={t('passkeyLoginTitle')} description={t('passkeyLoginDescription')} />

      <form onSubmit={handleLoginPasskey} className="flex flex-col gap-5">
        <EmailField
          label={t('email')}
          type="email"
          autoComplete="username webauthn"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          disabled={isLoading}
        />

        <AuthSubmit loading={isLoading} icon={Fingerprint}>
          {isLoading ? t('authenticating') : t('login')}
        </AuthSubmit>
      </form>

      <AuthAlternative href="/auth" icon={Lock}>
        {t('loginWithPassword')}
      </AuthAlternative>

      <AuthFooter text={t('noAccount')} linkText={t('register')} href="/auth/passkey/register" />
    </AuthForm>
  );
}
