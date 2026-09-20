'use client';

import { PasskeyLoginForm } from '@/app/auth/components/passkey-login-form';
import IsLoading from '@/components/main-loading';
import { useEffect, useState } from 'react';
import View from '../../components/view';

export default function Page() {
  const [loading, setLoading] = useState(true);
  useEffect(() => {
    const run = async () => {
      await new Promise((resolve) => setTimeout(resolve, 1500));

      setLoading(false);
    };

    run();
  }, []);

  if (loading) return <IsLoading timeout={1000} />;

  return (
    <View>
      <PasskeyLoginForm />
    </View>
  );
}
