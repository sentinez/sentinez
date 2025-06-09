# sentinez/ui monorepo

User interface of sentinez project

## Usage

```bash
pnpm run dev
```

## Adding components

To add components to your app, run the following command at the root of your `console` app:

```bash
pnpm dlx shadcn@latest add button
```

This will place the ui components in the `packages/ui/src/components` directory.

## Tailwind

Your `globals.css` are already set up to use the components from the `ui` package.

## Using components

To use the components in your app, import them from the `ui` package.

```tsx
import { Button } from '@sentinez/ui/components/button';
```
