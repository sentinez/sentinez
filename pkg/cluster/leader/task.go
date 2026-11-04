// Copyright 2026 Duc-Hung Ho.
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

package leader

import (
	"context"

	"golang.org/x/sync/errgroup"
)

// All runs tasks together as one Task. The first to fail cancels the
// others and ends the term; one returning nil leaves the others running.
// Nil tasks are skipped, so with none left All is nil: the node only
// leads.
func All(tasks ...Task) Task {
	var run []Task
	for _, task := range tasks {
		if task != nil {
			run = append(run, task)
		}
	}

	if len(run) == 0 {
		return nil
	}

	return func(ctx context.Context, term Term) error {
		g, ctx := errgroup.WithContext(ctx)
		for _, task := range run {
			g.Go(func() error { return task(ctx, term) })
		}

		return g.Wait()
	}
}
