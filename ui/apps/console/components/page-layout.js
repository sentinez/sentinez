export default function PageLayout({ children }) {
    return (<div className=" flex justify-center">
      <div className=" w-full">{children}</div>
    </div>);
}
