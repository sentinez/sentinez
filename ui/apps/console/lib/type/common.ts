import { ControlPlane } from '@sentinez/proto/sentinez/types/v1/known';

/** The signed-in user as kept in localStorage (see storeUser in lib/session.ts) */
export interface StoredUser {
  /** '' for a session stored before the id was kept */
  id: string;
  email: string;
  name: string;
  controlPlane: ControlPlane;
  /** ISO 8601, '' when unknown */
  createdAt: string;
  updatedAt: string;
}

/** StoredUser as components read it (see useLocalUser) */
export interface UserDataLocal extends StoredUser {
  /** Generated from the user id, '' when nobody is signed in */
  avatar: string;
}
