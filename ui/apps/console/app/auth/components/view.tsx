'use client';

import { LoginForm } from '@/app/auth/components/login-form';
import Image from 'next/image';
import { ReactNode, useEffect, useState } from 'react';
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

export default function View({ children }: { children: ReactNode }) {
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
    <div className="grid min-h-svh lg:grid-cols-2">
      <div className="flex flex-col gap-4 p-6 md:p-10">
        <div className="flex justify-center gap-2 md:justify-start">
          <a
            href="#"
            className="flex items-center gap-1 font-bold [font-family:'Phudu',sans-serif]"
          >
            <div className="flex size-6 items-center justify-center rounded-md text-primary-foreground">
              <Image width={600} height={600} src="/assets/sntz.png" alt="Image" loading="eager" />
            </div>
            <span className="text-[#0504aa] items-center text-xl">SENTINÉZ</span>
          </a>
        </div>
        <div className="flex flex-1 items-center justify-center">
          <div className="w-full max-w-xs">{children}</div>
        </div>
      </div>
      <div className="relative hidden lg:flex animated-gradient">
        <main className="flex-1 flex flex-col items-center justify-center text-center px-4 py-8 md:px-6 md:py-12">
          <h1 className="[font-family:'Phudu',sans-serif] text-[clamp(2.5rem,8vw,8rem)] font-semibold leading-[0.95] tracking-[-0.03em] uppercase flex flex-col items-center max-w-full">
            <span className=" text-white">SENTINÉZ</span>
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
          <p className="mt-6 md:mt-8 [font-family:'Phudu',sans-serif] text-[clamp(0.6rem,1.5vw,0.85rem)] tracking-[0.15em] md:tracking-[0.2em] text-black uppercase min-h-[1.4em] max-w-full px-2 break-words">
            {displayed}
            <span
              aria-hidden="true"
              className={`inline-block w-[0.65em] h-[0.9em] bg-black ml-0.5 align-middle transition-opacity duration-100 ${cursorVisible ? 'opacity-100' : 'opacity-0'}`}
            />
          </p>
        </main>
      </div>
    </div>
  );
}
