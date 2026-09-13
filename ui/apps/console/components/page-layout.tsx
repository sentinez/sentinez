export default function PageLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return <div className="flex flex-col gap-6 w-full">{children}</div>;
}
