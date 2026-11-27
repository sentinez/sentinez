import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@sentinez/ui/components/table';

type LinkRow = { href: string; name: string; detail?: string; extra?: string };

/** A table of links (files, packages, versions): a mono name, a detail, a right column */
export function LinkTable({ rows, head }: { rows: LinkRow[]; head: [string, string, string?] }) {
  return (
    <div className="overflow-hidden rounded-lg border">
      <Table>
        <TableHeader className="bg-muted/50">
          <TableRow>
            <TableHead className="w-1/3">{head[0]}</TableHead>
            <TableHead>{head[1]}</TableHead>
            {head[2] && <TableHead className="text-right">{head[2]}</TableHead>}
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((r) => (
            <TableRow key={r.href}>
              <TableCell>
                <a href={r.href} className="font-mono text-primary hover:underline">
                  {r.name}
                </a>
              </TableCell>
              <TableCell className="whitespace-normal text-muted-foreground">{r.detail}</TableCell>
              {head[2] && <TableCell className="text-right tabular-nums text-muted-foreground">{r.extra}</TableCell>}
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
