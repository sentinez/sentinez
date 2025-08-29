# Uber Go Style Guide - Cheatsheet

## 1. General

-   Clear \> Clever. Consistency \> Personal.
-   Prefer readability over brevity.

## 2. Packages & Imports

-   Lowercase package names, no underscore/dash.
-   Avoid vague names (`util`, `common`).
-   Import groups: stdlib, external, internal.

## 3. Variables & Constants

-   Short names for small scope, descriptive for larger.
-   MixedCaps, not snake_case.
-   Use const for constants, group with iota when needed.

## 4. Error Handling

-   Don't panic for normal errors.
-   Error messages: lowercase, no period.
-   Use errors.Is / errors.As.
-   Wrap with fmt.Errorf("context: %w", err).

## 5. Concurrency

-   Avoid goroutine leaks → use context or stop channels.
-   Close channels in sender, not receiver.
-   Use buffered channels when appropriate.

## 6. Slices & Maps

-   Preallocate capacity with make(\[\]T, 0, n).
-   Always check existence in maps (val, ok := m\[key\]).
-   Declare maps with make, not nil.

## 7. Structs & Interfaces

-   Export struct fields only if needed.
-   Keep interfaces small (1-2 methods).
-   Define interfaces at consumer side.

## 8. Functions

-   Avoid overusing named returns.
-   Use struct for multiple params of same type.
-   Function name should reflect side effects.

## 9. Testing

-   TestXxx naming.
-   Prefer table-driven tests.
-   Avoid global state.
-   Use t.Helper() for helpers.

## 10. Performance & Safety

-   Prefer immutability.
-   Avoid unnecessary allocations, use sync.Pool.
-   Always pass context.Context as first param in APIs.