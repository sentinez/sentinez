/** Holidays the loading screen decorates, in calendar order. */
export type HolidayId =
  | 'newYear'
  | 'tet'
  | 'womenDay'
  | 'hungKings'
  | 'reunification'
  | 'childrenDay'
  | 'nationalDay'
  | 'midAutumn'
  | 'vietnameseWomenDay'
  | 'teachersDay'
  | 'christmas';

export type Holiday = {
  id: HolidayId;
  /** Tailwind text color of the greeting. */
  color: string;
};

type MonthDay = { month: number; day: number };

type Rule = Holiday & {
  /** Solar (Gregorian) days, in Vietnam time. */
  solar?: (d: MonthDay) => boolean;
  /** Lunar days, in Vietnam time; leap months never match. */
  lunar?: (d: MonthDay) => boolean;
};

const _rules: Rule[] = [
  {
    id: 'tet',
    color: 'text-red-600',
    // From the 28th of the last lunar month to the 7th of the first.
    lunar: ({ month, day }) => (month === 12 && day >= 28) || (month === 1 && day <= 7),
  },
  {
    id: 'newYear',
    color: 'text-amber-500',
    solar: ({ month, day }) => (month === 12 && day === 31) || (month === 1 && day === 1),
  },
  {
    id: 'womenDay',
    color: 'text-rose-500',
    solar: ({ month, day }) => month === 3 && day === 8,
  },
  {
    id: 'hungKings',
    color: 'text-amber-600',
    lunar: ({ month, day }) => month === 3 && day === 10,
  },
  {
    id: 'reunification',
    color: 'text-red-600',
    solar: ({ month, day }) => (month === 4 && day === 30) || (month === 5 && day === 1),
  },
  {
    id: 'childrenDay',
    color: 'text-sky-500',
    solar: ({ month, day }) => month === 6 && day === 1,
  },
  {
    id: 'nationalDay',
    color: 'text-red-600',
    solar: ({ month, day }) => month === 9 && day >= 1 && day <= 2,
  },
  {
    id: 'midAutumn',
    color: 'text-orange-500',
    lunar: ({ month, day }) => month === 8 && day >= 14 && day <= 15,
  },
  {
    id: 'vietnameseWomenDay',
    color: 'text-pink-500',
    solar: ({ month, day }) => month === 10 && day === 20,
  },
  {
    id: 'teachersDay',
    color: 'text-emerald-600',
    solar: ({ month, day }) => month === 11 && day === 20,
  },
  {
    id: 'christmas',
    color: 'text-red-600',
    solar: ({ month, day }) => month === 12 || (month === 11 && day >= 24),
  },
];

const _timeZone = 'Asia/Ho_Chi_Minh';

const _solarFormat = new Intl.DateTimeFormat('en-US', {
  timeZone: _timeZone,
  month: 'numeric',
  day: 'numeric',
});

/** The ICU Chinese calendar: it matches the Vietnamese one but in rare years. */
const _lunarFormat = new Intl.DateTimeFormat('en-US-u-ca-chinese', {
  timeZone: _timeZone,
  month: 'numeric',
  day: 'numeric',
});

/** A holiday shows from this many days before it. */
const _leadDays = 7;
const _dayMillis = 24 * 60 * 60 * 1000;

/**
 * holidayOn returns the holiday of the given time in Vietnam, or the nearest
 * one within the week after it. Lunar holidays come first, so Tết wins over
 * the solar New Year.
 */
export function holidayOn(date: Date): Holiday | undefined {
  for (let ahead = 0; ahead <= _leadDays; ahead++) {
    const holiday = holidayOfDay(new Date(date.getTime() + ahead * _dayMillis));
    if (holiday) return holiday;
  }

  return undefined;
}

/** holidayById looks a holiday up by id, for previews (`?holiday=tet`). */
export function holidayById(id: string | null): Holiday | undefined {
  return _rules.find((r) => r.id === id);
}

/** holidayOfDay returns the holiday falling on the given day, if any. */
function holidayOfDay(date: Date): Holiday | undefined {
  const solar = monthDay(_solarFormat, date);
  const lunar = monthDay(_lunarFormat, date);
  return _rules.find((r) => (lunar && r.lunar?.(lunar)) || (solar && r.solar?.(solar)));
}

function monthDay(format: Intl.DateTimeFormat, date: Date): MonthDay | undefined {
  const parts = format.formatToParts(date);
  const month = parts.find((p) => p.type === 'month')?.value ?? '';
  const day = parts.find((p) => p.type === 'day')?.value ?? '';
  // A leap lunar month reads like "4bis": no holiday falls in one.
  if (!/^\d+$/.test(month) || !/^\d+$/.test(day)) return undefined;

  return { month: Number(month), day: Number(day) };
}
