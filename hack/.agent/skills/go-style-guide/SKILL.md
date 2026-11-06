---
name: go-style-guide
description: Go coding conventions and best practices based on the Uber Go Style Guide. Apply these rules when writing, reviewing, or refactoring Go code.
---

# Go Style Guide

Based on the [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md). Apply these conventions when writing Go code in this project.

## Project overrides

These win over the upstream guide wherever they differ:

- **Line length: 80 columns (tab = 4), enforced by `lll` in `.golangci.yaml`; function length limit 35 lines (`funlen`).** Upstream only recommends a soft 99-character limit.
- **Identifiers are spelled `Id`, not `ID`** (`resourceId`, `tenantId`, `UserId`), matching the protobuf-generated names.
- **Atomics:** the `sync/atomic` types (`atomic.Bool`, `atomic.Int64`, ...) are fine in place of `go.uber.org/atomic`.

---

## Guidelines

### Pointers to Interfaces

- **Never use pointers to interfaces.** Pass interfaces as values; the underlying data can still be a pointer.
- If interface methods must modify the underlying data, store a pointer in the interface.

### Verify Interface Compliance

- **Verify interface compliance at compile time** where appropriate:
  - exported types that must implement an interface as part of their API contract;
  - exported or unexported types that belong to a collection of types implementing the same interface;
  - any other case where violating the interface would break users.
- The right-hand side is the zero value of the asserted type: `nil` for pointers, slices and maps, an empty struct for struct types.
  ```go
  var _ http.Handler = (*Handler)(nil)   // pointer receiver
  var _ http.Handler = LogHandler{}      // value receiver
  ```

### Receivers and Interfaces

- Methods with **value receivers** can be called on pointers and values. Methods with **pointer receivers** can only be called on pointers or addressable values.
- An interface can be satisfied by a pointer even if the method has a value receiver.

### Zero-value Mutexes are Valid

- The zero value of `sync.Mutex` and `sync.RWMutex` is valid: no `new(sync.Mutex)`, no pointer to a mutex.
- If a struct is used by pointer, the mutex is a **non-pointer field** on it.
- **Do not embed mutexes** in structs, even unexported ones. Use a named field:
  ```go
  type SMap struct {
      mu   sync.Mutex
      data map[string]string
  }
  ```

### Copy Slices and Maps at Boundaries

- **Copy slices/maps received as arguments** if you store a reference, so callers cannot mutate your internal state.
- **Copy slices/maps before returning them**, so callers cannot mutate internal state.

### Defer to Clean Up

- **Use `defer`** to clean up resources such as files and locks. Its overhead only matters for functions that run in nanoseconds; readability wins everywhere else.

### Channel Size is One or None

- Channels should be **unbuffered or have a size of one**. Any other size needs a high level of scrutiny: how the size is determined, what prevents the channel from filling up under load and blocking writers, and what happens when it does.

### Start Enums at One

- **Start enums at one** (`iota + 1`), since variables default to 0:
  ```go
  const (
      Add Operation = iota + 1
      Subtract
      Multiply
  )
  ```
- Start at zero only when the zero value is the desirable default behavior.

### Use `"time"` to Handle Time

- Always use the `time` package: `time.Time` for instants (and its methods to compare, add, subtract), `time.Duration` for periods.
- Pick the method by intent: `Time.AddDate` for "same time of day on the next calendar day", `Time.Add` for "exactly 24 hours later".
- Use `time.Time` and `time.Duration` with external systems when possible (flags, JSON, SQL, YAML).
- When `time.Duration` is not possible, use `int` or `float64` and **put the unit in the field name** (`IntervalMillis`).
- When `time.Time` is not possible, use a `string` formatted as RFC 3339 (`time.RFC3339`) unless an alternative is agreed upon.

### Errors

#### Error Types

- Decide by two questions: does the caller need to match the error, and is the message static or dynamic?

  | Error matching? | Error message | Guidance                            |
  |-----------------|---------------|-------------------------------------|
  | No              | static        | `errors.New`                        |
  | No              | dynamic       | `fmt.Errorf`                        |
  | Yes             | static        | top-level `var` with `errors.New`   |
  | Yes             | dynamic       | custom `error` type                 |

- Exported error variables and types become part of the package's public API.

#### Error Wrapping

