// @ts-check
import react from '@astrojs/react';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'astro/config';

// https://astro.build/config
export default defineConfig({
  integrations: [react()],
  vite: {
    plugins: [tailwindcss()],
    // @sentinez/ui ships .tsx sources: bundle them instead of importing them in node
    ssr: { noExternal: ['@sentinez/ui'] },
  },
});
