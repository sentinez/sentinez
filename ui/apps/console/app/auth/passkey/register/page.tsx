import { PasskeyRegisterForm } from '@/app/auth/components/passkey-register-form';

export default function PasskeyRegisterPage() {
  return (
    <div className="bg-muted flex min-h-svh flex-col items-center justify-center gap-6 p-6 md:p-10">
      <div className="flex w-full max-w-sm flex-col gap-6">
        <PasskeyRegisterForm />
      </div>
    </div>
  );
}
