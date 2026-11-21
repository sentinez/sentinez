import Link from 'next/link';
import { notFound } from 'next/navigation';
import type { Metadata } from 'next';
import Markdown, { type Components } from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeSlug from 'rehype-slug';

import { LANGS, type Lang, isLang, listDocs, readDoc, resolveHref } from '@/lib/docs';

interface Params {
  lang: string;
  slug?: string[];
}

export async function generateStaticParams() {
  const out: Params[] = [];
  for (const lang of LANGS) {
    for (const d of await listDocs(lang)) {
      out.push({ lang, slug: d.slug ? [d.slug] : [] });
    }
  }
  return out;
}

export const dynamicParams = false;

export async function generateMetadata({ params }: { params: Promise<Params> }): Promise<Metadata> {
  const { lang, slug } = await params;
  if (!isLang(lang)) return {};
  const src = await readDoc(lang, slug?.[0] ?? '');
  const title = /^#\s+(.+)$/m.exec(src ?? '')?.[1];
  return { title: title ? `${title} — Sentinéz Engineering` : 'Sentinéz Engineering' };
}

const UI = {
  en: { title: 'Engineering', switch: 'Tiếng Việt', home: 'Home' },
  vi: { title: 'Kỹ thuật', switch: 'English', home: 'Trang chủ' },
} satisfies Record<Lang, object>;

function makeComponents(lang: Lang): Components {
  return {
    h1: (p) => <h1 className="mb-6 text-3xl font-semibold tracking-tight" {...p} />,
    h2: (p) => (
      <h2 className="mt-10 mb-4 border-b border-black/10 pb-2 text-2xl font-semibold" {...p} />
    ),
    h3: (p) => <h3 className="mt-8 mb-3 text-xl font-semibold" {...p} />,
    h4: (p) => <h4 className="mt-6 mb-2 font-semibold" {...p} />,
    p: (p) => <p className="my-4 leading-7" {...p} />,
    ul: (p) => <ul className="my-4 list-disc space-y-1 pl-6" {...p} />,
    ol: (p) => <ol className="my-4 list-decimal space-y-1 pl-6" {...p} />,
    blockquote: (p) => (
      <blockquote className="my-4 border-l-4 border-[#0057ff]/40 pl-4 text-neutral-600" {...p} />
    ),
    hr: () => <hr className="my-8 border-black/10" />,
    pre: (p) => (
      <pre
        className="my-4 overflow-x-auto rounded-lg bg-neutral-950 p-4 text-sm text-neutral-100 [&_code]:bg-transparent [&_code]:p-0 [&_code]:text-inherit"
        {...p}
      />
    ),
    code: (p) => (
      <code className="rounded bg-black/[0.06] px-1.5 py-0.5 font-mono text-[0.85em]" {...p} />
    ),
    table: (p) => (
      <div className="my-4 overflow-x-auto">
        <table className="w-full border-collapse text-sm" {...p} />
      </div>
    ),
    th: (p) => (
      <th className="border border-black/10 bg-black/[0.04] px-3 py-2 text-left" {...p} />
    ),
    td: (p) => <td className="border border-black/10 px-3 py-2 align-top" {...p} />,
    a: ({ href = '', ...p }) => {
      const to = resolveHref(href, lang);
      const cls = 'text-[#0057ff] underline-offset-2 hover:underline';
      return to.startsWith('/') ? (
        <Link href={to} className={cls} {...p} />
      ) : (
        <a href={to} className={cls} {...(to.startsWith('#') ? {} : { target: '_blank', rel: 'noreferrer' })} {...p} />
      );
    },
  };
}

export default async function Page({ params }: { params: Promise<Params> }) {
  const { lang, slug } = await params;
  if (!isLang(lang) || (slug?.length ?? 0) > 1) notFound();
  const page = slug?.[0] ?? '';

  const [src, docs] = await Promise.all([readDoc(lang, page), listDocs(lang)]);
  if (src === null) notFound();

  const t = UI[lang];
  const other: Lang = lang === 'en' ? 'vi' : 'en';

  return (
    <div className="min-h-svh bg-white text-[#0a0a0a]">
      <header className="sticky top-0 z-10 flex items-center justify-between border-b border-black/[0.08] bg-white/90 px-4 py-3 backdrop-blur md:px-8">
        <Link href="/" className="text-sm font-semibold tracking-wide uppercase">
          SENTINÉZ <span className="text-[#888]">/ {t.title}</span>
        </Link>
        <Link
          href={`/engineering/${other}${page ? `/${page}` : ''}`}
          className="text-xs tracking-widest text-[#0057ff] uppercase"
        >
          {t.switch}
        </Link>
      </header>
      <div className="mx-auto flex max-w-6xl flex-col gap-8 px-4 py-8 md:flex-row md:px-8">
        <nav aria-label={t.title} className="shrink-0 md:sticky md:top-20 md:w-64 md:self-start">
          <ul className="space-y-1 text-sm">
            {docs.map((d) => (
              <li key={d.slug}>
                <Link
                  href={`/engineering/${lang}${d.slug ? `/${d.slug}` : ''}`}
                  className={`block rounded px-2 py-1 hover:bg-black/[0.05] ${
                    d.slug === page ? 'bg-[#0057ff]/[0.08] font-medium text-[#0057ff]' : ''
                  }`}
                >
                  {d.title}
                </Link>
              </li>
            ))}
          </ul>
        </nav>
        <article className="min-w-0 flex-1">
          <Markdown
            remarkPlugins={[remarkGfm]}
            rehypePlugins={[rehypeSlug]}
            components={makeComponents(lang)}
          >
            {src}
          </Markdown>
        </article>
      </div>
    </div>
  );
}
