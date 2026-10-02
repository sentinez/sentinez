import * as React from "react"

type UseControllableStateParams<T> = {
  /** Controlled value. When defined, the component is controlled. */
  prop?: T
  /** Initial value used in uncontrolled mode. */
  defaultProp: T
  /** Called whenever the value changes, in both modes. */
  onChange?: (value: T) => void
}

/**
 * Merges controlled and uncontrolled state, mirroring Radix UI's
 * `useControllableState`.
 */
export function useControllableState<T>({
  prop,
  defaultProp,
  onChange,
}: UseControllableStateParams<T>) {
  const [uncontrolled, setUncontrolled] = React.useState(defaultProp)
  const isControlled = prop !== undefined
  const value = isControlled ? (prop as T) : uncontrolled

  const onChangeRef = React.useRef(onChange)
  React.useEffect(() => {
    onChangeRef.current = onChange
  })

  const setValue = React.useCallback(
    (next: React.SetStateAction<T>) => {
      const resolved =
        typeof next === "function" ? (next as (prev: T) => T)(value) : next
      if (Object.is(resolved, value)) return
      if (!isControlled) setUncontrolled(resolved)
      onChangeRef.current?.(resolved)
    },
    [isControlled, value]
  )

  return [value, setValue] as const
}
