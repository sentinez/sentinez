'use client';

import * as React from 'react';
import { createPortal } from 'react-dom';
import {
  DndContext,
  DragOverlay,
  KeyboardSensor,
  PointerSensor,
  closestCenter,
  pointerWithin,
  useDroppable,
  useSensor,
  useSensors,
  type CollisionDetection,
  type UniqueIdentifier,
} from '@dnd-kit/core';
import {
  SortableContext,
  sortableKeyboardCoordinates,
  useSortable,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { BracesIcon, ChevronDownIcon, GripVerticalIcon, PlusIcon, Trash2Icon } from 'lucide-react';
import { useTranslations } from 'next-intl';
import { Badge } from '@sentinez/ui/components/badge';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuTrigger,
} from '@sentinez/ui/components/dropdown-menu';
import { Button } from '@sentinez/ui/components/button';
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@sentinez/ui/components/card';
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@sentinez/ui/components/collapsible';
import { Input } from '@sentinez/ui/components/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@sentinez/ui/components/select';
import { cn } from '@sentinez/ui/lib/utils';

import {
  AndCondition,
  Condition,
  ConditionSchema,
  Expression,
  FieldSource,
  Operator,
  OperatorSpec,
  ValueKind,
} from '@sentinez/proto/sentinez/types/rule/v1/rule';
import { OPERATOR_OPTIONS } from '@/lib/type/security';
import {
  acceptsKey,
  conditionSpec,
  conformCondition,
  keyIsOptional,
  listToText,
  textToList,
} from '@/lib/type/rule-schema';
import {
  Join,
  MAX_CONDITIONS,
  addCondition,
  countConditions,
  createEmptyExpression,
  flatten,
  moveCondition,
  moveGroup,
  moveToNewGroup,
  removeCondition,
  setJoin,
  updateCondition,
} from '@/lib/type/rule-expr';
import { useSecurityOptions } from '@/hooks/use-security-options';
import { CodeBlock } from '@/components/code-block';

// ─── Factories ───────────────────────────────────────────────────────────────

/**
 * Sources that need a named key input (header name, query param, etc.),
 * while the condition schema of the server has not loaded.
 */
const SOURCES_WITH_KEY = new Set<FieldSource>([
  FieldSource.FIELD_SOURCE_HEADER,
  FieldSource.FIELD_SOURCE_QUERY,
  FieldSource.FIELD_SOURCE_BODY,
  FieldSource.FIELD_SOURCE_JA4,
  FieldSource.FIELD_SOURCE_TLS,
]);

// ─── Context ─────────────────────────────────────────────────────────────────

type QueryBuilderContextValue = {
  /** The expression, flat: undefined when flattening it makes too many conditions */
  flat?: Expression;
  /** What the form receives: the flat expression, or the value as it came */
  expression: Expression;
  schema?: ConditionSchema;
  /** Appends an empty condition to the group g, or as a new group past the last */
  add: (g: number) => void;
  /** Sets the operator before the condition r of the group g */
  setJoin: (g: number, r: number, join: Join) => void;
  remove: (id: string) => void;
  /** Moves the condition to the index of the group toGroup (an id) */
  move: (id: string, toGroup: string, toIndex: number) => void;
  /** Moves the condition out of its group, into a new one last */
  moveToNew: (id: string) => void;
  /** Moves the group (an id) to the index */
  moveBlock: (id: string, toIndex: number) => void;
  updateCondition: (id: string, updater: (condition: Condition) => Condition) => void;
};

const QueryBuilderContext = React.createContext<QueryBuilderContextValue | null>(null);

export function useQueryBuilder() {
  const context = React.useContext(QueryBuilderContext);
  if (!context) {
    throw new Error('useQueryBuilder must be used within <QueryBuilder>');
  }
  return context;
}

// ─── QueryBuilder (root) ─────────────────────────────────────────────────────

export type QueryBuilderProps = Omit<React.ComponentProps<'div'>, 'defaultValue'> & {
  /** Controlled expression. */
  value?: Expression;
  /** Initial expression in uncontrolled mode. */
  defaultValue?: Expression;
  onValueChange?: (value: Expression) => void;
  /**
   * The conditions the rule accepts, see useConditionSchema: the editor only
   * offers what it lists. Without it, every source and operator is offered.
   */
  schema?: ConditionSchema;
};

