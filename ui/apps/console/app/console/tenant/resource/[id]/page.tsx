import { use } from 'react';

type Props = {
  params: Promise<{
    id: string;
  }>;
};

export default function ResourceSetting({ params }: Props) {
  const props = use(params);

  return <div>User ID: {props.id}</div>;
}
