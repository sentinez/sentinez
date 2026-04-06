export default async function Page() {
  await new Promise((resolve) => setTimeout(resolve, 2000));

  return (
    <>
      {Array.from({ length: 24 }).map((_, index) => (
        <div key={index} className="bg-muted/50 aspect-video h-12 w-full rounded-lg" />
      ))}
    </>
  );
}