/**
 * Editor of a rule expression, on the Expression itself: a line of its
 * conditions, AND between two rules of a group, OR between two groups. It takes the width of the
 * form it sits in (a container query lays a condition out on one row only
 * when there is room), so it fits the narrow create / update pages. Without
 * children it shows the editor and a folded JSON preview.
 *
 * The flat expression is the state, so the ids of the conditions survive an
 * edit; it is only flattened again when `value` changes to something the
 * builder did not emit itself.
 */
export function QueryBuilder({
  value: valueProp,
  defaultValue,
  onValueChange,
  schema,
  className,
  children,
  ...props
}: QueryBuilderProps) {
  const initial = valueProp ?? defaultValue ?? createEmptyExpression();
  const [flat, setFlat] = React.useState(() => flatten(initial));
  // the last expression from or to the parent: it is the same object while
  // nothing changed, so only another value is serialised to compare
  const [emitted, setEmitted] = React.useState(initial);
  const onValueChangeRef = React.useRef(onValueChange);
  React.useEffect(() => {
    onValueChangeRef.current = onValueChange;
  });

  // the value changed from outside (a reset, another record): rebuild
  if (
    valueProp !== undefined &&
    valueProp !== emitted &&
    JSON.stringify(valueProp) !== JSON.stringify(emitted)
  ) {
    setFlat(flatten(valueProp));
    setEmitted(valueProp);
  }

  const context = React.useMemo<QueryBuilderContextValue>(() => {
    const commit = (next: Expression) => {
      setFlat(next);
      setEmitted(next);
      onValueChangeRef.current?.(next);
    };
    const edit = (fn: (expr: Expression) => Expression) => flat && commit(fn(flat));

    return {
      flat,
      expression: flat ?? emitted,
      schema,
      add: (g) => edit((e) => addCondition(e, g)),
      setJoin: (g, r, join) => edit((e) => setJoin(e, g, r, join)),
      remove: (id) => edit((e) => removeCondition(e, id)),
      move: (id, toGroup, toIndex) => edit((e) => moveCondition(e, id, toGroup, toIndex)),
      moveToNew: (id) => edit((e) => moveToNewGroup(e, id)),
      moveBlock: (id, toIndex) => edit((e) => moveGroup(e, id, toIndex)),
      updateCondition: (id, updater) => edit((e) => updateCondition(e, id, updater)),
    };
  }, [flat, emitted, schema]);

  return (
    <QueryBuilderContext.Provider value={context}>
      <div
        data-slot="query-builder"
        className={cn('@container/query-builder flex min-w-0 flex-col gap-3', className)}
        {...props}
      >
        {children ?? (
          <>
            <QueryBuilderEditor />
            <QueryBuilderPreview />
          </>
        )}
      </div>
    </QueryBuilderContext.Provider>
  );
}

// ─── QueryBuilderEditor ──────────────────────────────────────────────────────

/**
 * The groups of the expression as blocks, an OR line between two of them;
 * in a block, its conditions, an AND line between two of them. Then the
 * button appending a block. Setting a line to the other operator splits a
 * block, or merges it into the one before.
 */
export function QueryBuilderEditor({ className, ...props }: React.ComponentProps<'div'>) {
  const t = useTranslations('QueryBuilder');
  const { flat } = useQueryBuilder();
  const count = countConditions(flat);
  const canAdd = count < MAX_CONDITIONS;

  return (
    <div
      data-slot="query-builder-editor"
      className={cn('flex min-w-0 flex-col gap-3', className)}
      {...props}
    >
      {!flat ? (
        <p className="text-sm text-balance text-muted-foreground">
          {t('tooLarge', { n: MAX_CONDITIONS })}
        </p>
      ) : (
        <>
          {count === 0 ? (
            <p className="py-2 text-center text-sm text-balance text-muted-foreground">
              {t('noConditions')}
            </p>
          ) : (
            <QueryBuilderBlocks flat={flat} canAdd={canAdd} />
          )}
          <QueryBuilderActions count={count} groups={flat.orCondition.length} canAdd={canAdd} />
        </>
      )}
    </div>
  );
}

/**
 * Where a condition is: its group, its index there, and its position in the
 * line; for a group id, the end of that group.
 */
