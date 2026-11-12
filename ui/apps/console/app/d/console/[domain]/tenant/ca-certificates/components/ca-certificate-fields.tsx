'use client';

import { useId, useRef } from 'react';
import { useTranslations } from 'next-intl';
import { Upload } from 'lucide-react';
import { CaCertificateUsage } from '@sentinez/proto/sentinez/apps/tenant/v1/model';
import type { CreateCaCertificateRequest } from '@sentinez/proto/sentinez/apps/tenant/v1/tenant';
import { Button } from '@sentinez/ui/components/button';
import { Input } from '@sentinez/ui/components/input';
import { Label } from '@sentinez/ui/components/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@sentinez/ui/components/select';
import { Textarea } from '@sentinez/ui/components/textarea';
import { CA_USAGES, caUsageHelp } from './ca-certificate-status';

type ValidationTranslator = ReturnType<typeof useTranslations<'Validation'>>;

// Limits of the proto request (buf.validate): name and description are in
// characters, the PEM texts in bytes
const MAX_NAME = 120;
const MAX_DESCRIPTION = 255;
const MAX_PEM_BYTES = 262144;
const MAX_KEY_BYTES = 65536;

const CERTIFICATE_BEGIN = '-----BEGIN CERTIFICATE-----';

export interface CaCertificateFormValue {
  name: string;
  description: string;
  usage: CaCertificateUsage;
  pem: string;
  /** Optional; required for a Server certificate, may also be in the pem */
  privateKey: string;
}

export function createEmptyCaCertificateForm(): CaCertificateFormValue {
  return {
    name: '',
    description: '',
    usage: CaCertificateUsage.CA_CERTIFICATE_USAGE_UPSTREAM,
    pem: '',
    privateKey: '',
  };
}

/** Form value -> the request that creates the certificates of the resource */
export function caCertificateRequestOf(
  form: CaCertificateFormValue,
  resourceId: string,
): CreateCaCertificateRequest {
  return {
    resourceId,
    name: form.name.trim(),
    description: form.description.trim(),
    usage: form.usage,
    pem: form.pem.trim(),
    privateKey: form.privateKey.trim(),
  };
}

const byteLength = (s: string) => new TextEncoder().encode(s).length;

/** Mirrors the backend checks; returns the error message, or null when valid */
export function validateCaCertificateForm(
  form: CaCertificateFormValue,
  tv: ValidationTranslator,
): string | null {
  const pem = form.pem.trim();
  const key = form.privateKey.trim();
  // the key may come in the bundle itself
  const hasKeyBlock = /-----BEGIN [A-Z ]*PRIVATE KEY-----/.test(pem);

  if (form.name.trim().length > MAX_NAME) return tv('caNameTooLong');
  if (form.description.trim().length > MAX_DESCRIPTION) return tv('caDescriptionTooLong');
  if (!pem.includes(CERTIFICATE_BEGIN)) return tv('caPemRequired');
  if (byteLength(pem) > MAX_PEM_BYTES) return tv('caPemTooLarge');
  if (byteLength(key) > MAX_KEY_BYTES) return tv('caKeyTooLarge');
  // the certificate the edge serves is of no use without its key
  if (form.usage === CaCertificateUsage.CA_CERTIFICATE_USAGE_SERVER && !key && !hasKeyBlock) {
    return tv('caServerNeedsKey');
  }

  return null;
}

/** A textarea with a button that fills it from a file */
function PemField({
  id,
  label,
  help,
  value,
  accept,
  placeholder,
  onChange,
}: {
  id: string;
  label: string;
  help: string;
  value: string;
  accept: string;
  placeholder: string;
  onChange: (value: string) => void;
}) {
  const t = useTranslations('CaCertificate');
  const input = useRef<HTMLInputElement>(null);

  const load = async (file: File | undefined) => {
    if (file) onChange(await file.text());
    // the same file can be picked again
    if (input.current) input.current.value = '';
  };

  return (
    <div className="grid gap-2">
      <div className="flex items-center justify-between gap-2">
        <Label htmlFor={id}>{label}</Label>
        <Button type="button" variant="outline" size="sm" onClick={() => input.current?.click()}>
          <Upload className="w-4 h-4" />
          {t('uploadFile')}
        </Button>
        <input
          ref={input}
          type="file"
          accept={accept}
          className="hidden"
          onChange={(e) => load(e.target.files?.[0])}
        />
      </div>
      <Textarea
        id={id}
        value={value}
        placeholder={placeholder}
        spellCheck={false}
        autoComplete="off"
        className="min-h-40 font-mono text-xs"
        onChange={(e) => onChange(e.target.value)}
      />
      <p className="text-sm text-muted-foreground">{help}</p>
    </div>
  );
}

export function CaCertificateFields({
  value,
  onChange,
}: {
  value: CaCertificateFormValue;
  onChange: (patch: Partial<CaCertificateFormValue>) => void;
}) {
  const t = useTranslations('CaCertificate');
  const idPrefix = useId();

  return (
    <div className="grid gap-6">
      <div className="grid gap-2">
        <Label htmlFor={`${idPrefix}-name`}>{t('name')}</Label>
        <Input
          id={`${idPrefix}-name`}
          value={value.name}
          maxLength={MAX_NAME}
          onChange={(e) => onChange({ name: e.target.value })}
        />
      </div>

      <div className="grid gap-2">
        <Label htmlFor={`${idPrefix}-description`}>{t('description')}</Label>
        <Textarea
          id={`${idPrefix}-description`}
          value={value.description}
          maxLength={MAX_DESCRIPTION}
          onChange={(e) => onChange({ description: e.target.value })}
        />
      </div>

      <div className="grid gap-2">
        <Label>{t('usage')}</Label>
        <Select
          value={String(value.usage)}
          onValueChange={(v) => onChange({ usage: Number(v) as CaCertificateUsage })}
        >
          <SelectTrigger className="w-full">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {CA_USAGES.map((u) => (
              <SelectItem key={u.value} value={String(u.value)}>
                {t(u.label)}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <p className="text-sm text-muted-foreground">{caUsageHelp(t, value.usage)}</p>
      </div>

      <PemField
        id={`${idPrefix}-pem`}
        label={t('pem')}
        help={t('pemHelp')}
        value={value.pem}
        accept=".pem,.crt,.cer"
        placeholder={`${CERTIFICATE_BEGIN}\n...\n-----END CERTIFICATE-----`}
        onChange={(pem) => onChange({ pem })}
      />

      <PemField
        id={`${idPrefix}-private-key`}
        label={t('privateKeyOptional')}
        help={t('privateKeyHelp')}
        value={value.privateKey}
        accept=".key,.pem"
        placeholder={'-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----'}
        onChange={(privateKey) => onChange({ privateKey })}
      />
    </div>
  );
}
