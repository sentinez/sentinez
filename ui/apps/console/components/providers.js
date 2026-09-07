'use client';
import * as React from 'react';
import { ThemeProvider as NextThemesProvider } from 'next-themes';
import { Toaster } from '@sentinez/ui/components/sonner';
export function Providers({ children }) {
    return (<NextThemesProvider attribute="class" defaultTheme="light" disableTransitionOnChange enableColorScheme>
      {children}
      <Toaster position="top-right" toastOptions={{
            classNames: {
                toast: 'group toast group-[.toaster]:bg-background/80 group-[.toaster]:backdrop-blur-xl group-[.toaster]:text-foreground group-[.toaster]:border-border/50 group-[.toaster]:shadow-2xl group-[.toaster]:rounded-2xl group-[.toaster]:px-4 group-[.toaster]:py-3 group-[.toaster]:gap-3',
                description: 'group-[.toast]:text-muted-foreground group-[.toast]:text-[0.8rem] leading-relaxed',
                actionButton: 'group-[.toast]:bg-primary group-[.toast]:text-primary-foreground group-[.toast]:font-medium group-[.toast]:rounded-lg group-[.toast]:px-3',
                cancelButton: 'group-[.toast]:bg-muted group-[.toast]:text-muted-foreground group-[.toast]:font-medium group-[.toast]:rounded-lg group-[.toast]:px-3',
                closeButton: 'group-[.toast]:bg-background group-[.toast]:text-foreground group-[.toast]:border-border group-[.toast]:shadow-sm group-[.toast]:rounded-full',
                success: 'group-[.toaster]:border-primary/30 group-[.toaster]:bg-primary/5',
                error: 'group-[.toaster]:border-destructive/30 group-[.toaster]:bg-destructive/5',
                info: 'group-[.toaster]:border-accent/30 group-[.toaster]:bg-accent/5',
                warning: 'group-[.toaster]:border-orange-500/30 group-[.toaster]:bg-orange-500/5',
            },
        }}/>
    </NextThemesProvider>);
}