function locate(
  expr: Expression,
  id: UniqueIdentifier,
): { group: string; index: number; line: number; condition?: Condition } | undefined {
  let line = 0;
  for (const g of expr.orCondition) {
    if (g.id === id) return { group: g.id, index: g.rules.length, line: line + g.rules.length };
    const r = g.rules.findIndex((rule) => rule.condition?.id === id);
    if (r >= 0) return { group: g.id, index: r, line: line + r, condition: g.rules[r]?.condition };
    line += g.rules.length;
  }
  return undefined;
}

/** What is dragged or dropped on: the data of its draggable / droppable */
type DragKind = 'condition' | 'block' | 'new-block';
const kindOf = (data: { current?: { kind?: DragKind } } | undefined) => data?.current?.kind;

/** The drop zone making a new block of a condition */
const NEW_BLOCK = 'query-builder-new-block';

/**
 * A sortable condition or block: the ref and style of its element, the
 * props of its handle, and whether it is the one dragged.
 */
function useSortableItem(id: string, kind: DragKind) {
  const {
    attributes,
    listeners,
    setNodeRef,
    setActivatorNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id, data: { kind } });

  return {
    ref: setNodeRef,
    style: { transform: CSS.Translate.toString(transform), transition },
    handle: { ref: setActivatorNodeRef, ...attributes, ...listeners },
    isDragging,
  };
}

/**
 * The blocks, an OR line between two of them. Each is dragged by the handle
 * of its header (or moved with the keyboard) to another place among them.
 * A condition is dragged by its handle to another place of its block, into
 * another block, or out of them all into a new one: dropped on a condition,
 * it takes its place; on a block, it goes to its end; on the drop zone shown
 * under the blocks while it is dragged, it makes a new block. A copy of what
 * is dragged follows the pointer, and where it would go is outlined.
 */
function QueryBuilderBlocks({ flat, canAdd }: { flat: Expression; canAdd: boolean }) {
  const { setJoin, move, moveToNew, moveBlock } = useQueryBuilder();
  // a stable id, so its accessibility ids match between server and client
  const dndId = React.useId();
  const [dragged, setDragged] = React.useState<{ id: UniqueIdentifier; kind?: DragKind }>();
  const [overGroup, setOverGroup] = React.useState<string>();
  // the width of the query builder and of what is dragged, when it started:
  // the copy that follows the pointer is out of the builder, so its
  // container queries are answered by a container as wide as the builder
  const [widths, setWidths] = React.useState<{ builder: number; item: number }>();
  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 4 } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }),
  );

  const reset = () => {
    setDragged(undefined);
    setOverGroup(undefined);
    setWidths(undefined);
  };
  const blockIds = flat.orCondition.map((g) => g.id);
  // the position in the line of the first condition of each block
  const starts: number[] = [];
  let start = 0;
  for (const g of flat.orCondition) {
    starts.push(start);
    start += g.rules.length;
  }
  const draggedCondition = dragged?.kind === 'condition' ? locate(flat, dragged.id) : undefined;
  const draggedBlock =
    dragged?.kind === 'block' ? flat.orCondition.find((g) => g.id === dragged.id) : undefined;

  return (
    <DndContext
      id={dndId}
      sensors={sensors}
      collisionDetection={collision}
      onDragStart={({ active, activatorEvent }) => {
        setDragged({ id: active.id, kind: kindOf(active.data) });
        const target = activatorEvent.target instanceof Element ? activatorEvent.target : null;
        const builder = target?.closest('[data-slot=query-builder]');
        const item = active.rect.current.initial?.width;
        if (builder && item) setWidths({ builder: builder.clientWidth, item });
      }}
      onDragOver={({ active, over }) =>
        setOverGroup(
          kindOf(active.data) === 'condition' && over ? locate(flat, over.id)?.group : undefined,
        )
      }
      onDragCancel={reset}
      onDragEnd={({ active, over }) => {
        reset();
        if (!over) return;
        const id = String(active.id);
        if (kindOf(active.data) === 'block') {
          const to = blockIds.indexOf(String(over.id));
          if (to >= 0) moveBlock(id, to);
          return;
        }
        if (over.id === NEW_BLOCK) return moveToNew(id);
        const to = locate(flat, over.id);
        if (to) move(id, to.group, to.index);
      }}
    >
      <SortableContext
        id="query-builder-blocks"
        items={blockIds}
        strategy={verticalListSortingStrategy}
      >
        <div className="grid min-w-0 gap-2">
          {flat.orCondition.map((group, g) => (
            <React.Fragment key={group.id}>
              {g > 0 && <JoinDivider value="or" onValueChange={(join) => setJoin(g, 0, join)} />}
              <QueryBuilderBlock
                group={group}
                g={g}
                start={starts[g] ?? 0}
                canAdd={canAdd}
                isTarget={overGroup === group.id && draggedCondition?.group !== group.id}
              />
            </React.Fragment>
          ))}
          {draggedCondition && <NewBlockZone />}
        </div>
      </SortableContext>

      <OverlayPortal>
        <DragOverlay>
          <div
            className="pointer-events-none @container/query-builder"
            style={widths && { width: widths.builder }}
          >
            <div style={widths && { width: widths.item }}>
              {draggedCondition?.condition && (
                <QueryBuilderRule
                  condition={draggedCondition.condition}
                  index={draggedCondition.line}
                  handle={<DragHandle label="" />}
                  className="rounded-md border bg-card p-2 shadow-lg"
                />
              )}
              {draggedBlock && (
                <Card className="gap-0 rounded-md py-0 shadow-lg">
                  <BlockHeader handle={<DragHandle label="" />} count={draggedBlock.rules.length} />
                </Card>
              )}
            </div>
          </div>
        </DragOverlay>
      </OverlayPortal>
    </DndContext>
  );
}

