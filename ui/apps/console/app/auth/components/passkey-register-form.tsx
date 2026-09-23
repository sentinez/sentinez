'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { cn } from '@sentinez/ui/lib/utils';
import { Button } from '@sentinez/ui/components/button';
import { Card, CardContent } from '@sentinez/ui/components/card';
import { Input } from '@sentinez/ui/components/input';
import { Label } from '@sentinez/ui/components/label';
import { PasskeyRegister } from '@/lib/api/iam/passkey';
import { toast } from '@/lib/toast';
import { handleStatusError } from '@/lib/error-handler';

export function PasskeyRegisterForm({ className, ...props }: React.ComponentProps<'div'>) {
  const router = useRouter();
  const [email, setEmail] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  async function handleRegisterPasskey(e: React.ChangeEvent) {
    e.preventDefault();
    if (!email) {
      toast.warning('Input Required', 'Please enter your email address.');
      return;
    }

    try {
      setIsLoading(true);
      await PasskeyRegister({ emailOrUsername: email });
      toast.success('Registration Successful', 'Passkey registered successfully! Please sign in.');
      router.push('/auth/passkey/login');
    } catch (err: any) {
      console.error('Passkey registration error:', err);
      if (err.name === 'NotAllowedError') {
        toast.error('Registration Cancelled', 'The passkey registration prompt was cancelled.');
      } else {
        handleStatusError(err);
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
                  <Label htmlFor="email">Email</Label>
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
                  {isLoading ? 'Registering...' : 'Register Passkey'}
                </Button>
              </div>
              <div className="text-center text-sm">
                Already have an account? <br />
                <a href="/auth/passkey/login" className="underline underline-offset-4">
                  Sign in
                </a>
              </div>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}

