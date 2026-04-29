import {
  User,
  Server,
  Building2,
  Shield,
  ShieldCheck,
  Activity,
  ShieldAlert,
  SquareActivity,
  Users,
  Gamepad2,
} from 'lucide-react';

// This is sample data
export const dashboard = {
  user: {
    name: 'shadcn',
    email: 'm@example.com',
    avatar: '/assets/sntz.png',
  },
  tenant: [
    {
      name: 'sentinez',
      logo: Users,
      plan: 'member',
    },
  ],
  rootNavMain: [
    {
      title: 'Console',
      url: '/console/resource',
      icon: Gamepad2,
      isActive: true,
      items: [
        {
          title: 'resource',
          url: '/console/resource',
          icon: Server,
        },
        {
          title: 'member',
          url: '/console/member',
          icon: Building2,
        },
      ],
    },
  ],
  domainNavMain: [
    {
      title: 'Tenants',
      url: '/console/tenant/resource',
      icon: User,
      isActive: true,
      items: [
        {
          title: 'resource',
          url: '/console/tenant/resource',
          icon: Server,
        },
      ],
    },
    {
      title: 'Securities',
      url: '/console/security/rulebased',
      icon: Shield,
      isActive: false,
      items: [
        {
          title: 'rule based',
          url: '/console/security/rulebased',
          icon: ShieldCheck,
        },
      ],
    },
    {
      title: 'Analytics',
      url: '/console/analytic/logs',
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
