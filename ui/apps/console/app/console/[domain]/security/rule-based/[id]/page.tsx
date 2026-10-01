'use client';

import { useRouter } from 'next/navigation';
import { Button } from '@sentinez/ui/components/button';
import { Input } from '@sentinez/ui/components/input';
import { Label } from '@sentinez/ui/components/label';
import { toast } from '@/lib/toast';
import IsLoading from '@sentinez/ui/components/common/loading';
import { QueryBuilder, createEmptyExpression } from '../../components';
import { Expression } from '@sentinez/proto/sentinez/secure/rule/v1/engine';
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
import { getRuleBased } from '@/lib/api/security';
import { PageLayout, PageLayoutContent, PageLayoutHeader } from '@/components/page-layout';
import { useEffect, useState } from 'react';

export default function EditRuleBasedPage({ params }: { params: Promise<{ id: string }> }) {
  const router = useRouter();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [alignItemWithTrigger, setAlignItemWithTrigger] = useState(true);

  // form state
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [priority, setPriority] = useState('1');
  const [query, setQuery] = useState<Expression>(createEmptyExpression);
  const [actionJson, setActionJson] = useState('BLOCK');

  const [id, setId] = useState<string | null>(null);

  useEffect(() => {
    async function load() {
      try {
        const { id } = await params;
        setId(id);
        const rule = await getRuleBased(id);
        setName(rule.ingressRuntime?.name || '');
        setDescription(rule.ingressRuntime?.description || '');
        setPriority(String(rule.ingressRuntime?.priority || 1));
        setQuery(rule.ingressRuntime?.expr ?? createEmptyExpression());

        setActionJson(JSON.stringify(rule.ingressRuntime?.action || {}, null, 2));
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
    <PageLayout>
      <PageLayoutHeader title="Edit Rule" subtitle="Modify security rule configuration and logic.">
        <div className="flex justify-start gap-2">
          <Button disabled={saving} onClick={handleSave}>
            {saving ? 'Saving...' : 'Save'}
          </Button>
          <Button variant="secondary" onClick={() => router.back()}>
            Cancel
          </Button>
        </div>
      </PageLayoutHeader>
      <PageLayoutContent>
        <div className="grid gap-6 py-4">
          <Card className="grid gap-2 shadow-none border-none">
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

          <Card className="grid gap-2 shadow-none border-none">
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
              <QueryBuilder orientation="horizontal" value={query} onValueChange={setQuery} />
            </CardContent>
            <CardFooter></CardFooter>
          </Card>
        </div>
      </PageLayoutContent>
    </PageLayout>
  );
}
SelectLabel;
