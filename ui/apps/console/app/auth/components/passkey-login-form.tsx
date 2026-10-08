'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useTranslations } from 'next-intl';
import { cn } from '@sentinez/ui/lib/utils';
import { Button } from '@sentinez/ui/components/button';
import { Card, CardContent } from '@sentinez/ui/components/card';
import { Input } from '@sentinez/ui/components/input';
import { Label } from '@sentinez/ui/components/label';
import { PasskeyLogin } from '@/lib/api/iam/passkey';
import { toast } from '@/lib/toast';
import { handleStatusError } from '@/lib/error-handler';
import { SENTINEZ_ACCESS_TOKEN_KEY, SENTINEZ_USER_KEY } from '@/lib/const';

export function PasskeyLoginForm({ className, ...props }: React.ComponentProps<'div'>) {
  const t = useTranslations('Auth');
  const te = useTranslations('Errors');
  const router = useRouter();
  const [email, setEmail] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  async function handleLoginPasskey(e: React.ChangeEvent) {
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
        localStorage.setItem(
          SENTINEZ_USER_KEY,
          JSON.stringify({
            name: resp.user?.fullName || email,
            email: resp.user?.email || email,
            avatar: '/assets/sntz.png',
          }),
        );
      }

      router.push('/console');
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
    <div className={cn('flex flex-col gap-6', className)} {...props}>
      <Card>
        <CardContent>
          <form onSubmit={handleLoginPasskey}>
            <div className="grid gap-6">
              <div className="grid gap-6">
                <div className="grid gap-3">
                  <Label htmlFor="email">{t('email')}</Label>
                  <Input
                    id="email"
                    type="email"
                    placeholder="m@example.com"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    required
                    disabled={isLoading}
                  />
                </div>
                <Button type="submit" className="w-full cursor-pointer" disabled={isLoading}>
                  {isLoading ? t('authenticating') : t('login')}
                </Button>
              </div>
              <div className="text-center text-sm">
                {t('noAccount')} <br />
                <a href="/auth/passkey/register" className="underline underline-offset-4">
                  {t('register')}
                </a>
              </div>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
