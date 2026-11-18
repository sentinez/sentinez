'use client';

import { motion, useReducedMotion } from 'motion/react';
import { useMemo } from 'react';

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
  /** Particles drawn at random, emoji so no asset is needed. */
  particles: string[];
  /** fall: petals, snow; rise: lanterns, balloons, fireworks. */
  drift: 'fall' | 'rise';
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
    particles: ['🌼', '🌸', '🧧', '🌼', '🌸'],
    drift: 'fall',
    color: 'text-red-600',
    // From the 28th of the last lunar month to the 7th of the first.
    lunar: ({ month, day }) => (month === 12 && day >= 28) || (month === 1 && day <= 7),
  },
  {
    id: 'newYear',
    particles: ['🎆', '🎇', '✨', '🎉'],
    drift: 'rise',
    color: 'text-amber-500',
    solar: ({ month, day }) => (month === 12 && day === 31) || (month === 1 && day === 1),
  },
  {
    id: 'womenDay',
    particles: ['🌹', '💐', '🌷', '💝'],
    drift: 'fall',
    color: 'text-rose-500',
    solar: ({ month, day }) => month === 3 && day === 8,
  },
  {
    id: 'hungKings',
    particles: ['🥁', '🌾', '🐉', '🏯'],
    drift: 'fall',
    color: 'text-amber-600',
    lunar: ({ month, day }) => month === 3 && day === 10,
  },
  {
    id: 'reunification',
    particles: ['⭐', '🎉', '🎊', '⭐'],
    drift: 'fall',
    color: 'text-red-600',
    solar: ({ month, day }) => (month === 4 && day === 30) || (month === 5 && day === 1),
  },
  {
    id: 'childrenDay',
    particles: ['🎈', '🧸', '🍭', '🎈', '🪁'],
    drift: 'rise',
    color: 'text-sky-500',
    solar: ({ month, day }) => month === 6 && day === 1,
  },
  {
    id: 'nationalDay',
    particles: ['⭐', '🎆', '🎇', '⭐'],
    drift: 'rise',
    color: 'text-red-600',
    solar: ({ month, day }) => month === 9 && day >= 1 && day <= 2,
  },
  {
    id: 'midAutumn',
    particles: ['🏮', '🥮', '🌕', '🏮', '⭐'],
    drift: 'rise',
    color: 'text-orange-500',
    lunar: ({ month, day }) => month === 8 && day >= 14 && day <= 15,
  },
  {
    id: 'vietnameseWomenDay',
    particles: ['💐', '🌷', '🌹', '💝'],
    drift: 'fall',
    color: 'text-pink-500',
    solar: ({ month, day }) => month === 10 && day === 20,
  },
  {
    id: 'teachersDay',
    particles: ['💐', '📚', '✏️', '🌻', '🎓'],
    drift: 'fall',
    color: 'text-emerald-600',
    solar: ({ month, day }) => month === 11 && day === 20,
  },
  {
    id: 'christmas',
    particles: ['❄️', '❄️', '🎄', '🎁', '⛄'],
    drift: 'fall',
    color: 'text-green-600',
    solar: ({ month, day }) => month === 12 && day >= 24 && day <= 25,
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

/** HolidayEffect draws the particles of a holiday over the whole screen. */
export function HolidayEffect({ holiday, count = 22 }: { holiday: Holiday; count?: number }) {
  const reduceMotion = useReducedMotion();
  const particles = useMemo(
    () =>
      Array.from({ length: count }, (_, i) => ({
        key: i,
        glyph: holiday.particles[i % holiday.particles.length],
        left: Math.random() * 100,
        size: 16 + Math.random() * 18,
        delay: Math.random() * 1.2,
        duration: 3 + Math.random() * 3,
        sway: (Math.random() - 0.5) * 80,
        spin: (Math.random() - 0.5) * 360,
      })),
    [holiday, count],
  );

  if (reduceMotion) return null;

  const [from, to] = holiday.drift === 'fall' ? ['-10vh', '110vh'] : ['110vh', '-10vh'];

  return (
    <div aria-hidden className="pointer-events-none absolute inset-0 overflow-hidden">
      {particles.map((p) => (
        <motion.span
          key={p.key}
          className="absolute top-0 select-none"
          style={{ left: `${p.left}%`, fontSize: p.size }}
          initial={{ y: from, x: 0, rotate: 0, opacity: 0 }}
          animate={{
            y: to,
            x: [0, p.sway, 0],
            rotate: holiday.drift === 'fall' ? p.spin : 0,
            opacity: [0, 1, 1, 0],
          }}
          transition={{
            duration: p.duration,
            delay: p.delay,
            ease: 'linear',
            repeat: Infinity,
          }}
        >
          {p.glyph}
        </motion.span>
      ))}
    </div>
  );
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
