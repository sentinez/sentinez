/**
 * The two 44-character lines of a passport machine readable zone (ICAO 9303
 * TD3), filled from the account so the data page reads like a real passport.
 */

const LINE = 44;
const ISSUER = 'SNZ';
const WEIGHTS = [7, 3, 1];

/** Latin letters and digits only; everything else becomes the filler '<' */
function mrzText(value: string) {
  return value
    .normalize('NFD')
    .replace(/[̀-ͯ]/g, '')
    .replace(/[đĐ]/g, 'D')
    .toUpperCase()
    .replace(/[^A-Z0-9]+/g, '<');
}

function pad(value: string, length: number) {
  return value.slice(0, length).padEnd(length, '<');
}

/** ICAO check digit: weights 7-3-1 over digits, A=10..Z=35, filler=0 */
function checkDigit(value: string) {
  let sum = 0;
  for (let i = 0; i < value.length; i++) {
    const c = value.charCodeAt(i);
    let n = 0;
    if (c >= 48 && c <= 57) n = c - 48;
    else if (c >= 65 && c <= 90) n = c - 55;
    sum += n * (WEIGHTS[i % 3] ?? 0);
  }
  return String(sum % 10);
}

function yymmdd(date?: Date) {
  if (!date) return '<<<<<<';
  const two = (n: number) => String(n).padStart(2, '0');
  return two(date.getFullYear() % 100) + two(date.getMonth() + 1) + two(date.getDate());
}

/** The passport number shown on the page: the first 9 id characters */
export function passportNumberOf(id: string) {
  return pad(mrzText(id).replace(/</g, ''), 9).replace(/</g, '0');
}

/**
 * The surname is the first word of the full name (Vietnamese order, e.g.
 * "HO DUC HUNG" -> HO<<DUC<HUNG); the given names are the rest.
 */
export function mrzLinesOf({
  name,
  id,
  issuedAt,
  updatedAt,
}: Readonly<{ name: string; id: string; issuedAt?: Date; updatedAt?: Date }>) {
  const words = mrzText(name).split('<').filter(Boolean);
  const surname = words.slice(0, 1).join('<');
  const given = words.slice(1).join('<');
  const line1 = pad(`P<${ISSUER}${surname}<<${given}`, LINE);

  const doc = passportNumberOf(id);
  const issued = yymmdd(issuedAt);
  const updated = yymmdd(updatedAt);
  const personal = pad(mrzText(id).replace(/</g, '').slice(9), 14);
  const docPart = doc + checkDigit(doc);
  const issuedPart = issued + checkDigit(issued);
  const updatedPart = updated + checkDigit(updated);
  const personalPart = personal + checkDigit(personal);
  const composite = checkDigit(docPart + issuedPart + updatedPart + personalPart);
  const line2 = `${docPart}${ISSUER}${issuedPart}<${updatedPart}${personalPart}${composite}`;

  return [line1, line2] as const;
}
