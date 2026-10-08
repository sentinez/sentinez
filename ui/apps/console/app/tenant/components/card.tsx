import * as React from 'react';

import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@sentinez/ui/components/card';
import { Button } from '@sentinez/ui/components/button';
import { MoveRight, Plus } from 'lucide-react';
import Link from 'next/link';
import { useTranslations } from 'next-intl';

export function Tenant() {
  const t = useTranslations('Tenant');

  return (
    <Card className="w-[300px] hover:scale-105 transition-transform duration-300">
      <CardHeader>
        <CardTitle>Sentinez</CardTitle>
        <CardDescription>{t('member')}</CardDescription>
      </CardHeader>
      <CardContent></CardContent>
      <CardFooter className="flex justify-between">
        <Link href="/console" className="w-full">
          <Button className="w-full cursor-pointer">
            {t('go')} <MoveRight />
          </Button>
        </Link>
      </CardFooter>
    </Card>
  );
}

export function TeamJoin() {
  const t = useTranslations('Tenant');

  return (
    <Card className="w-[300px] hover:scale-105 transition-transform duration-300">
      <CardHeader>
        <CardTitle>{t('createOrJoin')}</CardTitle>
        <CardDescription>{t('createOrJoinDescription')}</CardDescription>
      </CardHeader>
      <CardContent></CardContent>
      <CardFooter className="flex justify-between">
        <Button variant="outline" className=" cursor-pointer">
          {t('join')} <Plus />
        </Button>
        <Button className=" cursor-pointer">
          {t('create')} <MoveRight />
        </Button>
      </CardFooter>
    </Card>
  );
}
