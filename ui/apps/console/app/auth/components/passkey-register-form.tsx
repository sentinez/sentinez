'use client';

import { useState } from 'react';
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
import Image from 'next/image';
import { PasskeyRegister } from '@sentinez/api/iam/passkey';

export function PasskeyRegisterForm({ className, ...props }: React.ComponentProps<'div'>) {
  const [email, setEmail] = useState('');

  async function handleRegisterPasskey(e: React.FormEvent) {
    e.preventDefault();
    if (!email) {
      alert('Please enter your email first');
      return;
    }

    try {
      const _ = PasskeyRegister({ emailOrUsername: email });
    } catch (err: any) {
      console.error(err);
      alert('Registration failed: ' + err.message);
    }
  }

  return (
    <div className={cn('flex flex-col gap-6', className)} {...props}>
      <Card>
        <CardHeader className="text-center">
          <CardTitle className="text-xl">
            <a href="#" className="flex items-center gap-2 self-center font-medium">
              <div className="text-primary-foreground flex size-6 items-center justify-center rounded-md">
                <Image width={600} height={600} src="/assets/sntz.png" alt="Image" />
              </div>
              SENTINEZ
            </a>
          </CardTitle>
          <CardDescription className=" text-left">
            Register with your Sentinez account
          </CardDescription>
        </CardHeader>
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
                  />
                </div>
                <Button type="submit" className="w-full cursor-pointer">
                  Register Passkey
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
      <div className="text-muted-foreground *:[a]:hover:text-primary text-center text-xs text-balance *:[a]:underline *:[a]:underline-offset-4">
        By clicking continue, you agree to our <a href="#">Terms of Service</a> and{' '}
        <a href="#">Privacy Policy</a>.
      </div>
    </div>
  );
}