/**
 * Renders the drag overlay in the body. The overlay is fixed to the viewport,
 * but in a dialog, whose content is translated to the center, a fixed element
 * is placed against the dialog instead, so the copy would leave the pointer.
 */
function OverlayPortal({ children }: { children: React.ReactNode }) {
  const [mounted, setMounted] = React.useState(false);
  React.useEffect(() => setMounted(true), []);
  return mounted ? createPortal(children, document.body) : null;
}

/**
 * Where a drop lands. A block goes among the blocks: the closest one. A
 * condition goes on the condition under the pointer; between the conditions
 * of a block, on the closest of them; on the new block zone or a block, on
 * it; out of them all (or with the keyboard), on the closest condition.
 */
const collision: CollisionDetection = (args) => {
  const containers = args.droppableContainers;
  if (kindOf(args.active.data) === 'block') {
    const blocks = containers.filter((c) => kindOf(c.data) === 'block');
    return closestCenter({ ...args, droppableContainers: blocks });
  }

  const within = pointerWithin(args);
  const conditions = within.filter((c) => {
    const container = containers.find((d) => d.id === c.id);
    return kindOf(container?.data) === 'condition';
  });
  if (conditions.length > 0) return conditions;
  const target = within[0];
  if (!target) {
    const all = containers.filter((c) => kindOf(c.data) === 'condition');
    return closestCenter({ ...args, droppableContainers: all });
  }
  if (target.id === NEW_BLOCK) return [target];
  const inBlock = containers.filter((c) => c.data.current?.sortable?.containerId === target.id);
  return inBlock.length > 0 ? closestCenter({ ...args, droppableContainers: inBlock }) : [target];
};

/** The zone under the blocks a condition is dropped on to make a new block */
function NewBlockZone() {
  const t = useTranslations('QueryBuilder');
  const { setNodeRef, isOver } = useDroppable({ id: NEW_BLOCK, data: { kind: 'new-block' } });

  return (
    <div
      ref={setNodeRef}
      className={cn(
        'flex min-h-14 items-center justify-center gap-2 rounded-md border border-dashed px-3 text-center text-sm text-muted-foreground',
        isOver && 'border-primary bg-primary/5 text-primary',
      )}
    >
      <PlusIcon className="size-4 shrink-0" aria-hidden />
      <span className="text-balance">{t('dropNewBlock')}</span>
    </div>
  );
}

/**
 * The header of a block, and of the copy of it following the pointer: its
 * drag handle and the number of its conditions, 4px above and below the
 * handle (40px, 44px when wide).
 */
function BlockHeader({ handle, count }: { handle: React.ReactNode; count: number }) {
  const t = useTranslations('QueryBuilder');

  return (
    <CardHeader className="flex items-center gap-1 px-2.5 py-1">
      {handle}
      <CardTitle className="text-xs font-medium text-muted-foreground">
        {t('conditionCount', { n: count })}
      </CardTitle>
    </CardHeader>
  );
}

