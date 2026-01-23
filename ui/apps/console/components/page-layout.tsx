export default function PageLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <div className=" flex justify-center">
      <div className=" w-full">{children}</div>
    </div>
  );
}
