import {
  User,
  Server,
  Building2,
  Shield,
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
  tenant: [
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
          title: 'resource',
          url: '/console/tenant/resource',
          icon: Server,
        },
        {
          title: 'member',
          url: '/console/tenant/member',
          icon: Building2,
        },
        // {
        //   title: 'origin',
        //   url: '/console/tenant/origin',
        //   icon: Cloud,
        // },
      ],
    },
    {
      title: 'Securities',
      url: '/console/security',
      icon: Shield,
      isActive: false,
      items: [
        {
          title: 'logs',
          url: '/console/security/logs',
          icon: SquareActivity,
        },
        {
          title: 'activity',
          url: '/console/security/activity',
          icon: ShieldAlert,
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
          title: 'logs',
          url: '/console/analytic/logs',
          icon: SquareActivity,
        },
        {
          title: 'activity',
          url: '/console/analytic/activity',
          icon: ShieldAlert,
        },
      ],
    },
  ],
};