type QueryBuilderBlockProps = {
  group: AndCondition;
  /** Position of the group in the expression */
  g: number;
  /** Position of its first condition in the line, for the labels */
  start: number;
  canAdd: boolean;
  /** Whether a condition dragged from another block would be dropped in it */
  isTarget?: boolean;
};

/**
 * A group as a block, a Card: a header with the handle moving it and the
 * number of its conditions, its conditions joined by AND lines, and a footer
 * with the button appending one. The spacing of the card is tightened to fit
 * the form, and the three line up on the same left edge. A condition dropped
 * on it, not on one of its conditions, goes last.
 */
function QueryBuilderBlock({ group, g, start, canAdd, isTarget }: QueryBuilderBlockProps) {
  const t = useTranslations('QueryBuilder');
  const { add, setJoin } = useQueryBuilder();
  const sortable = useSortableItem(group.id, 'block');
  const ids = group.rules.flatMap((r) => (r.condition ? [r.condition.id] : []));

  return (
    <Card
      ref={sortable.ref}
      style={sortable.style}
      role="group"
      aria-label={t('blockN', { n: g + 1 })}
      data-slot="query-builder-block"
      className={cn(
        'min-w-0 gap-0 overflow-hidden rounded-md py-0 shadow-none transition-shadow',
        isTarget && 'ring-2 ring-primary/40',
        sortable.isDragging && 'opacity-40',
      )}
    >
      <BlockHeader
        handle={<DragHandle label={t('moveBlock', { n: g + 1 })} {...sortable.handle} />}
        count={group.rules.length}
      />

      <CardContent className="grid min-w-0 gap-2 px-2.5 py-2">
        <SortableContext id={group.id} items={ids} strategy={verticalListSortingStrategy}>
          {group.rules.map(({ condition }, r) =>
            condition ? (
              <React.Fragment key={condition.id}>
                {r > 0 && <JoinDivider value="and" onValueChange={(join) => setJoin(g, r, join)} />}
                <SortableRule condition={condition} index={start + r} />
              </React.Fragment>
            ) : null,
          )}
        </SortableContext>
      </CardContent>

      {/* as high as the header (40px, 44px when wide), the button centred in it */}
      <CardFooter className="h-10 px-0.5 py-0 @md/query-builder:h-11">
        <Button
          type="button"
          variant="ghost"
          size="sm"
          // 2px of the footer and 8px of its own put the icon on the left
          // edge of the fields (10px), its hover inside the card
          className="h-7 max-w-full text-primary hover:text-primary has-[>svg]:px-2"
          disabled={!canAdd}
          onClick={() => add(g)}
        >
          <PlusIcon aria-hidden />
          <span className="truncate">{t('and')}</span>
        </Button>
      </CardFooter>
    </Card>
  );
}

/**
 * The operator joining two conditions (AND, in a block) or two blocks (OR),
 * on the left: a primary Badge opening a menu to set it, on a vertical line
 * from the one above to the one below. The line runs under the drag handles;
 * an OR one is inset by the padding of a block, to its header handle.
 */
function JoinDivider({
  value,
  onValueChange,
}: {
  value: Join;
  onValueChange: (join: Join) => void;
}) {
  const t = useTranslations('QueryBuilder');
  const label = { and: t('and'), or: t('or') };
  // in a column as wide as a drag handle, so the line is under its icon
  const line = (
    <div aria-hidden className="flex w-8 justify-center @md/query-builder:w-9">
      <div className="h-3 w-0.5 origin-center scale-x-50 bg-border" />
    </div>
  );

  return (
    // over the gaps of the list (-my-2), so the line touches both sides
    <div className={cn('-my-2 grid justify-start', value === 'or' && 'pl-2.5')}>
      {line}
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Badge asChild className="cursor-pointer font-semibold hover:bg-primary/90">
            <button type="button" aria-label={`${t('operator')}: ${label[value]}`}>
              {/* as wide as the longer label, so switching does not resize it */}
              <span className="grid justify-items-center *:col-start-1 *:row-start-1">
                <span>{label[value]}</span>
                <span aria-hidden className="invisible">
                  {label.and}
                </span>
                <span aria-hidden className="invisible">
                  {label.or}
                </span>
              </span>
              <ChevronDownIcon aria-hidden />
            </button>
          </Badge>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="start" className="min-w-28">
          <DropdownMenuRadioGroup
            value={value}
            onValueChange={(join) => onValueChange(join as Join)}
          >
            <DropdownMenuRadioItem value="and">{label.and}</DropdownMenuRadioItem>
            <DropdownMenuRadioItem value="or">{label.or}</DropdownMenuRadioItem>
          </DropdownMenuRadioGroup>
        </DropdownMenuContent>
      </DropdownMenu>
      {line}
    </div>
  );
}

