'use client';
import * as React from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@sentinez/ui/components/button';
import { Input } from '@sentinez/ui/components/input';
import { Label } from '@sentinez/ui/components/label';
import { toast } from '@/lib/toast';
import { getRuleBased, updateRuleBased } from '@/lib/api/security';
import { ChevronLeft } from 'lucide-react';
import IsLoading from '@sentinez/ui/components/common/loading';
import { QueryBuilder } from '../components';
export default function EditRuleBasedPage({ params }) {
    const router = useRouter();
    const [loading, setLoading] = React.useState(true);
    const [saving, setSaving] = React.useState(false);
    // form state
    const [name, setName] = React.useState('');
    const [description, setDescription] = React.useState('');
    const [priority, setPriority] = React.useState('1');
    const [query, setQuery] = React.useState(undefined);
    const [actionJson, setActionJson] = React.useState('{}');
    const [id, setId] = React.useState(null);
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
                    setQuery(rule.node);
                }
                setActionJson(JSON.stringify(rule.action || {}, null, 2));
            }
            catch (err) {
                toast.error('Failed to load rule details');
            }
            finally {
                setLoading(false);
            }
        }
        load();
    }, [params]);
    const handleSave = async () => {
        if (!id)
            return;
        setSaving(true);
        try {
            let parsedAction;
            try {
                parsedAction = JSON.parse(actionJson);
            }
            catch (err) {
                toast.error('Invalid JSON syntax for action');
                setSaving(false);
                return;
            }
            const payload = {
                name,
                description,
                priority: parseInt(priority, 10),
                status: 'STATUS_ACTIVE',
                node: query,
                action: parsedAction,
            };
            await updateRuleBased(id, payload);
            toast.success('Rule updated successfully');
            router.back();
        }
        catch (err) {
            toast.error('Failed to save rule');
        }
        finally {
            setSaving(false);
        }
    };
    if (loading) {
        return <IsLoading />;
    }
    return (<div className="flex flex-col gap-6 max-w-7xl w-full">
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="icon" onClick={() => router.back()}>
          <ChevronLeft className="w-5 h-5"/>
        </Button>
        <div>
          <h3 className="text-xl font-bold tracking-tight">Edit Rule</h3>
          <p className="text-muted-foreground">Modify WAF rule configuration and node logic.</p>
        </div>
      </div>

      <div className="grid gap-6 py-4">
        <div className="grid grid-cols-1 lg:grid-cols-5 gap-6">
          <div className="lg:col-span-3 grid gap-6">
            <div className="grid gap-2">
              <Label htmlFor="name">Name</Label>
              <Input id="name" value={name} onChange={(e) => setName(e.target.value)}/>
            </div>

            <div className="grid gap-2">
              <Label htmlFor="description">Description</Label>
              <Input id="description" value={description} onChange={(e) => setDescription(e.target.value)}/>
            </div>

            <div className="grid gap-2">
              <Label htmlFor="priority">Priority</Label>
              <Input id="priority" type="number" value={priority} onChange={(e) => setPriority(e.target.value)}/>
            </div>
          </div>
        </div>

        <div className="grid gap-2">
          <Label>Condition Logic (Rule Builder)</Label>
          <p className="text-xs text-muted-foreground">
            Visually assemble natural expressions for routing and security rules.
          </p>
          <div className="pt-2">
            <QueryBuilder layout="horizontal" initialQuery={query} onChange={setQuery}/>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-5 gap-6">
          <div className="lg:col-span-3 grid gap-2">
            <Label htmlFor="actionLogic">Action Logic (JSON)</Label>
            <textarea id="actionLogic" className="flex min-h-[100px] w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50 font-mono" value={actionJson} onChange={(e) => setActionJson(e.target.value)}/>
          </div>
        </div>

        <div className="flex justify-start gap-2 pt-4">
          <Button disabled={saving} onClick={handleSave}>
            {saving ? 'Saving...' : 'Save Rule'}
          </Button>
          <Button variant="secondary" onClick={() => router.back()}>
            Cancel
          </Button>
        </div>
      </div>
    </div>);
}
