import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from '@sentinez/ui/components/input-group';
import { cn } from '@sentinez/ui/lib/utils';
import { SearchIcon } from 'lucide-react';

/** The search of the header and of the home page: a form going to /search */
export function SearchBox({
  defaultValue,
  size = 'default',
  className,
}: {
  defaultValue?: string;
  size?: 'default' | 'lg';
  className?: string;
}) {
  return (
    <form action="/search" method="get" role="search" className={cn('w-full', className)}>
      <InputGroup className={cn(size === 'lg' && 'h-11 bg-background shadow-sm')}>
        <InputGroupAddon>
          <SearchIcon />
        </InputGroupAddon>
        <InputGroupInput
          name="q"
          type="search"
          defaultValue={defaultValue}
          placeholder="Search packages or rules"
          aria-label="Search packages or rules"
          autoComplete="off"
          className={cn(size === 'lg' && 'text-base')}
        />
        {size === 'lg' && (
          <InputGroupAddon align="inline-end">
            <InputGroupButton type="submit" variant="default" size="sm">
              Search
            </InputGroupButton>
          </InputGroupAddon>
        )}
      </InputGroup>
    </form>
  );
}
