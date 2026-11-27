import type { ListUsersRequest, LoginResponse } from '@sentinez/proto/sentinez/apps/iam/v1/iam';
import { User } from '@sentinez/proto/sentinez/apps/iam/v1/model';
import { ControlPlane } from '@sentinez/proto/sentinez/types/v1/known';
import { SampleStore, byIds } from '@/lib/faker/store';

// Written as proto JSON (the wire format), parsed with User.fromJSON
const SAMPLE_USERS: User[] = [
  {
    metadata: { createdAt: '2026-08-20T03:00:00Z', updatedAt: '2026-08-20T03:00:00Z' },
    id: 'user_0199a6f0-4c1e-7a2b-9d3e-400000000001',
    fullName: 'Alice Nguyen',
    email: 'alice@example.com',
    controlPlane: 'CONTROL_PLANE_PORTAL',
  },
  {
    metadata: { createdAt: '2026-08-25T09:30:00Z', updatedAt: '2026-09-10T11:00:00Z' },
    id: 'user_0199a6f0-4c1e-7a2b-9d3e-400000000002',
    fullName: 'Bob Tran',
    email: 'bob@example.com',
    controlPlane: 'CONTROL_PLANE_PORTAL',
  },
  {
    metadata: { createdAt: '2026-09-02T08:00:00Z', updatedAt: '2026-09-02T08:00:00Z' },
    id: 'user_0199a6f0-4c1e-7a2b-9d3e-400000000003',
    fullName: 'Carol Le',
    email: 'carol@example.com',
    controlPlane: 'CONTROL_PLANE_PORTAL',
  },
].map((u) => User.fromJSON(u));

// Extra users so member lists have several pages; the last few belong to no
// tenant, so they can be invited (try dung@example.com)
const SAMPLE_NAMES = [
  'Duc Pham',
  'Emma Hoang',
  'Felix Vu',
  'Giang Dang',
  'Hana Bui',
  'Ivan Do',
  'Julia Ho',
  'Khoa Ngo',
  'Linh Duong',
  'Minh Ly',
  'Nam Truong',
  'Oanh Mai',
  'Phuc Dinh',
  'Quynh Lam',
  'Dung Cao',
  'Rin Ta',
];

const generatedUsers: User[] = SAMPLE_NAMES.map((fullName, i) => {
  const n = String(i + 4).padStart(2, '0');
  const at = new Date(Date.UTC(2026, 8, 1 + i, 8));
  return User.fromPartial({
    metadata: { createdAt: at, updatedAt: at },
    id: `user_0199a6f0-4c1e-7a2b-9d3e-4000000000${n}`,
    fullName,
    // "Duc Pham" -> duc@example.com
    email: `${fullName.slice(0, fullName.indexOf(' ')).toLowerCase()}@example.com`,
    controlPlane: ControlPlane.CONTROL_PLANE_PORTAL,
  });
});

/** Every sample user: SAMPLE_USERS first, then the generated ones */
export const ALL_SAMPLE_USERS: User[] = [...SAMPLE_USERS, ...generatedUsers];

export const userStore = new SampleStore(ALL_SAMPLE_USERS);

export const filterSampleUsers = (req: Partial<ListUsersRequest>): User[] =>
  byIds(userStore.all(), req.ids).filter(
    (u) => !req.emails?.length || req.emails.includes(u.email),
  );

// Lifetime of a sample access token
const SAMPLE_TOKEN_TTL_SECONDS = 24 * 60 * 60;

/**
 * Signs in as the sample user whose email or first name matches
 * emailOrUsername (any password), else as a new one named after it.
 */
export function sampleLogin(emailOrUsername: string): LoginResponse {
  const key = emailOrUsername.trim().toLowerCase();
  const user =
    ALL_SAMPLE_USERS.find((u) => u.email === key || u.email.split('@')[0] === key) ??
    User.fromPartial({
      id: 'user_sample',
      fullName: emailOrUsername,
      email: key.includes('@') ? key : `${key}@example.com`,
      controlPlane: ControlPlane.CONTROL_PLANE_PORTAL,
    });
  return { user, accessToken: sampleToken(user.id) };
}

// An unsigned JWT: the console only reads its expiry
function sampleToken(sub: string): string {
  const b64 = (o: object) =>
    btoa(JSON.stringify(o)).replace(/=+$/, '').replace(/\+/g, '-').replace(/\//g, '_');
  const exp = Math.floor(Date.now() / 1000) + SAMPLE_TOKEN_TTL_SECONDS;
  return `${b64({ alg: 'none', typ: 'JWT' })}.${b64({ sub, exp })}.sample`;
}
