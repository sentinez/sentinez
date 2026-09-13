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
  Globe,
} from 'lucide-react';

export interface NavItem {
  title: string;
  url: string;
  icon?: any;
  isActive?: boolean;
  items?: NavItem[];
}

// This is sample data
export const dashboard: {
  user: { name: string; email: string; avatar: string };
  tenant: { name: string; logo: any; plan: string }[];
  rootNavMain: NavItem[];
  domainNavMain: NavItem[];
} = {
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
      url: '/console/domain',
      icon: Gamepad2,
      isActive: true,
      items: [
        {
          title: 'Domains',
          url: '/console/domain',
          icon: Globe,
          items: [],
        },
        {
          title: 'Member',
          url: '/console/member',
          icon: Building2,
          items: [],
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
