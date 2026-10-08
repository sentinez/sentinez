'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useTranslations } from 'next-intl';
import { cn } from '@sentinez/ui/lib/utils';
import { Button } from '@sentinez/ui/components/button';
import { Card, CardContent } from '@sentinez/ui/components/card';
import { Input } from '@sentinez/ui/components/input';
import { Label } from '@sentinez/ui/components/label';
import { PasskeyRegister } from '@/lib/api/iam/passkey';
import { toast } from '@/lib/toast';
import { handleStatusError } from '@/lib/error-handler';

export function PasskeyRegisterForm({ className, ...props }: React.ComponentProps<'div'>) {
  const t = useTranslations('Auth');
  const te = useTranslations('Errors');
  const router = useRouter();
  const [email, setEmail] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  async function handleRegisterPasskey(e: React.ChangeEvent) {
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
    <div className={cn('flex flex-col gap-6', className)} {...props}>
      <Card>
        <CardContent>
          <form onSubmit={handleRegisterPasskey}>
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
                  {isLoading ? t('registering') : t('registerPasskey')}
                </Button>
              </div>
              <div className="text-center text-sm">
                {t('hasAccount')} <br />
                <a href="/auth/passkey/login" className="underline underline-offset-4">
                  {t('signIn')}
                </a>
              </div>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
