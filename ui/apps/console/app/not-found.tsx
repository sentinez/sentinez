import Image from 'next/image';

export default function NotFound() {
  return (
    <div className="flex items-center justify-center h-screen">
      <div className="flex gap-2 items-center">
        <Image src="/assets/cels.png" alt="sen" width={150} height={150} />
        <div className=" flex flex-col gap-2">
          <div className="font-medium text-7xl font-mono mx-2">404</div>
          <div className="font-medium text-2xl font-mono mx-2">Not Found!</div>
        </div>
      </div>
    </div>
  );
}
