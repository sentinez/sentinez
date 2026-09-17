'use client';

import Image from 'next/image';
import { useEffect, useState } from 'react';
import { AnimatePresence, motion } from 'motion/react';

const TYPEWRITER_TEXT = 'INTELLIGENT SECURITY FOR THE MODERN WEB.';

const MULTI_LANG_TEXTS = [
  'On Your Side', // US — English
  'Luôn Bên Bạn', // VN — Vietnamese
  '站在你这边', // CN — Chinese
  'На твоей стороне', // RU — Russian
  'आपके साथ', // IN — Hindi
  '당신의 편에서', // KR — Korean
  'あなたの味方', // JP — Japanese
  'À vos côtés', // FR — French
  'Di Sisi Anda', // MY — Malay
  'อยู่เคียงข้างคุณ', // TH — Thai
  'An Ihrer Seite', // EU — German
];

const CARDS = [
  {
    label: 'WEB APPLICATION FIREWALL',
    title: 'PROTECT',
    link: '/dashboard',
    linkLabel: 'GET STARTED ↗',
  },
  {
    label: 'REAL-TIME THREAT DETECTION',
    title: 'DETECT',
    link: '/threats',
    linkLabel: 'VIEW THREATS ↗',
  },
  {
    label: 'OPEN PLATFORM',
    title: 'GITHUB',
    link: 'https://github.com/sentinez',
    linkLabel: '/SENTINEZ ↗',
  },
];

export default function Page() {
  const [displayed, setDisplayed] = useState('');
  const [cursorVisible, setCursorVisible] = useState(true);
  const [langIndex, setLangIndex] = useState(0);

  useEffect(() => {
    let i = 0;
    const interval = setInterval(() => {
      if (i < TYPEWRITER_TEXT.length) {
        setDisplayed(TYPEWRITER_TEXT.slice(0, i + 1));
        i++;
      } else {
        clearInterval(interval);
      }
    }, 38);
    return () => clearInterval(interval);
  }, []);

  useEffect(() => {
    const blink = setInterval(() => {
      setCursorVisible((v) => !v);
    }, 530);
    return () => clearInterval(blink);
  }, []);

  useEffect(() => {
    const langTimer = setInterval(() => {
      setLangIndex((prev) => (prev + 1) % MULTI_LANG_TEXTS.length);
    }, 2500);
    return () => clearInterval(langTimer);
  }, []);

  return (
    <>
      {/* Phudu font must be loaded via @import */}
      <style>{`@import url('https://fonts.googleapis.com/css2?family=Phudu:wght@400;600;700;900&display=swap');`}</style>

      {/* Root */}
      <div className="bg-white text-[#0a0a0a] min-h-svh flex flex-col overflow-x-hidden [font-family:'Phudu',sans-serif]">
        {/* Header — logo | card nav | studio */}
        <header className="flex flex-col md:flex-row items-stretch justify-between border-b border-black/[0.08]">
          {/* Top bar on mobile (Logo + Studio stamp) */}
          <div className="flex items-center justify-between w-full md:w-auto shrink-0 border-b md:border-b-0 border-black/[0.08]">
            {/* Logo */}
            <div className="flex items-center px-4 py-3 md:px-8 md:py-6 shrink-0 select-none md:border-r border-black/[0.08]">
              <Image
                src="/sntz.png"
                alt="Sentinez"
                width={36}
                height={36}
                className="md:w-[44px] md:h-[44px]"
              />
            </div>

            {/* Studio stamp — mobile top-right */}
            <div className="flex md:hidden items-center px-4 py-3 text-right text-[0.55rem] tracking-[0.15em] text-[#888] leading-[1.5] uppercase border-l border-black/[0.08]">
              Security Platform
              <br />
              Est. 2025
            </div>
          </div>

          {/* Cards nav — fills remaining space */}
          <nav aria-label="Products" className="flex flex-col md:flex-row flex-1 w-full md:w-auto">
            {CARDS.map((card, i) => (
              <a
                key={card.title}
                href={card.link}
                aria-label={card.title}
                className={`group flex flex-col justify-between flex-1 px-4 py-3.5 md:px-6 md:py-5 no-underline text-inherit transition-colors duration-200 hover:bg-[#0057ff]/[0.04] border-b md:border-b-0 ${
                  i < CARDS.length - 1
                    ? 'border-black/[0.08] md:border-r'
                    : 'border-black/[0.08] md:border-r-0'
                }`}
              >
                <span className="text-[0.5rem] tracking-[0.15em] md:tracking-[0.2em] text-[#aaa] uppercase mb-1.5 md:mb-2">
                  {card.label}
                </span>
                <div className="flex items-end justify-between gap-2">
                  <span className="text-sm md:text-base font-semibold tracking-[-0.02em] uppercase leading-none">
                    {card.title}
                  </span>
                  <span className="text-[0.55rem] tracking-[0.15em] text-[#888] group-hover:text-[#0057ff] uppercase whitespace-nowrap transition-colors duration-200">
                    {card.linkLabel}
                  </span>
                </div>
              </a>
            ))}
          </nav>

          {/* Studio stamp — desktop only */}
          <div className="hidden md:flex items-start px-8 py-6 shrink-0 border-l border-black/[0.08] text-right text-[0.6rem] tracking-[0.18em] text-[#888] leading-[1.7] uppercase">
            Security Platform
            <br />
            Est. 2025
          </div>
        </header>

        {/* Hero */}
        <main className="flex-1 flex flex-col items-center justify-center text-center px-4 py-8 md:px-6 md:py-12">
          <h1 className="[font-family:'Phudu',sans-serif] text-[clamp(2.5rem,8vw,8rem)] font-semibold leading-[0.95] tracking-[-0.03em] uppercase flex flex-col items-center max-w-full">
            <span className="text-[#0504aa]">SENTINÉZ</span>
            <span className="relative h-[1.3em] w-full flex items-center justify-center overflow-hidden max-w-full">
              <AnimatePresence mode="wait">
                <motion.span
                  key={langIndex}
                  initial={{ y: 25, opacity: 0 }}
                  animate={{ y: 0, opacity: 1 }}
                  exit={{ y: -25, opacity: 0 }}
                  transition={{ duration: 0.4, ease: 'easeInOut' }}
                  className="text-[#0a0a0a] text-[clamp(1.5rem,5.5vw,5.5rem)] whitespace-nowrap block"
                >
                  {MULTI_LANG_TEXTS[langIndex]}
                </motion.span>
              </AnimatePresence>
            </span>
          </h1>
          <p className="mt-6 md:mt-8 [font-family:'Phudu',sans-serif] text-[clamp(0.6rem,1.5vw,0.85rem)] tracking-[0.15em] md:tracking-[0.2em] text-[#999] uppercase min-h-[1.4em] max-w-full px-2 break-words">
            {displayed}
            <span
              aria-hidden="true"
              className={`inline-block w-[0.65em] h-[0.9em] bg-[#0057ff] ml-0.5 align-middle transition-opacity duration-100 ${cursorVisible ? 'opacity-100' : 'opacity-0'}`}
            />
          </p>
        </main>
      </div>
    </>
  );
}
