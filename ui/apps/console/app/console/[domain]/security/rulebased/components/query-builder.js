'use client';
import React, { useState } from 'react';
import { Card, CardContent } from '@sentinez/ui/components/card';
import { Button } from '@sentinez/ui/components/button';
import { Input } from '@sentinez/ui/components/input';
import { Plus, Trash2 } from 'lucide-react';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue, } from '@sentinez/ui/components/select';
import { transformUiToApi } from '@/lib/api/security';
const generateId = () => Math.random().toString(36).substring(2, 9);
export function QueryBuilder({ initialQuery, onChange, layout = 'vertical' }) {
    const [query, setQuery] = useState(initialQuery || {
        id: generateId(),
        combinator: 'and',
        rules: [],
    });
    React.useEffect(() => {
        if (initialQuery) {
            setQuery(initialQuery);
        }
    }, [initialQuery]);
    const notifyChange = (newQuery) => {
        setQuery(newQuery);
        if (onChange)
            onChange(newQuery);
    };
    const updateGroup = (groupId, updater, currentGroup = query) => {
        if (currentGroup.id === groupId) {
            return updater(currentGroup);
        }
        return {
            ...currentGroup,
            rules: currentGroup.rules.map((rule) => {
                if ('combinator' in rule) {
                    return updateGroup(groupId, updater, rule);
                }
                return rule;
            }),
        };
    };
    const updateRule = (ruleId, updater, currentGroup = query) => {
        return {
            ...currentGroup,
            rules: currentGroup.rules.map((rule) => {
                if ('combinator' in rule) {
                    return updateRule(ruleId, updater, rule);
                }
                if (rule.id === ruleId) {
                    return updater(rule);
                }
                return rule;
            }),
        };
    };
    const removeNode = (nodeId, currentGroup = query) => {
        if (currentGroup.id === nodeId)
            return null;
        const newRules = currentGroup.rules
            .map((rule) => {
            if ('combinator' in rule) {
                return removeNode(nodeId, rule);
            }
            return rule.id === nodeId ? null : rule;
        })
            .filter(Boolean);
        return { ...currentGroup, rules: newRules };
    };
    const addRule = (groupId) => {
        const newGroup = updateGroup(groupId, (group) => ({
            ...group,
            rules: [...group.rules, { id: generateId(), field: '', operator: '==', value: '' }],
        }));
        notifyChange(newGroup);
    };
    const addGroup = (groupId) => {
        const newGroup = updateGroup(groupId, (group) => ({
            ...group,
            rules: [...group.rules, { id: generateId(), combinator: 'and', rules: [] }],
        }));
        notifyChange(newGroup);
    };
    const SelectMenu = ({ value, onChange, options, className = '' }) => (<Select value={value} onValueChange={onChange}>
      <SelectTrigger className={`h-9 ${className}`}>
        <SelectValue placeholder="Select..."/>
      </SelectTrigger>
      <SelectContent>
        {options.map((opt) => (<SelectItem key={opt.value} value={opt.value}>
            {opt.label}
          </SelectItem>))}
      </SelectContent>
    </Select>);
    const renderGroup = (group, isRoot = false) => {
        return (<Card key={group.id} className={`w-full overflow-hidden ${isRoot ? 'border-primary/20' : 'border-dashed mt-4'}`}>
        <div className="flex flex-wrap items-center gap-2 p-3 bg-muted/40 border-b border-border/50">
          <SelectMenu value={group.combinator} onChange={(val) => notifyChange(updateGroup(group.id, (g) => ({ ...g, combinator: val })))} options={[
                { label: 'AND', value: 'and' },
                { label: 'OR', value: 'or' },
            ]} className="w-24 font-bold"/>
          <Button variant="outline" size="sm" onClick={() => addRule(group.id)}>
            <Plus className="w-4 h-4 mr-1"/> Rule
          </Button>
          <Button variant="outline" size="sm" onClick={() => addGroup(group.id)}>
            <Plus className="w-4 h-4 mr-1"/> Group
          </Button>
          {!isRoot && (<Button variant="ghost" size="icon" className="ml-auto text-destructive" onClick={() => {
                    const res = removeNode(group.id);
                    if (res)
                        notifyChange(res);
                }}>
              <Trash2 className="w-4 h-4"/>
            </Button>)}
        </div>
        <CardContent className="p-4 flex flex-col gap-3">
          {group.rules.length === 0 && (<div className="text-sm text-muted-foreground text-center py-2">
              No conditions in this group.
            </div>)}
          {group.rules.map((rule) => {
                if ('combinator' in rule) {
                    return renderGroup(rule);
                }
                return (<div key={rule.id} className="flex flex-col gap-2 p-3 border rounded-md bg-background">
                <div className="flex flex-wrap md:flex-nowrap items-center gap-2">
                  <SelectMenu value={rule.field.startsWith('http.header.')
                        ? 'http.header'
                        : rule.field.startsWith('http.query.')
                            ? 'http.query'
                            : rule.field} onChange={(val) => {
                        let newField = val;
                        if (val === 'http.header')
                            newField = 'http.header.New-Header';
                        if (val === 'http.query')
                            newField = 'http.query.param';
                        notifyChange(updateRule(rule.id, (r) => ({ ...r, field: newField })));
                    }} options={[
                        { label: 'HTTP Method', value: 'http.method' },
                        { label: 'HTTP Host', value: 'http.host' },
                        { label: 'HTTP Path', value: 'http.path' },
                        { label: 'HTTP Header', value: 'http.header' },
                        { label: 'HTTP Query', value: 'http.query' },
                        { label: 'Source IP', value: 'ip.src' },
                    ]} className="w-full md:w-48"/>

                  {(rule.field.startsWith('http.header.') ||
                        rule.field.startsWith('http.query.')) && (<Input placeholder="Key (e.g. User-Agent)" value={rule.field.split('.').slice(2).join('.')} onChange={(e) => {
                            const prefix = rule.field.split('.').slice(0, 2).join('.');
                            notifyChange(updateRule(rule.id, (r) => ({
                                ...r,
                                field: `${prefix}.${e.target.value}`,
                            })));
                        }} className="w-full md:w-48 font-mono text-xs"/>)}

                  <SelectMenu value={rule.operator} onChange={(val) => notifyChange(updateRule(rule.id, (r) => ({ ...r, operator: val })))} options={[
                        { label: '==', value: '==' },
                        { label: '!=', value: '!=' },
                        { label: '>', value: '>' },
                        { label: '<', value: '<' },
                        { label: 'contains', value: 'contains' },
                        { label: 'startsWith', value: 'startsWith' },
                        { label: 'endsWith', value: 'endsWith' },
                        { label: 'matches', value: 'matches' },
                        { label: 'in', value: 'in' },
                    ]} className="w-full md:w-32"/>

                  <Input placeholder="Value" value={rule.value} onChange={(e) => notifyChange(updateRule(rule.id, (r) => ({ ...r, value: e.target.value })))} className="w-full flex-grow"/>

                  <Button variant="ghost" size="icon" className="text-destructive shrink-0" onClick={() => {
                        const res = removeNode(rule.id);
                        if (res)
                            notifyChange(res);
                    }}>
                    <Trash2 className="w-4 h-4"/>
                  </Button>
                </div>
              </div>);
            })}
        </CardContent>
      </Card>);
    };
    return (<div className={layout === 'horizontal' ? 'grid grid-cols-1 lg:grid-cols-5 gap-6' : 'flex flex-col gap-4'}>
      <div className={layout === 'horizontal' ? 'lg:col-span-3' : ''}>
        {renderGroup(query, true)}
      </div>
      <div className={`p-4 bg-muted/20 rounded-xl border h-fit sticky top-4 ${layout === 'horizontal' ? 'lg:col-span-2' : ''}`}>
        <div className="flex items-center justify-between mb-2">
          <h4 className="text-sm font-semibold">Generated Expression JSON</h4>
        </div>
        <pre className="text-xs text-muted-foreground overflow-auto max-h-[600px] scrollbar-thin">
          {JSON.stringify(transformUiToApi(query), null, 2)}
        </pre>
      </div>
    </div>);
}
