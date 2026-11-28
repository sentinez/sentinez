import { Notification } from '@sentinez/proto/sentinez/apps/communication/v1/model';
import { NotiType } from '@sentinez/proto/sentinez/types/notify/v1/notify';

import { SAMPLE_TENANT_ID } from '@/lib/faker/tenant';
import { SampleStore } from '@/lib/faker/store';

// Fixed dates: sample data is rendered on the server too
const at = (iso: string) => ({ createdAt: new Date(iso), updatedAt: new Date(iso) });

const SAMPLE_NOTIFICATIONS: Notification[] = [
  Notification.fromPartial({
    id: 'ntf_sample-3',
    tenantId: SAMPLE_TENANT_ID,
    metadata: at('2026-11-28T09:12:00Z'),
    event: {
      title: 'Rate limit triggered',
      message: 'api.example.com blocked 1,204 requests from 3 clients.',
      notiType: NotiType.NOTI_TYPE_WARNING,
    },
  }),
  Notification.fromPartial({
    id: 'ntf_sample-2',
    tenantId: SAMPLE_TENANT_ID,
    metadata: at('2026-11-28T08:40:00Z'),
    event: {
      title: 'Certificate renewed',
      message: 'The certificate of shop.example.com is valid until 2027-02-26.',
      notiType: NotiType.NOTI_TYPE_SUCCESS,
    },
  }),
  Notification.fromPartial({
    id: 'ntf_sample-1',
    tenantId: SAMPLE_TENANT_ID,
    metadata: at('2026-11-27T17:05:00Z'),
    event: {
      title: 'Upstream unreachable',
      message: 'The edge could not reach the upstream of api.example.com for 2 minutes.',
      notiType: NotiType.NOTI_TYPE_ERROR,
    },
  }),
];

export const notificationStore = new SampleStore(SAMPLE_NOTIFICATIONS);
