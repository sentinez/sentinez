'use client';
import { Card, CardContent, CardDescription, CardHeader, CardTitle, } from '@sentinez/ui/components/card';
import { Badge } from '@sentinez/ui/components/badge';
import { Separator } from '@sentinez/ui/components/separator';
import { getResourceByDomain } from '@/lib/api/tenant';
import { useApi } from '@/hooks/use-api';
import IsLoading from '@sentinez/ui/components/common/loading';
export function ResourceView({ domain }) {
    const { data: resource, isLoading, error } = useApi(getResourceByDomain, domain);
    if (isLoading) {
        return <IsLoading />;
    }
    if (error || !resource) {
        return (<div className="p-8 flex flex-col items-center justify-center text-center space-y-4">
        <div className="text-muted-foreground">{error?.message || 'Resource unavailable'}</div>
        <p className="text-sm text-muted-foreground max-w-xs">
          The domain <strong>{domain}</strong> might be incorrect or have been removed.
        </p>
      </div>);
    }
    return (<div className="flex flex-col gap-6 w-full max-w-6xl p-4 animate-in fade-in duration-500">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">{resource.resourceName}</h1>
          <p className="text-muted-foreground text-sm mt-1">
            {resource.resourceDomain} &mdash; {resource.id}
          </p>
        </div>
        <Badge variant={resource.status === 'STATUS_ACTIVE' ? 'default' : 'secondary'} className="text-sm">
          {resource.status}
        </Badge>
      </div>

      <Separator />

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Overview</CardTitle>
            <CardDescription>General information about the resource</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-3 items-center gap-4">
              <span className="font-semibold text-sm">ID:</span>
              <span className="col-span-2 text-sm break-all">{resource.id}</span>
            </div>
            <div className="grid grid-cols-3 items-center gap-4">
              <span className="font-semibold text-sm">Name:</span>
              <span className="col-span-2 text-sm">{resource.resourceName}</span>
            </div>
            <div className="grid grid-cols-3 items-center gap-4">
              <span className="font-semibold text-sm">Domain:</span>
              <span className="col-span-2 text-sm">{resource.resourceDomain}</span>
            </div>
            <div className="grid grid-cols-3 items-center gap-4">
              <span className="font-semibold text-sm">Plan:</span>
              <span className="col-span-2 text-sm">{resource.plan}</span>
            </div>
            <div className="grid grid-cols-3 items-center gap-4">
              <span className="font-semibold text-sm">Status:</span>
              <span className="col-span-2 text-sm">{resource.status}</span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Metadata</CardTitle>
            <CardDescription>Timestamps and management data</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-3 gap-4">
              <span className="font-semibold text-sm">Created At:</span>
              <span className="col-span-2 text-sm">
                {resource.metadata?.createdAt
            ? new Date(resource.metadata.createdAt).toLocaleString()
            : 'N/A'}
              </span>
            </div>
            <div className="grid grid-cols-3 gap-4">
              <span className="font-semibold text-sm">Updated At:</span>
              <span className="col-span-2 text-sm">
                {resource.metadata?.updatedAt
            ? new Date(resource.metadata.updatedAt).toLocaleString()
            : 'N/A'}
              </span>
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Settings (Raw JSON)</CardTitle>
          <CardDescription>Edge configurations applied to this resource</CardDescription>
        </CardHeader>
        <CardContent>
          {resource.resourceSetting ? (<pre className="bg-muted text-foreground p-4 rounded-md overflow-x-auto text-xs min-h-[300px]">
              {JSON.stringify(resource.resourceSetting, null, 2)}
            </pre>) : (<div className="text-sm text-muted-foreground p-4 border rounded-md">
              No settings applied.
            </div>)}
        </CardContent>
      </Card>
    </div>);
}
