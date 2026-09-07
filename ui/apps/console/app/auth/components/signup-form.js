'use client';
import { useState } from 'react';
import { cn } from '@sentinez/ui/lib/utils';
import { Button } from '@sentinez/ui/components/button';
import { Card, CardContent, } from '@sentinez/ui/components/card';
import { Input } from '@sentinez/ui/components/input';
import { Label } from '@sentinez/ui/components/label';
import { PasskeyRegister } from '@/lib/api/iam/passkey';
export function SignUpForm({ className, ...props }) {
    const [email, setEmail] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    async function handleSignUp(e) {
        e.preventDefault();
        if (!email) {
            alert('Please enter your email first');
            return;
        }
        try {
            setIsLoading(true);
            await PasskeyRegister({ emailOrUsername: email });
            alert('Account created successfully!');
        }
        catch (err) {
            console.error(err);
            alert('Sign up failed: ' + err.message);
        }
        finally {
            setIsLoading(false);
        }
    }
    return (<div className={cn('flex flex-col gap-6', className)} {...props}>
      <Card>
        {/* <CardHeader className="text-center">
          <CardTitle className="text-xl">
            <a href="#" className="flex items-center gap-2 self-center font-medium">
              <div className="text-primary-foreground flex size-6 items-center justify-center rounded-md">
                <Image width={600} height={600} src="/assets/sntz.png" alt="Image" />
              </div>
              Sentinéz
            </a>
          </CardTitle>
          <CardDescription className="text-left">
            Create your Sentinéz account
          </CardDescription>
        </CardHeader> */}
        <CardContent>
          <form onSubmit={handleSignUp}>
            <div className="grid gap-6">
              <div className="grid gap-6">
                <div className="grid gap-3">
                  <Label htmlFor="email">Email</Label>
                  <Input id="email" type="email" placeholder="m@example.com" value={email} onChange={(e) => setEmail(e.target.value)} required disabled={isLoading}/>
                </div>
                <Button type="submit" className="w-full cursor-pointer" disabled={isLoading}>
                  {isLoading ? 'Creating account…' : 'Sign Up'}
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
    </div>);
}