- Three options when a call fails:
  - **return the original error as-is** when there is no context to add;
  - **add context with `%w`** when the caller should be able to match the underlying error. This is a good default, but callers may come to rely on it: when the wrapped error is a known `var` or type, document and test it as part of the contract;
  - **add context with `%v`** to obfuscate the underlying error (you can switch to `%w` later).
- **Keep context succinct**: avoid phrases like "failed to", which state the obvious and pile up through the stack:
  ```go
  // Bad:  fmt.Errorf("failed to create new store: %w", err)
  // Good: fmt.Errorf("new store: %w", err)
  ```
- Once the error is sent to another system, make it clear the message is an error (an `err` tag or a "Failed" prefix in logs).

#### Error Naming

- Global error values use the prefix `Err` (exported) or `err` (unexported). This supersedes the `_` prefix for unexported globals.
- Custom error types use the suffix `Error` (`NotFoundError`, `resolveError`).

#### Handle Errors Once

- Handle each error **only once**. Do not log an error and then return it: callers further up will handle it too, which only adds noise to the logs.
- Valid ways to handle it:
  - wrap the error and return it;
  - log the error and degrade gracefully, when the operation is not strictly necessary;
  - match a specific error with `errors.Is`/`errors.As` and degrade gracefully, wrapping and returning every other case;
  - return a well-defined error for a domain-specific failure.

### Handle Type Assertion Failures

- The single-value form panics on an incorrect type. **Always use the "comma ok" idiom:**
  ```go
  t, ok := i.(string)
  if !ok {
      // handle the error gracefully
  }
  ```

### Don't Panic

- Production code must **avoid panics**: return an error and let the caller decide. Panic/recover is not an error handling strategy.
- Panic only when something irrecoverable happens (such as a nil dereference). Exception: program initialization, where a failure that should abort the program may panic (`template.Must(...)`).
- In tests, prefer `t.Fatal` / `t.FailNow` over panics so the test is marked as failed.

### Atomics

- Raw `sync/atomic` functions operate on `int32`/`int64`, so it is easy to forget the atomic operation on a read or write. Use typed atomics (`go.uber.org/atomic` upstream; see the project override above).

### Avoid Mutable Globals

- **Avoid mutating global variables**; use dependency injection instead. This applies to function pointers as well as other values.

### Avoid Embedding Types in Public Structs

- Embedded types **leak implementation details, inhibit type evolution, and obscure documentation**. Use a named field and hand-write the delegate methods.
- This holds for embedded structs and embedded interfaces alike:
  - adding methods to an embedded interface is a breaking change;
  - removing methods from an embedded struct is a breaking change;
  - removing the embedded type is a breaking change;
  - replacing the embedded type, even with one that satisfies the same interface, is a breaking change.

### Avoid Using Built-In Names

- **Never use predeclared identifiers as names** (`error`, `string`, `len`, `make`, `new`, `copy`, ...). They shadow the original in the current scope or make the code confusing; `go vet` points these out, the compiler does not.

### Avoid `init()`

- **Avoid `init()` where possible.** When it is unavoidable, it must:
  1. be completely deterministic, regardless of program environment or invocation;
  2. not depend on the ordering or side effects of other `init()` functions;
  3. not access or manipulate global or environment state (machine information, environment variables, working directory, program arguments);
  4. avoid I/O: filesystem, network, and system calls.
- Code that cannot meet these belongs in a helper called from `main()` or elsewhere in the program's lifecycle. Libraries must not perform "init magic".
- Cases where `init()` may be preferable: complex expressions that cannot be single assignments, pluggable hooks (`database/sql` dialects, encoding registries), deterministic precomputation.

### Exit in Main

- Call `os.Exit` or `log.Fatal*` **only in `main()`**. All other functions return errors. Exiting elsewhere gives non-obvious control flow, is hard to test, and skips `defer`red cleanup.
- **Exit once:** prefer at most one `os.Exit`/`log.Fatal` in `main()`, with the logic in a separate, testable function:
  ```go
  func main() {
      if err := run(); err != nil {
          log.Fatal(err)
      }
  }
  ```
- `run()` is not prescriptive: it may take arguments, return an exit code, or live in another package. The only requirement is a single place in `main()` that exits the process.

