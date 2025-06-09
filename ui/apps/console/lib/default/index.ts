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
      title: 'Management',
      url: '/console/management',
      icon: User,
      isActive: true,
      items: [
        {
          title: 'organization',
          url: '/console/management/organization',
          icon: Building2,
        },
        {
          title: 'resource',
          url: '/console/management/resource',
          icon: Server,
        },
        {
          title: 'origin',
          url: '/console/management/origin',
          icon: Cloud,
        },
      ],
    },
    {
      title: 'Activity',
      url: '/console/activity',
      icon: Activity,
      isActive: false,
      items: [
        {
          title: 'access log',
          url: '/console/activity/access-log',
          icon: SquareActivity,
        },
        {
          title: 'alert',
          url: '/console/activity/alert',
          icon: ShieldAlert,
        },
      ],
    },
  ],
};
