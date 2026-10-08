'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useTranslations } from 'next-intl';
import { cn } from '@sentinez/ui/lib/utils';
import { Button } from '@sentinez/ui/components/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@sentinez/ui/components/card';
import { Input } from '@sentinez/ui/components/input';
import { Label } from '@sentinez/ui/components/label';
import { Login } from '@/lib/api/iam/auth';
import { toast } from '@/lib/toast';
import { handleStatusError } from '@/lib/error-handler';
import { SENTINEZ_ACCESS_TOKEN_KEY, SENTINEZ_USER_KEY } from '@/lib/const';

export function LoginForm({ className, ...props }: React.ComponentProps<'div'>) {
  const t = useTranslations('Auth');
  const te = useTranslations('Errors');
  const router = useRouter();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  async function handleLogin(e: React.ChangeEvent) {
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
        localStorage.setItem(
          SENTINEZ_USER_KEY,
          JSON.stringify({
            name: res.user?.fullName,
            email: res.user?.fullName,
            avatar: '/assets/sntz.png',
          }),
        );
      }

      router.push('/console');
    } catch (error: any) {
      console.error('Login error:', error);
      handleStatusError(error, te);
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <div className={cn('flex flex-col gap-6', className)} {...props}>
      <Card>
        <CardHeader className="text-center">
          <CardTitle className="text-xl">{t('login')}</CardTitle>
          <CardDescription className=" text-center">{t('loginDescription')}</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleLogin}>
            <div className="grid gap-6">
              <div className="grid gap-6">
                <div className="grid gap-3">
                  <Label htmlFor="email">{t('email')}</Label>
                  <Input
                    id="email"
                    placeholder="m@example.com"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    required
                    disabled={isLoading}
                  />
                </div>
                <div className="grid gap-3">
                  <div className="flex items-center">
                    <Label htmlFor="password">{t('password')}</Label>
                    <a href="#" className="ml-auto text-sm underline-offset-4 hover:underline">
                      {t('forgotPassword')}
                    </a>
                  </div>
                  <Input
                    id="password"
                    type="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    required
                    disabled={isLoading}
                  />
                </div>
                <Button type="submit" className="w-full cursor-pointer" disabled={isLoading}>
                  {isLoading ? t('loggingIn') : t('login')}
                </Button>
              </div>
              <div className="text-center text-sm">
                {t('noAccount')} <br />
                <a href="/auth/signup" className="underline underline-offset-4">
                  {t('signUp')}
                </a>
                <br />
                {t('or')}
                <br />
                <a href="/auth/passkey/login" className="underline underline-offset-4">
                  {t('passkey')}
                </a>
              </div>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
