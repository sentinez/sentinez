'use server';

export default async function Page() {
  await new Promise((resolve) => setTimeout(resolve, 1500));

  return (
    <>
      <div>access-log</div>
    </>
  );
}
