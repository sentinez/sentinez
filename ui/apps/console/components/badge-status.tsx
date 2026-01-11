import { Badge } from '@sentinez/ui/components/badge';
import { Check, X } from 'lucide-react';

export default function BadgeStatus(props: { status: string; value: string }) {
  if (props.status === 'active') {
    return (
      <Badge variant="default">
        Active <Check />
      </Badge>
    );
  }

  if (props.status === 'disable') {
    return (
      <Badge variant="destructive">
        Disable <X />
      </Badge>
    );
  }

  return <Badge variant="secondary">{props.value}</Badge>;
}