/** The button appending a block: OR, or the first condition while there is none */
function QueryBuilderActions({
  count,
  groups,
  canAdd,
}: {
  count: number;
  groups: number;
  canAdd: boolean;
}) {
  const t = useTranslations('QueryBuilder');
  const { add } = useQueryBuilder();

  return (
    <div className="grid gap-1">
      <Button
        type="button"
        variant="outline"
        size="sm"
        className="max-w-full justify-self-start"
        disabled={!canAdd}
        onClick={() => add(groups)}
      >
        <PlusIcon aria-hidden />
        <span className="truncate">{count === 0 ? t('addCondition') : t('addBlock')}</span>
      </Button>
      {!canAdd && (
        <p className="text-xs text-muted-foreground">{t('maxConditions', { n: MAX_CONDITIONS })}</p>
      )}
    </div>
  );
}

function RemoveButton({ label, onClick }: { label: string; onClick: () => void }) {
  return (
    <Button
      type="button"
      variant="ghost"
      size="icon"
      className="size-8 shrink-0 text-muted-foreground hover:text-destructive @md/query-builder:size-9"
      aria-label={label}
      onClick={onClick}
    >
      <Trash2Icon className="size-4" aria-hidden />
    </Button>
  );
}

type DragHandleProps = Omit<React.ComponentProps<typeof Button>, 'children'> & {
  label: string;
};

/** The handle a condition is dragged by, or moved with the keyboard */
function DragHandle({ label, className, ...props }: DragHandleProps) {
  return (
    <Button
      type="button"
      variant="ghost"
      size="icon"
      className={cn(
        'size-8 shrink-0 cursor-grab touch-none text-muted-foreground active:cursor-grabbing',
        '@md/query-builder:size-9',
        className,
      )}
      aria-label={label}
      {...props}
    >
      <GripVerticalIcon className="size-4" aria-hidden />
    </Button>
  );
}

/**
 * A condition of a block that can be moved: by its handle, with the pointer
 * or the keyboard. While dragged, it stays faded in its place.
 */
function SortableRule({ condition, index }: { condition: Condition; index: number }) {
  const t = useTranslations('QueryBuilder');
  const sortable = useSortableItem(condition.id, 'condition');

  return (
    <QueryBuilderRule
      ref={sortable.ref}
      style={sortable.style}
      className={cn(sortable.isDragging && 'opacity-40')}
      condition={condition}
      index={index}
      handle={<DragHandle label={t('moveCondition', { n: index + 1 })} {...sortable.handle} />}
    />
  );
}

// ─── QueryBuilderRule ────────────────────────────────────────────────────────

export type QueryBuilderRuleProps = Omit<React.ComponentProps<'div'>, 'children'> & {
  condition: Condition;
  /** Position in the line, for the labels */
  index: number;
  /** The handle moving it, left of its fields */
  handle?: React.ReactNode;
};

/**
 * One condition: source, key (header, query... name), operator and value,
 * and the button removing it, on one line. Two fields per row in a narrow
 * form, all on one row when the builder is wide.
 * The condition schema of the builder, when it has one, decides which
 * sources, keys and operators it offers and how the value is entered; an
 * edit keeps the condition inside it.
 */
