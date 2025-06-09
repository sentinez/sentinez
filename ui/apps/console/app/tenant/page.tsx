import { Tenant, TenantJoin } from './components/card';

export default async function Page() {
  return (
    <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2">
      <div className="flex gap-10">
        <Tenant />
        <TenantJoin />
      </div>
    </div>
  );
}
