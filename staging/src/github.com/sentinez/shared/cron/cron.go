// Copyright 2025 Sentinéz Labs.
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

package cron

import (
	"context"
	"time"
)

// Start starts a cron job that runs at a given interval.
// Parameters:
// - ctx: used to cancel the job when needed
// - interval: the duration between each job run
// - job: the function to execute on each interval
//
// The returned channel is closed once the job has stopped, after ctx is
// done.
func Start(
	ctx context.Context, interval time.Duration, job func(),
) <-chan struct{} {
	ticker := time.NewTicker(interval)
	done := make(chan struct{})
	job()

	go func() {
		defer close(done)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				job()
			case <-ctx.Done():
				return
			}
		}
	}()

	return done
}