export function QueryBuilderRule({
  condition,
  index,
  handle,
  className,
  ...props
}: QueryBuilderRuleProps) {
  const t = useTranslations('QueryBuilder');
  const { fieldSourceOptions } = useSecurityOptions();
  const { schema, updateCondition, remove } = useQueryBuilder();

  const spec = conditionSpec(schema, condition);
  const needsKey = schema ? acceptsKey(spec.source) : SOURCES_WITH_KEY.has(condition.source);
  const sourceOptions = schema
    ? fieldSourceOptions.filter((o) => schema.sources.some((s) => s.source === o.value))
    : fieldSourceOptions;
  const operatorOptions = spec.variant
    ? OPERATOR_OPTIONS.filter((o) => spec.variant?.operators.some((s) => s.operator === o.value))
    : OPERATOR_OPTIONS;
  const update = (patch: Partial<Condition>) =>
    updateCondition(condition.id, (c) => conformCondition(schema, { ...c, ...patch }));

  return (
    <div
      role="group"
      aria-label={t('conditionN', { n: index + 1 })}
      data-slot="query-builder-rule"
      data-has-key={needsKey || undefined}
      className={cn('flex min-w-0 items-start gap-1', className)}
      {...props}
    >
      {handle}
      {/* the operator column is narrow, to leave the width to the inputs */}
      <div
        className={cn(
          'grid min-w-0 flex-1 grid-cols-1 gap-2',
          needsKey
            ? '@xs/query-builder:grid-cols-[minmax(0,1fr)_7.5rem] @2xl/query-builder:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_7.5rem_minmax(0,1.5fr)]'
            : '@xs/query-builder:grid-cols-[7.5rem_minmax(0,1fr)] @2xl/query-builder:grid-cols-[minmax(0,1fr)_7.5rem_minmax(0,2fr)]',
        )}
      >
        <EnumSelect
          aria-label={t('source')}
          value={condition.source}
          onValueChange={(source) => update({ source: source as FieldSource, key: '' })}
          options={sourceOptions}
          // a row of its own until the four fit on one line
          className="@xs/query-builder:col-span-2 @2xl/query-builder:col-span-1"
        />

        {needsKey && (
          <Input
            aria-label={t('key')}
            placeholder={
              keyIsOptional(spec.source) ? t('keyOptionalPlaceholder') : t('keyPlaceholder')
            }
            spellCheck={false}
            autoComplete="off"
            value={condition.key}
            onChange={(e) => update({ key: e.target.value })}
            className="min-w-0 bg-background font-mono text-xs text-ellipsis"
          />
        )}

        <EnumSelect
          aria-label={t('operator')}
          value={condition.operator}
          onValueChange={(operator) => update({ operator: operator as Operator })}
          options={operatorOptions}
        />

        <ConditionValue
          spec={spec.operator}
          value={condition.value}
          onValueChange={(value) => update({ value })}
          // with a key, the value has a row of its own on two columns
          className={
            needsKey ? '@xs/query-builder:col-span-2 @2xl/query-builder:col-span-1' : undefined
          }
        />
      </div>

      <RemoveButton
        label={t('removeCondition', { n: index + 1 })}
        onClick={() => remove(condition.id)}
      />
    </div>
  );
}

// ─── ConditionValue ──────────────────────────────────────────────────────────

type ConditionValueProps = {
  /** The operator of the condition in the schema, undefined without a schema */
  spec?: OperatorSpec;
  value: unknown;
  onValueChange: (value: string | string[]) => void;
  className?: string;
};

/**
 * The value input the operator takes: a select of the allowed values, a
 * comma separated list, or a text with a hint of the expected kind.
 */
function ConditionValue({ spec, value, onValueChange, className }: ConditionValueProps) {
  const t = useTranslations('QueryBuilder');
  const inputClass = cn('min-w-0 bg-background text-ellipsis', className);

  if (spec?.valueKind === ValueKind.VALUE_KIND_ENUM && !spec.list) {
    return (
      <StringSelect
        aria-label={t('value')}
        value={typeof value === 'string' ? value : ''}
        onValueChange={onValueChange}
        options={spec.enumValues}
        className={className}
      />
    );
  }

  if (spec?.list) {
    const placeholder =
      spec.valueKind === ValueKind.VALUE_KIND_ENUM
        ? t('valuePlaceholder.enumList', { values: spec.enumValues.join(', ') })
        : t('valuePlaceholder.list');
    return (
      <ListInput
        aria-label={t('value')}
        placeholder={placeholder}
        value={value}
        onValueChange={onValueChange}
        className={inputClass}
      />
    );
  }

  return (
    <Input
      aria-label={t('value')}
      placeholder={t(`valuePlaceholder.${valueHint(spec?.valueKind)}`)}
      spellCheck={false}
      autoComplete="off"
      inputMode={spec?.valueKind === ValueKind.VALUE_KIND_NUMBER ? 'decimal' : undefined}
      value={listToText(value)}
      onChange={(e) => onValueChange(e.target.value)}
      className={inputClass}
    />
  );
}

