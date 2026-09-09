'use client';

import * as React from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@sentinez/ui/components/button';
import { Input } from '@sentinez/ui/components/input';
import { Label } from '@sentinez/ui/components/label';
import { toast } from '@/lib/toast';
import { getRuleBased, updateRuleBased, RuleBased } from '@/lib/api/security';
import IsLoading from '@sentinez/ui/components/common/loading';
import { QueryBuilder, RuleGroup } from '../../components';
import Title from '@/components/title';
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@sentinez/ui/components/card';
import { Textarea } from '@sentinez/ui/components/textarea';
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from '@sentinez/ui/components/select';
import {
  Field,
  FieldContent,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from '@sentinez/ui/components/field';
import { Switch } from '@sentinez/ui/components/switch';

export default function EditRuleBasedPage({ params }: { params: Promise<{ id: string }> }) {
  const router = useRouter();
  const [loading, setLoading] = React.useState(true);
  const [saving, setSaving] = React.useState(false);
  const [alignItemWithTrigger, setAlignItemWithTrigger] = React.useState(true);

  // form state
  const [name, setName] = React.useState('');
  const [description, setDescription] = React.useState('');
  const [priority, setPriority] = React.useState('1');
  const [query, setQuery] = React.useState<RuleGroup | undefined>(undefined);
  const [actionJson, setActionJson] = React.useState('BLOCK');

  const [id, setId] = React.useState<string | null>(null);

  React.useEffect(() => {
    async function load() {
      try {
        const { id } = await params;
        setId(id);
        const rule = await getRuleBased(id);
        setName(rule.name || '');
        setDescription(rule.description || '');
        setPriority(String(rule.priority || 1));
        if (rule.node) {
          setQuery(rule.node as any);
        }
        setActionJson(JSON.stringify(rule.action || {}, null, 2));
      } catch (err: any) {
        toast.error('Failed to load rule details');
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [params]);

  const handleSave = async () => {
    if (!id) return;
    setSaving(true);
    try {
      let parsedAction;
      try {
        parsedAction = JSON.parse(actionJson);
      } catch (err) {
        toast.error('Invalid JSON syntax for action');
        setSaving(false);
        return;
      }

      const payload: RuleBased = {
        name,
        description,
        priority: parseInt(priority, 10),
        status: 'STATUS_ACTIVE',
        node: query as any,
        action: parsedAction,
      };

      await updateRuleBased(id, payload);
      toast.success('Rule updated successfully');
      router.back();
    } catch (err: any) {
      toast.error('Failed to save rule');
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return <IsLoading />;
  }

  return (
    <div className="flex flex-col gap-6 w-full">
      <Title title="Edit Rule" subtitle="Modify security rule configuration and logic.">
        <div className="flex justify-start gap-2">
          <Button disabled={saving} onClick={handleSave}>
            {saving ? 'Saving...' : 'Save'}
          </Button>
          <Button variant="secondary" onClick={() => router.back()}>
            Cancel
          </Button>
        </div>
      </Title>
      <div className="grid gap-6 py-4">
        <Card className="grid gap-2 shadow-none">
          <CardContent className="max-w-md grid gap-6">
            <div className="grid gap-2">
              <Label htmlFor="name">Name</Label>
              <Input id="name" value={name} onChange={(e) => setName(e.target.value)} />
            </div>

            <div className="grid gap-2">
              <Label htmlFor="description">Description</Label>
              <Textarea
                id="description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
              />
            </div>

            <div className="grid gap-2">
              <Label htmlFor="priority">Priority</Label>
              <Input
                id="priority"
                type="number"
                value={priority}
                onChange={(e) => setPriority(e.target.value)}
              />
            </div>
          </CardContent>
        </Card>

        <Card className="grid gap-2 shadow-none">
          <CardHeader>
            <CardTitle>Condition Logic (Rule Builder)</CardTitle>
            <CardDescription className="text-muted-foreground">
              Visually assemble natural expressions for routing and security rules.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <FieldGroup className="w-full max-w-md py-6">
              <Field orientation="horizontal">
                <FieldContent>
                  <FieldLabel htmlFor="align-item">Align Item</FieldLabel>
                  <FieldDescription>Toggle to align the item with the trigger.</FieldDescription>
                </FieldContent>
                <Switch
                  id="align-item"
                  checked={alignItemWithTrigger}
                  onCheckedChange={setAlignItemWithTrigger}
                />
              </Field>
              <Field>
                <Select defaultValue="banana">
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent position={alignItemWithTrigger ? 'item-aligned' : 'popper'}>
                    <SelectGroup>
                      <SelectItem value="apple">Apple</SelectItem>
                      <SelectItem value="banana">Banana</SelectItem>
                      <SelectItem value="blueberry">Blueberry</SelectItem>
                      <SelectItem value="grapes">Grapes</SelectItem>
                      <SelectItem value="pineapple">Pineapple</SelectItem>
                    </SelectGroup>
                  </SelectContent>
                </Select>
              </Field>
            </FieldGroup>
            <QueryBuilder layout="horizontal" initialQuery={query} onChange={setQuery} />
          </CardContent>
          <CardFooter></CardFooter>
        </Card>
      </div>
    </div>
  );
}
SelectLabel;
