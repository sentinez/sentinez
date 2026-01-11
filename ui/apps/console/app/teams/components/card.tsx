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

export function Tenant() {
  return (
    <Card className="w-[300px] hover:scale-105 transition-transform duration-300">
      <CardHeader>
        <CardTitle>Sentinez</CardTitle>
        <CardDescription>Teams member</CardDescription>
      </CardHeader>
      <CardContent></CardContent>
      <CardFooter className="flex justify-between">
        <Link href="/console" className="w-full">
          <Button className="w-full cursor-pointer">
            Go <MoveRight />
          </Button>
        </Link>
      </CardFooter>
    </Card>
  );
}

export function TeamJoin() {
  return (
    <Card className="w-[300px] hover:scale-105 transition-transform duration-300">
      <CardHeader>
        <CardTitle>Create or Join</CardTitle>
        <CardDescription>Create or join a team in one-click.</CardDescription>
      </CardHeader>
      <CardContent></CardContent>
      <CardFooter className="flex justify-between">
        <Button variant="outline" className=" cursor-pointer">
          Join <Plus />
        </Button>
        <Button className=" cursor-pointer">
          Create <MoveRight />
        </Button>
      </CardFooter>
    </Card>
  );
}