/** The placeholder key of a single value of the kind */
function valueHint(kind?: ValueKind): 'string' | 'number' | 'regex' | 'ip' | 'name' {
  switch (kind) {
    case ValueKind.VALUE_KIND_NUMBER:
      return 'number';
    case ValueKind.VALUE_KIND_REGEX:
      return 'regex';
    case ValueKind.VALUE_KIND_IP:
      return 'ip';
    case ValueKind.VALUE_KIND_NAME:
      return 'name';
    default:
      return 'string';
  }
}

type ListInputProps = Omit<React.ComponentProps<typeof Input>, 'value' | 'onChange'> & {
  value: unknown;
  onValueChange: (value: string[]) => void;
};

/**
 * A list value typed as comma separated text. The text is kept as typed
 * (a trailing comma, spaces) while the list it makes is what changes.
 */
function ListInput({ value, onValueChange, ...props }: ListInputProps) {
  const text = listToText(value);
  const [draft, setDraft] = React.useState(text);
  const [synced, setSynced] = React.useState(text);

  // the value changed from outside (another operator, a reset): show it
  if (text !== synced) {
    setSynced(text);
    if (textToList(draft).join(', ') !== text) setDraft(text);
  }

  return (
    <Input
      spellCheck={false}
      autoComplete="off"
      value={draft}
      onChange={(e) => {
        setDraft(e.target.value);
        onValueChange(textToList(e.target.value));
      }}
      {...props}
    />
  );
}

// ─── QueryBuilderPreview ─────────────────────────────────────────────────────

export type QueryBuilderPreviewProps = React.ComponentProps<typeof Collapsible>;

/** The expression as JSON, folded by default. */
export function QueryBuilderPreview({ className, ...props }: QueryBuilderPreviewProps) {
  const t = useTranslations('QueryBuilder');
  const { expression } = useQueryBuilder();

  return (
    <Collapsible
      data-slot="query-builder-preview"
      className={cn('group/preview flex min-w-0 flex-col gap-2', className)}
      {...props}
    >
      <CollapsibleTrigger asChild>
        <Button type="button" variant="ghost" size="sm" className="max-w-full self-start">
          <BracesIcon className="size-4" aria-hidden />
          <span className="truncate">{t('json')}</span>
          <ChevronDownIcon
            className="size-4 transition-transform group-data-[state=open]/preview:rotate-180"
            aria-hidden
          />
        </Button>
      </CollapsibleTrigger>
      <CollapsibleContent>
        <CodeBlock value={expression} />
      </CollapsibleContent>
    </Collapsible>
  );
}

// ─── StringSelect ────────────────────────────────────────────────────────────

type StringSelectProps = Omit<EnumSelectProps, 'value' | 'onValueChange' | 'options'> & {
  value: string;
  onValueChange: (value: string) => void;
  options: string[];
};

/** A select of literal values, shown as they are sent (GET, true...) */
function StringSelect({ value, onValueChange, options, ...props }: StringSelectProps) {
  return (
    <EnumSelect
      {...props}
      value={options.indexOf(value)}
      onValueChange={(i) => onValueChange(options[i] ?? '')}
      options={options.map((label, value) => ({ label, value }))}
    />
  );
}

// ─── EnumSelect ──────────────────────────────────────────────────────────────

type EnumSelectProps = {
  value: number;
  onValueChange: (value: number) => void;
  options: { label: string; value: number }[];
  className?: string;
  'aria-label'?: string;
};

function EnumSelect({
  value,
  onValueChange,
  options,
  className,
  'aria-label': ariaLabel,
}: EnumSelectProps) {
  const tc = useTranslations('Common');

  return (
    <Select value={String(value)} onValueChange={(v) => onValueChange(Number(v))}>
      <SelectTrigger
        aria-label={ariaLabel}
        className={cn(
          'w-full min-w-0 bg-background',
          // the value truncates instead of widening the trigger
          '*:data-[slot=select-value]:block *:data-[slot=select-value]:min-w-0',
          '*:data-[slot=select-value]:truncate',
          className,
        )}
      >
        <SelectValue placeholder={tc('select')} />
      </SelectTrigger>
      <SelectContent>
        {options.map((opt) => (
          <SelectItem key={opt.value} value={String(opt.value)}>
            {opt.label}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
