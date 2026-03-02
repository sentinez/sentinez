---
name: Go Style Guide
description: Go coding conventions and best practices based on the Uber Go Style Guide. Apply these rules when writing, reviewing, or refactoring Go code.
---

# Go Style Guide

Based on the [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md). Apply these conventions when writing Go code in this project.

---

## Guidelines

### Interfaces

- **Never use pointers to interfaces.** Pass interfaces as values; the underlying data can still be a pointer.
- **Verify interface compliance at compile time:**
  ```go
  var _ http.Handler = (*Handler)(nil)   // pointer receiver
  var _ http.Handler = LogHandler{}      // value receiver
  ```
- Methods with **value receivers** can be called on pointers and values. Methods with **pointer receivers** can only be called on pointers or addressable values.

### Mutexes

- **Zero-value mutexes are valid** — no need for `new(sync.Mutex)`.
- **Do not embed mutexes** in structs (even unexported). Use a named field:
  ```go
  type SMap struct {
      mu   sync.Mutex
      data map[string]string
  }
  ```

### Slices and Maps at Boundaries

- **Copy slices/maps** received as arguments if you store a reference, to prevent callers from mutating your internal state.
- **Copy slices/maps** before returning them, to prevent callers from mutating internal state.

### Defer

- **Use `defer`** to clean up resources (files, locks, connections). The overhead is negligible compared to readability gains.

### Channels

- Channels should have a **size of 0 (unbuffered) or 1**. Any other size requires strong justification.

### Enums

- **Start enums at one** (`iota + 1`) unless the zero value is a meaningful default:
  ```go
  const (
      Add Operation = iota + 1
      Subtract
      Multiply
  )
  ```

### Time

- Use `time.Time` for instants, `time.Duration` for periods.
- Use `time.Time` and `time.Duration` in external system interactions (JSON, SQL, flags).
- When `time.Duration` is not possible, include the unit in the field name (e.g., `IntervalMillis`).

### Errors

- **Error types decision table:**

  | Matching needed? | Message  | Use                                  |
  |------------------|----------|--------------------------------------|
  | No               | static   | `errors.New`                         |
  | No               | dynamic  | `fmt.Errorf`                         |
  | Yes              | static   | top-level `var` + `errors.New`       |
  | Yes              | dynamic  | custom `error` type                  |

- **Error wrapping:** Use `%w` if callers should match the underlying error, `%v` to obfuscate it.
- **Keep context succinct** — avoid "failed to" prefixes that pile up through the stack:
  ```go
  // Bad:  fmt.Errorf("failed to create new store: %w", err)
  // Good: fmt.Errorf("new store: %w", err)
  ```
- **Error naming:** exported errors use `Err` prefix (`ErrNotFound`), custom error types use `Error` suffix (`NotFoundError`). Unexported errors use `err` prefix (`errNotFound`).
- **Handle errors once** — don't log AND return the same error.

### Type Assertions

- **Always use the "comma ok" idiom:**
  ```go
  t, ok := i.(string)
  if !ok {
      // handle gracefully
  }
  ```

### Don't Panic

- Production code must **never panic**. Return errors instead.
- In tests, use `t.Fatal` / `t.FailNow` instead of panic.
- Exception: program initialization with `template.Must()` or similar.

### Atomics

- Prefer `go.uber.org/atomic` (or `sync/atomic` types in Go 1.19+) over raw `int32`/`int64` atomic operations for type safety.

### Globals

- **Avoid mutable globals.** Use dependency injection instead.

### Embedding in Public Structs

- **Avoid embedding types in public structs** — it leaks implementation details, inhibits type evolution, and obscures docs. Use named fields with explicit delegate methods instead.

### Built-in Names

- **Never shadow built-in identifiers** (`error`, `string`, `len`, `make`, `new`, `copy`, etc.).

### `init()`

- **Avoid `init()`** where possible. It should be deterministic, avoid I/O, and not depend on other `init()` ordering. Move logic to `main()` or explicit setup functions.

### Exit

- Call `os.Exit` or `log.Fatal` **only in `main()`**. All other functions must return errors.
- Prefer a single exit point via a `run()` function pattern:
  ```go
  func main() {
      if err := run(); err != nil {
          log.Fatal(err)
      }
  }
  ```

### Goroutines

- **Don't fire-and-forget goroutines.** Every goroutine must have a predictable stop mechanism and a way to wait for completion (`sync.WaitGroup`, done channel).
- **No goroutines in `init()`.**
- Expose an object with `Close`/`Stop`/`Shutdown` method for managed lifecycle.

