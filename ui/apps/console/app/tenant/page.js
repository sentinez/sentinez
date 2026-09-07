import { Tenant, TeamJoin } from './components/card';
export default async function Page() {
    await new Promise((resolve) => setTimeout(resolve, 2000));
    return (<div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2">
      <div className="flex gap-10">
        <Tenant />
        <TeamJoin />
      </div>
    </div>);
}
