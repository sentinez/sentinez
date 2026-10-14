'use client';

import { Checkbox } from '@sentinez/ui/components/checkbox';
import { Input } from '@sentinez/ui/components/input';
import { Label } from '@sentinez/ui/components/label';
import { useTranslations } from 'next-intl';
import type { Member } from '@sentinez/proto/sentinez/apps/tenant/v1/model';
import { Role } from '@sentinez/proto/sentinez/types/v1/known';

export interface MemberFormValue {
  /** Email of the user to invite; read-only once the member exists */
  email: string;
  roles: Role[];
}

type RoleTranslator = ReturnType<typeof useTranslations<'Enum'>>;
type ValidationTranslator = ReturnType<typeof useTranslations<'Validation'>>;

/** Message key of each role under Enum.role / Enum.roleDescription */
export const ROLE_KEY: Record<Role, string> = {
  [Role.ROLE_UNSPECIFIED]: 'unspecified',
  [Role.ROLE_SUPER_ADMIN]: 'superAdmin',
  [Role.ROLE_ADMIN]: 'admin',
  [Role.ROLE_LEADER]: 'leader',
  [Role.ROLE_MEMBER]: 'member',
  [Role.UNRECOGNIZED]: 'unknown',
};

// Roles a tenant can grant; super admin is a platform role
const ASSIGNABLE_ROLES = [Role.ROLE_ADMIN, Role.ROLE_LEADER, Role.ROLE_MEMBER];

// Loose check before the lookup; the backend matches the exact address
const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export function roleLabel(t: RoleTranslator, role: Role): string {
  return t(`role.${ROLE_KEY[role] ?? 'unknown'}` as Parameters<RoleTranslator>[0]);
}

export function createEmptyMemberForm(): MemberFormValue {
  return { email: '', roles: [Role.ROLE_MEMBER] };
}

/** Member (API) and the email of its user -> form value */
export function memberFormOf(m: Member, email: string): MemberFormValue {
  return { email, roles: [...m.roles] };
}

/** Returns an error message, or null when the invite form is valid */
export function validateMemberForm(f: MemberFormValue, t: ValidationTranslator): string | null {
  if (!EMAIL_RE.test(f.email.trim())) {
    return t('email');
  }
  return validateMemberRoles(f.roles, t);
}

/** CreateMemberRequest / UpdateMemberRequest: roles min_items 1 */
export function validateMemberRoles(roles: Role[], t: ValidationTranslator): string | null {
  return roles.length === 0 ? t('rolesRequired') : null;
}

export function MemberFields({
  value,
  onChange,
  idPrefix,
  emailReadOnly = false,
}: {
  value: MemberFormValue;
  onChange: (patch: Partial<MemberFormValue>) => void;
  idPrefix: string;
  emailReadOnly?: boolean;
}) {
  const t = useTranslations('Members');
  const te = useTranslations('Enum');

  const toggleRole = (role: Role, checked: boolean) =>
    onChange({
      roles: checked
        ? ASSIGNABLE_ROLES.filter((r) => r === role || value.roles.includes(r))
        : value.roles.filter((r) => r !== role),
    });

  return (
    <div className="grid gap-6">
      <div className="grid gap-2">
        <Label htmlFor={`${idPrefix}-email`}>{t('email')}</Label>
        <Input
          id={`${idPrefix}-email`}
          type="email"
          placeholder={t('emailPlaceholder')}
          spellCheck={false}
          autoComplete="off"
          readOnly={emailReadOnly}
          disabled={emailReadOnly}
          value={value.email}
          onChange={(e) => onChange({ email: e.target.value })}
        />
        {!emailReadOnly && <p className="text-sm text-muted-foreground">{t('emailDescription')}</p>}
      </div>

      <div className="grid gap-3">
        <Label>{t('roles')}</Label>
        {ASSIGNABLE_ROLES.map((role) => {
          const key = ROLE_KEY[role];
          const id = `${idPrefix}-role-${key}`;
          return (
            <div key={role} className="flex items-start gap-3">
              <Checkbox
                id={id}
                checked={value.roles.includes(role)}
                onCheckedChange={(c) => toggleRole(role, c === true)}
              />
              <div className="grid gap-1">
                <Label htmlFor={id}>{roleLabel(te, role)}</Label>
                <p className="text-sm text-muted-foreground">
                  {te(`roleDescription.${key}` as Parameters<RoleTranslator>[0])}
                </p>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
