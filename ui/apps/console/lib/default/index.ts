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
  Gamepad,
  BrickWallFire,
  Signal,
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
          title: 'Resource',
          url: '/console/resource',
          icon: Server,
        },
        {
          title: 'Member',
          url: '/console/member',
          icon: Building2,
        },
      ],
    },
  ],
  domainNavMain: [
    {
      title: 'Tenant',
      url: '/console/tenant/resource',
      icon: User,
      isActive: true,
      items: [
        {
          title: 'Resource',
          url: '/console/tenant/resource',
          icon: Server,
        },
      ],
    },
    {
      title: 'Controller',
      url: '/console/controller/cdn',
      icon: Gamepad,
      isActive: false,
      items: [
        {
          title: 'CDN',
          url: '/console/controller/cdn',
          icon: ShieldCheck,
        },
      ],
    },
    {
      title: 'Security',
      url: '/console/security/rate-limiter',
      icon: Shield,
      isActive: false,
      items: [
        {
          title: 'Rate limiter',
          url: '/console/security/rate-limiter',
          icon: Signal,
        },
        {
          title: 'Rule based',
          url: '/console/security/rule-based',
          icon: ShieldCheck,
        },
        {
          title: 'Rulesets',
          url: '/console/security/rulesets',
          icon: BrickWallFire,
        },
      ],
    },
    {
      title: 'Analytic',
      url: '/console/analytic/logs',
      icon: Activity,
      isActive: false,
      items: [
        {
          title: 'Logs',
          url: '/console/analytic/logs',
          icon: SquareActivity,
          isActive: true,
          items: [
            {
              title: 'Rule Based',
              url: '/console/analytic/logs/security/rule-based',
              icon: ShieldCheck,
            },
            {
              title: 'Rulesets',
              url: '/console/analytic/logs/security/rulesets',
              icon: BrickWallFire,
            },
          ],
        },
        {
          title: 'Activity',
          url: '/console/analytic/activity',
          icon: ShieldAlert,
        },
      ],
    },
  ],
};
