import { use } from 'react';

type Props = {
  params: Promise<{
    domain: string;
    resourceId: string;
  }>;
};

export default function ResourceSetting({ params }: Props) {
  const props = use(params);

  return <div>Resource ID: {props.resourceId} (Tenant Domain: {props.domain})</div>;
}
