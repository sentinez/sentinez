'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useTranslations } from 'next-intl';
import { Eye, EyeOff, KeyRound, Lock } from 'lucide-react';
import { Login } from '@/lib/api/iam/auth';
import { toast } from '@/lib/toast';
import { handleStatusError } from '@/lib/error-handler';
import { SENTINEZ_ACCESS_TOKEN_KEY } from '@/lib/const';
import { homePath, storeUser } from '@/lib/session';
import {
  AuthAlternative,
  AuthField,
  AuthFooter,
  AuthForm,
  AuthHeader,
  AuthSubmit,
  EmailField,
  IconInput,
} from './auth-form';

export function LoginForm() {
  const t = useTranslations('Auth');
  const te = useTranslations('Errors');
  const router = useRouter();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  async function handleLogin(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();

    if (!email || !password) {
      toast.warning(t('inputRequired'), t('enterEmailPassword'));
      return;
    }

    try {
      setIsLoading(true);
      const res = await Login({ emailOrUsername: email, password });
      toast.success(t('loginSuccess'), t('welcome'));

      if (res.accessToken) {
        localStorage.setItem(SENTINEZ_ACCESS_TOKEN_KEY, res.accessToken);
        storeUser(res.user, email);
      }

      router.push(homePath());
    } catch (error: any) {
      console.error('Login error:', error);
      handleStatusError(error, te);
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <AuthForm>
      <AuthHeader title={t('welcomeBack')} description={t('loginDescription')} />

      <form onSubmit={handleLogin} className="flex flex-col gap-5">
        <EmailField
          label={t('emailOrUsername')}
          autoComplete="username"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          disabled={isLoading}
        />

        <AuthField
          id="password"
          label={t('password')}
          aside={
            <a
              href="#"
              className="text-muted-foreground hover:text-foreground text-xs underline-offset-4 hover:underline"
            >
              {t('forgotPassword')}
            </a>
          }
        >
          <IconInput
            icon={Lock}
            id="password"
            name="password"
            type={showPassword ? 'text' : 'password'}
            autoComplete="current-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            disabled={isLoading}
            trailing={
              <button
                type="button"
                onClick={() => setShowPassword((v) => !v)}
                aria-label={showPassword ? t('hidePassword') : t('showPassword')}
                aria-pressed={showPassword}
                className="text-muted-foreground hover:text-foreground cursor-pointer"
              >
                {showPassword ? <EyeOff className="size-4" /> : <Eye className="size-4" />}
              </button>
            }
          />
        </AuthField>

        <AuthSubmit loading={isLoading}>{isLoading ? t('loggingIn') : t('login')}</AuthSubmit>
      </form>

      <AuthAlternative href="/auth/passkey/login" icon={KeyRound}>
        {t('loginWithPasskey')}
      </AuthAlternative>

      <AuthFooter text={t('noAccount')} linkText={t('signUp')} href="/auth/signup" />
    </AuthForm>
  );
}
