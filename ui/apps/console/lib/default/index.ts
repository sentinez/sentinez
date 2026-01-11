import {
  User,
  Server,
  Building2,
  Cloud,
  Activity,
  ShieldAlert,
  SquareActivity,
  Users,
} from 'lucide-react';

// This is sample data
export const dashboard = {
  user: {
    name: 'shadcn',
    email: 'm@example.com',
    avatar: '/avatars/shadcn.jpg',
  },
  teams: [
    {
      name: 'sentinez',
      logo: Users,
      plan: 'member',
    },
  ],
  navMain: [
    {
      title: 'Tenants',
      url: '/console/tenant',
      icon: User,
      isActive: true,
      items: [
        {
          title: 'organization',
          url: '/console/tenant/organization',
          icon: Building2,
        },
        {
          title: 'resource',
          url: '/console/tenant/resource',
          icon: Server,
        },
        {
          title: 'origin',
          url: '/console/tenant/origin',
          icon: Cloud,
        },
      ],
    },
    {
      title: 'Analytics',
      url: '/console/analytic',
      icon: Activity,
      isActive: false,
      items: [
        {
          title: 'access log',
          url: '/console/analytic/access-log',
          icon: SquareActivity,
        },
        {
          title: 'alert',
          url: '/console/analytic/alert',
          icon: ShieldAlert,
        },
      ],
    },
  ],
};
