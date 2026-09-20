'use client';

import { useEffect, useState } from 'react';
import View from './components/view';
import IsLoading from '@/components/main-loading';
import { LoginForm } from './components/login-form';

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
      <LoginForm />
    </View>
  );
}