### Field Tags

- **Use field tags** in marshaled structs (JSON, YAML, etc.) — field names are part of the serialized contract.

---

## Performance

*Apply only to hot paths.*

- **Prefer `strconv` over `fmt`** for primitive-to-string conversions (~2x faster).
- **Avoid repeated `[]byte("...")` conversions** — convert once and reuse.
- **Specify container capacity:**
  ```go
  m := make(map[string]T, len(items))     // map hint
  s := make([]T, 0, len(items))            // slice capacity
  ```

---

## Style

### Formatting

- **Soft line length limit: 99 characters.**
- **Be consistent** — apply changes at package level or larger.
- Run `goimports` on save; run `golint` and `go vet` for checks.

### Declarations

- **Group similar declarations** (imports, consts, vars, types). Only group **related** items.
- **Import groups:** stdlib first, then everything else (separated by blank line).
  ```go
  import (
      "fmt"
      "os"

      "github.com/example/pkg"
  )
  ```

### Naming

- **Packages:** all lowercase, no underscores, short, not plural, not generic ("common", "util", "shared").
- **Functions:** `MixedCaps`. Test functions may use underscores: `TestMyFunction_WhatIsBeingTested`.
- **Import aliasing:** only when package name doesn't match the last element of the import path, or on direct conflicts.
- **Prefix unexported globals** with `_` (e.g., `_defaultPort`). Exception: error vars use `err` prefix.

### Organization

- **Function ordering:** sorted in rough call order, grouped by receiver.
- **Exported functions** appear first, after type/const/var definitions.
- `NewXYZ()` appears after the type definition, before other receiver methods.
- Utility functions go at the end of the file.

### Control Flow

- **Reduce nesting:** handle error cases first and return early.
- **Eliminate unnecessary else:**
  ```go
  // Bad                    // Good
  var a int                 a := 10
  if b {                    if b {
      a = 100                   a = 100
  } else {                  }
      a = 10
  }
  ```
- **Reduce variable scope** — use `:=` in `if` statements where possible:
  ```go
  if err := os.WriteFile(name, data, 0644); err != nil {
      return err
  }
  ```

### Variables

- Use `:=` for local variables with explicit initial values.
- Use `var` for zero-value declarations (slices, structs):
  ```go
  var filtered []int     // not filtered := []int{}
  var user User          // not user := User{}
  ```
- **`nil` is a valid slice** — return `nil` instead of `[]int{}`. Check emptiness with `len(s) == 0`, not `s == nil`.
- Constants used in a single function should be **local**, not package-level.

### Struct Initialization

- **Always use field names** (enforced by `go vet`).
- **Omit zero-value fields** unless they provide context.
- Use `var s T` for zero-value structs, `&T{...}` for struct references (not `new(T)`).

### Parameters

- **Avoid naked bool parameters.** Use C-style comments or custom types:
  ```go
  printInfo("foo", true /* isLocal */, true /* done */)
  ```
- Use **raw string literals** (backticks) to avoid escaping.

### Maps

- Use `make(map[K]V)` for non-empty maps, `var m map[K]V` for read-only or nil maps.
- Declare maps with `make` when you want to add elements, and `var` when the map is declared but may not be used.

### Printf

- Declare format strings outside of Printf when possible (`const msg = "unexpected values %v"`).
- Name Printf-style functions with `f` suffix: `Wrapf`, `Errorf`.

---

## Patterns

### Table-Driven Tests

- Use table-driven tests with subtests for repetitive test logic.
- Convention: slice is `tests`, each case is `tt`, inputs prefixed `give`, outputs prefixed `want`.
  ```go
  tests := []struct {
      give     string
      wantHost string
      wantPort string
  }{
      {give: "192.0.2.0:8000", wantHost: "192.0.2.0", wantPort: "8000"},
  }

  for _, tt := range tests {
      t.Run(tt.give, func(t *testing.T) {
          host, port, err := net.SplitHostPort(tt.give)
          require.NoError(t, err)
          assert.Equal(t, tt.wantHost, host)
          assert.Equal(t, tt.wantPort, port)
      })
  }
  ```
- **Avoid complex table tests** — no conditional assertions, no branching logic inside the loop. Split into separate tests instead.

### Functional Options

- Use the **functional options pattern** for constructors with 3+ optional parameters:
  ```go
  type Option interface {
      apply(*options)
  }

  func Open(addr string, opts ...Option) (*Connection, error)
  ```
- Prefer `Option` interface over closures for testability and debugging.
