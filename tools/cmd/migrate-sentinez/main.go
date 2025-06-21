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

package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	_ "github.com/joho/godotenv/autoload"
	migratepgx "github.com/sentinez/sentinez/plugins/migrate/pgx"
)

func runMigrations(action string, step int) error {
	return migratepgx.Run(
		"file://../boot/migrations/timescale",
		os.Getenv("SENTINEZ_PUBLIC_TIMESCALEDB_URL"),
		action,
		step,
	)
}

func main() {
	args := os.Args

	var types = "pgx"

	flag.StringVar(&types, "driver", types, "driver type: pgx")
	flag.Parse()

	if len(args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  go run cmd/migrate-sentinez/main.go up")
		fmt.Println("  go run cmd/migrate-sentinez/main.go down [steps]")
		os.Exit(1)
	}

	action := args[1]
	step := 1 // default step for "down"

	if action == "down" && len(args) >= 3 {
		n, err := strconv.Atoi(args[2])
		if err == nil {
			step = n
		}
	}

	switch types {
	case "pgx":
		if err := runMigrations(action, step); err != nil {
			panic(err)
		}
	}
}
