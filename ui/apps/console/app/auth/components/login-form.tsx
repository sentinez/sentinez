'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
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
  const router = useRouter();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  async function handleLogin(e: React.ChangeEvent) {
    e.preventDefault();

    if (!email || !password) {
      toast.warning('Input Required', 'Please enter both email and password.');
      return;
    }

    try {
      setIsLoading(true);
      const res = await Login({ emailOrUsername: email, password });
      toast.success('Login Successful', 'Welcome back to Sentinez!');

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
      handleStatusError(error);
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <div className={cn('flex flex-col gap-6', className)} {...props}>
      <Card>
        <CardHeader className="text-center">
          <CardTitle className="text-xl">Login</CardTitle>
          <CardDescription className=" text-center">
            Login with your Sentinez account
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleLogin}>
            <div className="grid gap-6">
              <div className="grid gap-6">
                <div className="grid gap-3">
                  <Label htmlFor="email">Email</Label>
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
                    <Label htmlFor="password">Password</Label>
                    <a href="#" className="ml-auto text-sm underline-offset-4 hover:underline">
                      Forgot your password?
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
                  {isLoading ? 'Logging in...' : 'Login'}
                </Button>
              </div>
              <div className="text-center text-sm">
                Don&apos;t have an account? <br />
                <a href="/auth/signup" className="underline underline-offset-4">
                  Sign up
                </a>
                <br />
                or
                <br />
                <a href="/auth/passkey/login" className="underline underline-offset-4">
                  Passkey
                </a>
              </div>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