### Use Field Tags in Marshaled Structs

- Any struct field marshaled into JSON, YAML, or another tag-based format **must carry the relevant tag**. The serialized form is a contract; tags make it explicit and guard against breaking it by renaming fields.

### Don't Fire-and-Forget Goroutines

- **Do not leak goroutines** in production code. Use `go.uber.org/goleak` to test for leaks in packages that spawn goroutines.
- Every goroutine must have a predictable time at which it stops, or a way to signal it to stop. In both cases there must be a way to **block and wait for it to finish**:
  - several goroutines: `sync.WaitGroup` (`wg.Go(...)`, then `wg.Wait()`);
  - a single goroutine: a `done := make(chan struct{})` it closes with `defer close(done)`, then `<-done`.
- **No goroutines in `init()`.** A package that needs a background goroutine exposes an object managing its lifetime, with a `Close`/`Stop`/`Shutdown` method that signals it to stop and waits for it to exit.

---

## Performance

*Apply only to the hot path.*

- **Prefer `strconv` over `fmt`** when converting primitives to/from strings.
- **Avoid repeated string-to-byte conversions**: convert a fixed string once and reuse the result.
- **Specify container capacity** where possible:
  ```go
  m := make(map[string]T, len(items))     // hint: approximates the buckets
  s := make([]T, 0, len(items))            // exact: no allocation until full
  ```
  A map hint does not guarantee preemptive allocation; a slice capacity does.

---

## Style

### Avoid Overly Long Lines

- Avoid lines that require horizontal scrolling. See the project override: 80 columns here.

### Be Consistent

- Above all else, **be consistent**. Apply a change of convention at package level or larger, never to part of a package.

### Group Similar Declarations

- **Group similar declarations** (imports, consts, vars, types), inside functions too. **Only group related** items.
- Exception: variables declared adjacent to each other, particularly inside functions, are grouped together even if unrelated.

### Import Group Ordering

- **Two import groups**: standard library, then everything else (the `goimports` default).
  ```go
  import (
      "fmt"
      "os"

      "go.uber.org/atomic"
      "golang.org/x/sync/errgroup"
  )
  ```

### Package Names

- All lower-case, no capitals or underscores.
- Does not need to be renamed with a named import at most call sites.
- Short and succinct; the name appears in full at every call site.
- Not plural (`net/url`, not `net/urls`).
- Not "common", "util", "shared", or "lib".

### Function Names

- `MixedCaps`. Test functions may contain underscores to group related cases: `TestMyFunction_WhatIsBeingTested`.

### Import Aliasing

- **Required** when the package name does not match the last element of the import path.
- Otherwise avoid aliases unless there is a direct conflict between imports.

### Function Grouping and Ordering

- Functions are sorted in **rough call order** and **grouped by receiver**.
- Exported functions appear first, after `struct`, `const`, `var` definitions.
- `newXYZ()`/`NewXYZ()` may appear after the type definition, before the rest of the receiver's methods.
- Plain utility functions go towards the end of the file.

### Reduce Nesting

- Handle error cases and special conditions first, and **return early or `continue`**. Keep deeply nested code to a minimum.

### Unnecessary Else

- If a variable is set in both branches of an `if`, use a single `if`:
  ```go
  // Bad                    // Good
  var a int                 a := 10
  if b {                    if b {
      a = 100                   a = 100
  } else {                  }
      a = 10
  }
  ```

### Top-level Variable Declarations

- At the top level use the `var` keyword and **do not specify the type** unless it differs from the type of the expression:
  ```go
  var _s = F()            // not: var _s string = F()
  var _e error = F()      // F returns myError, but we want error
  ```

### Prefix Unexported Globals with _

- **Prefix unexported top-level `var`s and `const`s with `_`** (e.g., `_defaultPort`), so it is clear at the use site that they are global symbols and a generic name is not picked up by accident from another file.
- Exception: unexported error values use the prefix `err` without the underscore.

### Embedding in Structs

- Embedded types go **at the top of the field list**, separated from regular fields by an empty line.
- Embedding must give a tangible benefit in a semantically appropriate way, with zero adverse user-facing effects. Mutexes are never embedded.
- Embedding should not:
  - be purely cosmetic or convenience-oriented;
  - make the outer type harder to construct or use;
  - affect the outer type's zero value or copy semantics;
  - expose unrelated functions or fields, unexported types, or implementation details;
  - change the outer type's API or type semantics;
  - let users observe or control type internals;
  - change the behavior of inner functions in a way that would surprise users.
