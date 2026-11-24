// Copyright 2026 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package secure

import (
	"errors"
	"fmt"

	"github.com/sentinez/core/szhttp"
	"github.com/sentinez/pkg/szlog"
	"github.com/sentinez/scripting"
	rulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/rule/v1"
	secrulepb "github.com/sentinez/sentinez/api/proto/sentinez/types/secrule/v1"
	"github.com/sentinez/sentinez/internal/memory/secrules"
	"github.com/sentinez/sentinez/pkg/pools/exchange"
	"google.golang.org/protobuf/types/known/structpb"
)

// _maxKeyLen the longest key a limiter script returns, in bytes.
const _maxKeyLen = 256

var (
	errScriptResult = errors.New("rule must return an action name or table")
	errScriptAction = errors.New("unknown action")
	errScriptKey    = errors.New("key must be a string, true or nil")
	errScriptKeyLen = errors.New("key is too long")
)

// _scriptActions the actions a script names by a string: read only, shared
// by the requests like _rateLimited.
var _scriptActions = map[string]*rulepb.Action{
	"block":         {Type: rulepb.ActionType_ACTION_TYPE_BLOCK},
	"log":           {Type: rulepb.ActionType_ACTION_TYPE_LOG},
	"modify_header": {Type: rulepb.ActionType_ACTION_TYPE_MODIFY_HEADER},
	"redirect":      {Type: rulepb.ActionType_ACTION_TYPE_REDIRECT},
	"set_tag":       {Type: rulepb.ActionType_ACTION_TYPE_SET_TAG},
	"route_to":      {Type: rulepb.ActionType_ACTION_TYPE_ROUTE_TO},
	"challenge":     {Type: rulepb.ActionType_ACTION_TYPE_CHALLENGE},
}

// scriptRun the exchange the rule scripts of a request read it through,
// acquired by the first one that runs: a request without script rules
// takes none. The zero value is ready; Close releases it.
type scriptRun struct {
	ctx szhttp.Context
	ex  *exchange.Exchange
}

func (r *scriptRun) exchange() *exchange.Exchange {
	if r.ex == nil {
		r.ex = exchange.Acquire(r.ctx)
	}

	return r.ex
}

func (r *scriptRun) Close() {
	if r.ex != nil {
		exchange.Release(r.ex)
		r.ex = nil
	}
}

// match reports whether e matches the request, and the action it takes.
// A script that fails does not match: the request goes on.
func (r *scriptRun) match(e secrules.Entry, mr *rulepb.MatchedRules,
) (*rulepb.Action, bool) {
	if e.Script == nil {
		return e.Rule.GetAction(), e.Eval(r.ctx, mr)
	}

	v, err := e.Script.Invoke(r.ctx.Context(), scripting.MatchFunc,
		r.exchange())
	if err != nil {
		szlog.Warnf("secrule: request %s: %v", r.ctx.RequestId(), err)
		return nil, false
	}

	action, err := scriptAction(v)
	if err != nil {
		szlog.Warnf("secrule: request %s: script %s: %v",
			r.ctx.RequestId(), e.Rule.GetName(), err)
		return nil, false
	}

	if action == nil {
		return nil, false
	}

	matched(mr, e.Rule)

	return action, true
}

// key the key a limiter script counts the request under: ok is false when
// it does not count it, and ip the client IP for a script answering true.
func (r *scriptRun) key(p *scripting.Program, ip string,
) (key string, ok bool) {
	v, err := p.Invoke(r.ctx.Context(), scripting.KeyFunc, r.exchange())
	if err != nil {
		szlog.Warnf("limiter: request %s: %v", r.ctx.RequestId(), err)
		return "", false
	}

	key, ok, err = scriptKey(v, ip)
	if err != nil {
		szlog.Warnf("limiter: request %s: script %s: %v",
			r.ctx.RequestId(), p.Name(), err)
	}

	return key, ok
}

// matched records the script rule as matched by the request.
func matched(mr *rulepb.MatchedRules, rule *secrulepb.SecRule) {
	mr.Ids = append(mr.Ids, rule.GetId())
	mr.Names = append(mr.Names, rule.GetName())
}

// scriptKey the key of the value a limiter script returned.
func scriptKey(v scripting.Value, ip string) (string, bool, error) {
	switch v.Kind {
	case scripting.KindNil:
		return "", false, nil
	case scripting.KindBool:
		if v.Bool {
			return ip, true, nil
		}

		return "", false, nil
	case scripting.KindString:
		if len(v.Str) > _maxKeyLen {
			return "", false, errScriptKeyLen
		}

		return v.Str, v.Str != "", nil
	}

	return "", false, errScriptKey
}

// scriptAction the action a secrule script returned: nil for no match.
func scriptAction(v scripting.Value) (*rulepb.Action, error) {
	switch v.Kind {
	case scripting.KindNil:
		return nil, nil
	case scripting.KindBool:
		if v.Bool {
			return nil, errScriptResult
		}

		return nil, nil
	case scripting.KindString:
		return namedAction(v.Str)
	case scripting.KindTable:
		return tableAction(v.Table)
	}

	return nil, errScriptResult
}

func namedAction(name string) (*rulepb.Action, error) {
	a, ok := _scriptActions[name]
	if !ok {
		return nil, fmt.Errorf("%w %q", errScriptAction, name)
	}

	return a, nil
}

// tableAction {type = "redirect", params = {location = "/"}}
func tableAction(t map[string]any) (*rulepb.Action, error) {
	name, _ := t["type"].(string)

	base, err := namedAction(name)
	if err != nil {
		return nil, err
	}

	action := &rulepb.Action{Type: base.GetType()}
	if params, ok := t["params"].(map[string]any); ok {
		action.Params, err = structpb.NewStruct(params)
		if err != nil {
			return nil, fmt.Errorf("params: %w", err)
		}
	}

	return action, nil
}
