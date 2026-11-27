import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@sentinez/ui/components/breadcrumb';
import { Fragment } from 'react';

/**
 * Discover › each folder of `path` › its last part. A folder links to its
 * package when there is one (`known`), else to the search of its name.
 */
export function PathBreadcrumb({ path, known, file }: { path: string; known: Set<string>; file?: string }) {
  const parts = path.split('/');
  return (
    <Breadcrumb>
      <BreadcrumbList>
        <BreadcrumbItem>
          <BreadcrumbLink href="/">Discover</BreadcrumbLink>
        </BreadcrumbItem>
        {parts.map((part, i) => {
          const sub = parts.slice(0, i + 1).join('/');
          const last = i === parts.length - 1 && !file;
          return (
            <Fragment key={sub}>
              <BreadcrumbSeparator />
              <BreadcrumbItem>
                {last ? (
                  <BreadcrumbPage className="font-mono">{part}</BreadcrumbPage>
                ) : (
                  <BreadcrumbLink
                    href={known.has(sub) ? `/pkg/${sub}` : `/search?q=${encodeURIComponent(sub)}`}
                    className="font-mono"
                  >
                    {part}
                  </BreadcrumbLink>
                )}
              </BreadcrumbItem>
            </Fragment>
          );
        })}
        {file && (
          <>
            <BreadcrumbSeparator />
            <BreadcrumbItem>
              <BreadcrumbPage className="font-mono">{file}</BreadcrumbPage>
            </BreadcrumbItem>
          </>
        )}
      </BreadcrumbList>
    </Breadcrumb>
  );
}
