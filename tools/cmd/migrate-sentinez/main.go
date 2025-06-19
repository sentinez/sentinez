package main

import (
	"fmt"
	"os"
	"strconv"

	_ "github.com/joho/godotenv/autoload"
	migratepgx "github.com/sentinez/sentinez/mods/migrate/pgx"
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

	if err := runMigrations(action, step); err != nil {
		panic(err)
	}
}
