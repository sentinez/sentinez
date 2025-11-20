// Copyright 2025 Duc-Hung Ho.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package corerule

import (
	"testing"

	corehttp "github.com/sentinez/core/http"
)

// nolint
func TestAST(t *testing.T) {
	cmd1 := newNode(func(_ corehttp.RequestContext) bool {
		t.Log("cmd1 is running")
		return false
	})

	cmd2 := newNode(func(_ corehttp.RequestContext) bool {
		t.Log("cmd2 is running")
		return false
	})

	cmd3 := newNode(func(_ corehttp.RequestContext) bool {
		t.Log("cmd3 is running")
		return true
	})

	ans := cmd1.eval(nil) || cmd2.eval(nil) && cmd3.eval(nil)
	tree := newLogic(cmd1, logicOr, newLogic(cmd2, logicAnd, cmd3))
	if ok := tree.eval(nil); ok != ans {
		t.Error("cmd does not match with answer")
		return
	}

	t.Log("pass")
}

// nolint
func TestAST2(t *testing.T) {
	tests := []struct {
		name string
		f1   bool
		f2   bool
		f3   bool
		op1  logicType
		op2  logicType
	}{
		{"T1: all false AND", false, false, false, logicAnd, logicAnd},
		{"T2: all false OR", false, false, false, logicOr, logicOr},
		{"T3: true OR false AND false", true, false, false, logicOr, logicAnd},
		{"T4: false OR true AND true", false, true, true, logicOr, logicAnd},
		{"T5: false AND true OR true", false, true, true, logicAnd, logicOr},
		{"T6: true AND false OR true", true, false, true, logicAnd, logicOr},
		{"T7: true AND true OR false", true, true, false, logicAnd, logicOr},
		{"T8: false OR false AND true", false, false, true, logicOr, logicAnd},
		{"T9: true OR false OR false", true, false, false, logicOr, logicOr},
		{"T10: false AND false OR true", false, false, true, logicAnd, logicOr},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// mock node functions
			cmd1 := newNode(func(_ corehttp.RequestContext) bool {
				t.Log("cmd1 running:", tt.f1)
				return tt.f1
			})
			cmd2 := newNode(func(_ corehttp.RequestContext) bool {
				t.Log("cmd2 running:", tt.f2)
				return tt.f2
			})
			cmd3 := newNode(func(_ corehttp.RequestContext) bool {
				t.Log("cmd3 running:", tt.f3)
				return tt.f3
			})

			// Tính kết quả logic thật (tuần tự)
			var ans bool
			if tt.op1 == logicAnd {
				if tt.op2 == logicAnd {
					ans = cmd1.eval(nil) && cmd2.eval(nil) && cmd3.eval(nil)
				} else {
					ans = cmd1.eval(nil) && (cmd2.eval(nil) || cmd3.eval(nil))
				}
			} else {
				if tt.op2 == logicAnd {
					ans = cmd1.eval(nil) || (cmd2.eval(nil) && cmd3.eval(nil))
				} else {
					ans = cmd1.eval(nil) || cmd2.eval(nil) || cmd3.eval(nil)
				}
			}

			// Xây cây AST tương ứng
			tree := newLogic(cmd1, tt.op1, newLogic(cmd2, tt.op2, cmd3))
			got := tree.eval(nil)

			// So sánh kết quả
			if got != ans {
				t.Errorf("AST result mismatch:\nwant: %v\ngot : %v", ans, got)
			} else {
				t.Logf("pass (result = %v)", got)
			}
		})
	}
}

// nolint
func BenchmarkAST(b *testing.B) {
	cmd1 := newNode(func(_ corehttp.RequestContext) bool { return false })
	cmd2 := newNode(func(_ corehttp.RequestContext) bool { return false })
	cmd3 := newNode(func(_ corehttp.RequestContext) bool { return false })
	tree := newLogic(cmd1, logicOr, newLogic(cmd2, logicAnd, cmd3))

	for b.Loop() {
		_ = tree.eval(nil)
	}
}

// nolint
func BenchmarkAST_TableDriven(b *testing.B) {
	tests := []struct {
		name string
		f1   bool
		f2   bool
		f3   bool
		op1  logicType
		op2  logicType
	}{
		{"T1_all_false_AND", false, false, false, logicAnd, logicAnd},
		{"T2_all_false_OR", false, false, false, logicOr, logicOr},
		{"T3_true_OR_false_AND_false", true, false, false, logicOr, logicAnd},
		{"T4_false_OR_true_AND_true", false, true, true, logicOr, logicAnd},
		{"T5_false_AND_true_OR_true", false, true, true, logicAnd, logicOr},
		{"T6_true_AND_false_OR_true", true, false, true, logicAnd, logicOr},
		{"T7_true_AND_true_OR_false", true, true, false, logicAnd, logicOr},
		{"T8_false_OR_false_AND_true", false, false, true, logicOr, logicAnd},
		{"T9_true_OR_false_OR_false", true, false, false, logicOr, logicOr},
		{"T10_false_AND_false_OR_true", false, false, true, logicAnd, logicOr},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			cmd1 := newNode(func(_ corehttp.RequestContext) bool { return tt.f1 })
			cmd2 := newNode(func(_ corehttp.RequestContext) bool { return tt.f2 })
			cmd3 := newNode(func(_ corehttp.RequestContext) bool { return tt.f3 })
			tree := newLogic(cmd1, tt.op1, newLogic(cmd2, tt.op2, cmd3))

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = tree.eval(nil)
			}

			free(cmd1)
			free(cmd2)
			free(cmd3)
			free(tree)
		})
	}
}