- Litmus test: "would all of these exported inner methods/fields be added directly to the outer type?" If the answer is "some" or "no", use a field instead.

### Local Variable Declarations

- Use `:=` when a variable is set to a value explicitly.
- Use `var` where the default value reads clearer, such as an empty slice:
  ```go
  var filtered []int     // not filtered := []int{}
  ```

### nil is a Valid Slice

- **Return `nil`**, not an explicit zero-length slice (`[]int{}`).
- **Check emptiness with `len(s) == 0`**, not `s == nil`.
- A slice declared with `var` is usable immediately, without `make()`.
- A nil slice is still not equivalent to an allocated empty slice; the two may be treated differently, for example in serialization.

### Reduce Scope of Variables

- Reduce the scope of variables and constants where possible, unless it conflicts with reducing nesting:
  ```go
  if err := os.WriteFile(name, data, 0644); err != nil {
      return err
  }
  ```
- Do not reduce the scope when the result is needed outside the `if`.
- Constants stay local unless they are used in several functions or files, or are part of the package's external contract.

### Avoid Naked Parameters

- Add C-style comments for parameter names when their meaning is not obvious:
  ```go
  printInfo("foo", true /* isLocal */, true /* done */)
  ```
- Better: replace naked `bool` parameters with custom types, which are more readable, type-safe, and allow more than two states later.

### Use Raw String Literals to Avoid Escaping

- Use raw string literals (backticks) instead of hand-escaped strings.

### Initializing Structs

- **Use field names** (enforced by `go vet`). Exception: field names may be omitted in test tables with 3 or fewer fields.
- **Omit zero-value fields** unless they provide meaningful context, as in test tables.
- **Use `var s T`** when all fields are omitted, not `s := T{}`.
- **Use `&T{...}`** for struct references, not `new(T)`.

### Initializing Maps

- Use `make(map[K]V)` for empty maps and maps populated programmatically, with a size hint when available; `var m map[K]V` only declares.
- Use a map literal when the map holds a fixed set of elements.

### Format Strings outside Printf

- A format string declared outside the call must be a `const`, so `go vet` can analyze it.

### Naming Printf-style Functions

- Use the predefined `Printf`-style names when possible; otherwise end the name with `f` (`Wrapf`, not `Wrap`) so `go vet` can check the format string.

---

## Patterns

### Test Tables

- Use table-driven tests with subtests when the core test logic is repetitive across multiple conditions.
- Convention: the slice is `tests`, each case is `tt`, inputs are prefixed `give`, outputs `want`.
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
- **Avoid unnecessary complexity.** Do not use a table when subtests need complex or conditional logic; split into several tables or separate `Test...` functions. Aim to:
  - focus on the narrowest unit of behavior;
  - minimize "test depth" and avoid conditional assertions;
  - use every table field in every test;
  - run all test logic for every case.
- Warning signs: several branching fields (`shouldError`, `expectCall`), many `if` statements for mock expectations, functions inside the table (`setupMocks func(*FooMock)`).
- A single branch for success versus failure (a `shouldErr`-style field) is acceptable when the test body is short and straightforward.
- **Parallel tests** (and loops that spawn goroutines or capture references) must make sure each iteration holds the expected loop variable.

### Functional Options

- Use functional options for optional arguments in constructors and other public APIs you expect to expand, especially with three or more arguments already.
- Implement them with an `Option` interface holding an unexported method that records onto an unexported `options` struct:
  ```go
  type options struct {
      cache  bool
      logger *zap.Logger
  }

  type Option interface {
      apply(*options)
  }

  func Open(addr string, opts ...Option) (*Connection, error)
  ```
- Prefer this over closures: options can be compared in tests and mocks, and can implement other interfaces such as `fmt.Stringer`.

---

## Linting

- Lint consistently across the codebase. Minimum set: `errcheck`, `goimports`, `revive`, `govet`, `staticcheck`, run through `golangci-lint`.
- This repo runs them per module with `.golangci.yaml` (plus `lll`, `funlen`, `misspell`); see `make lint`.
