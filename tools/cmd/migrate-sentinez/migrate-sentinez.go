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
	"fmt"
	"os"
	"strconv"

	flag "github.com/spf13/pflag"

	_ "github.com/joho/godotenv/autoload"
	"github.com/sentinez/sentinez/api/gen/go/sentinez/common/v1"
	migratepgx "github.com/sentinez/sentinez/tools/internal/migrate/pgx"
)

var (
	timescale  = os.Getenv(common.SNTZPublic_SNTZ_PUBLIC_TIMESCALEDB.String())
	postgresql = os.Getenv(common.SNTZPublic_SNTZ_PUBLIC_POSTGRES.String())
	clickhouse = os.Getenv(common.SNTZPublic_SNTZ_PUBLIC_CLICKHOUSE.String())
)

var sourceFileMap = map[string]string{
	timescale:  "file://../deploy/migrate/timescale",
	postgresql: "file://../deploy/migrate/postgresql",
	clickhouse: "file://../deploy/migrate/clickhouse",
}

func runMigrations(srcFile, dbUrl, action string, step int) error {
	return migratepgx.Run(srcFile, dbUrl, action, step)
}

func main() {
	var (
		driver      = "postgresql"
		usageDriver = "postgresql|timescale|clickhouse"
	)

	flag.StringVarP(&driver, "driver", "d", driver, usageDriver)

	flag.Usage = func() {
		fmt.Println("Usage:")
		fmt.Println("  go run cmd/migrate-sentinez/main.go up --driver=postgresql")
		fmt.Println("  go run cmd/migrate-sentinez/main.go down [steps] --driver=clickhouse")
		fmt.Println("Options:")
		flag.PrintDefaults()

		os.Exit(0)
	}

	flag.Parse()

	args := os.Args
	if len(args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  go run cmd/migrate-sentinez/main.go up")
		fmt.Println("  go run cmd/migrate-sentinez/main.go down [steps]")

		return
	}

	action := args[1]
	step := 1 // default step for "down"

	if action == "help" {
		fmt.Println("Usage:")
		fmt.Println("  go run cmd/migrate-sentinez/main.go up")
		fmt.Println("  go run cmd/migrate-sentinez/main.go down [steps]")

		return
	}

	if action == "down" && len(args) >= 3 {
		n, err := strconv.Atoi(args[2])
		if err == nil {
			step = n
		}
	}

	switch driver {
	case "postgresql":
		src := sourceFileMap[postgresql]
		if err := runMigrations(src, postgresql, action, step); err != nil {
			panic(err)
		}
	case "timescale":
		src := sourceFileMap[timescale]
		if err := runMigrations(src, timescale, action, step); err != nil {
			panic(err)
		}
	case "clickhouse":
		src := sourceFileMap[clickhouse]
		if err := runMigrations(src, clickhouse, action, step); err != nil {
			panic(err)
		}
	default:
		fmt.Printf("Unknown driver type: %s\n", driver)
		fmt.Println("Supported types: postgresql, timescale, clickhouse")
	}
}
