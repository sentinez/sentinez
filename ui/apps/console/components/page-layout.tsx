export default function PageLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return <div className="flex flex-col gap-6 max-w-7xl p-4">{children}</div>;
}
