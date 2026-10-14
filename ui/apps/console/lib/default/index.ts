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
  Truck,
  BrickWallFire,
  Signal,
  Globe,
  Settings,
  BadgeCheck,
} from 'lucide-react';
import type messages from '@/messages/en.json';

export type NavTitleKey = keyof (typeof messages)['Nav'];

export interface NavItem {
  /** Key in the "Nav" messages namespace */
  titleKey: NavTitleKey;
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
      titleKey: 'console',
      url: '/d/console/domain',
      icon: Gamepad2,
      isActive: true,
      items: [
        {
          titleKey: 'domains',
          url: '/d/console/domain',
          icon: Globe,
          items: [],
        },
        {
          titleKey: 'member',
          url: '/d/console/member',
          icon: Building2,
          items: [],
        },
      ],
    },
    {
      titleKey: 'setting',
      url: '/d/setting',
      icon: Settings,
      isActive: false,
      items: [
        {
          titleKey: 'account',
          url: '/d/setting/account',
          icon: BadgeCheck,
          items: [],
        },
        {
          titleKey: 'tenant',
          url: '/d/setting/tenant',
          icon: Building2,
          items: [],
        },
      ],
    },
  ],
  domainNavMain: [
    {
      titleKey: 'tenant',
      url: '/d/console/tenant/resource',
      icon: User,
      isActive: true,
      items: [
        {
          titleKey: 'resource',
          url: '/d/console/tenant/resource',
          icon: Server,
        },
      ],
    },
    {
      titleKey: 'delivery',
      url: '/d/console/delivery/cdn',
      icon: Truck,
      isActive: false,
      items: [
        {
          titleKey: 'cdn',
          url: '/d/console/delivery/cdn',
          icon: Globe,
        },
      ],
    },
    {
      titleKey: 'security',
      url: '/d/console/security/rate-limiter',
      icon: Shield,
      isActive: false,
      items: [
        {
          titleKey: 'rateLimiter',
          url: '/d/console/security/rate-limiter',
          icon: Signal,
        },
        {
          titleKey: 'secRule',
          url: '/d/console/security/sec-rule',
          icon: ShieldCheck,
        },
        {
          titleKey: 'rulesets',
          url: '/d/console/security/rulesets',
          icon: BrickWallFire,
        },
      ],
    },
    {
      titleKey: 'analytic',
      url: '/d/console/analytic/logs',
      icon: Activity,
      isActive: false,
      items: [
        {
          titleKey: 'logs',
          url: '/d/console/analytic/logs',
          icon: SquareActivity,
          isActive: true,
          items: [
            {
              titleKey: 'secRule',
              url: '/d/console/analytic/logs/security/sec-rule',
              icon: ShieldCheck,
            },
            {
              titleKey: 'rulesets',
              url: '/d/console/analytic/logs/security/rulesets',
              icon: BrickWallFire,
            },
          ],
        },
        {
          titleKey: 'activity',
          url: '/d/console/analytic/activity',
          icon: ShieldAlert,
        },
      ],
    },
  ],
};
