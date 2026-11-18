import { Badge } from '@sentinez/ui/components/badge';
import { Check, Clock, X } from 'lucide-react';
import { useTranslations } from 'next-intl';

export default function BadgeStatus(props: { status: string; value: string }) {
  const t = useTranslations('Enum.status');

  if (props.status === 'active') {
    return (
      <Badge variant="default">
        {t('active')} <Check />
      </Badge>
    );
  }

  if (props.status === 'disable') {
    return (
      <Badge variant="destructive">
        {t('disable')} <X />
      </Badge>
    );
  }

  if (props.status === 'pending') {
    return (
      <Badge variant="secondary">
        {t('pending')} <Clock />
      </Badge>
    );
  }

  return <Badge variant="secondary">{props.value}</Badge>;
}
