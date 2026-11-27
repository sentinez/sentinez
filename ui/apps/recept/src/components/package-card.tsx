import { Badge } from '@sentinez/ui/components/badge';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@sentinez/ui/components/card';

import { formatDate, KINDS, type PackageSummary } from '@/lib/summary';

/** A package of the home page, the whole card linking to it */
export function PackageCard({ pkg }: { pkg: PackageSummary }) {
  return (
    <Card size="sm" className="relative transition-colors hover:bg-muted/40">
      <CardHeader>
        <CardTitle className="font-mono">
          <a href={`/pkg/${pkg.path}`} className="after:absolute after:inset-0">
            {pkg.path}
          </a>
        </CardTitle>
        <CardDescription className="line-clamp-2">{pkg.synopsis}</CardDescription>
      </CardHeader>
      <CardContent className="flex flex-wrap gap-1.5">
        {pkg.kinds.map(({ kind, count }) => (
          <Badge key={kind} variant="secondary">
            {count} {KINDS[kind].label}
          </Badge>
        ))}
      </CardContent>
      <CardFooter className="mt-auto gap-3 text-xs text-muted-foreground">
        {pkg.version && <span className="font-mono">{pkg.version}</span>}
        <span>Updated {formatDate(pkg.updated)}</span>
        <span className="ml-auto font-mono">{pkg.files.join(', ')}</span>
      </CardFooter>
    </Card>
  );
}
