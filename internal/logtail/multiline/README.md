# Multiline Matching

This package decides whether a log line starts a new multiline record.

## Priority

Multiline has two modes:

1. Manual mode: `multiline_match` is set.
   - Call path: `tailer.WithMultilinePattern(pattern)` -> `multiline.New([]string{pattern})`.
   - Only this configured pattern is used.
   - Built-in automatic patterns and `auto_multiline_extra_patterns` are not used.

2. Automatic mode: multiline is enabled and `multiline_match` is empty.
   - Call path: `tailer.WithAutoMultilineExtraPatterns(extraPatterns)` -> `multiline.NewAuto(extraPatterns)`.
   - Built-in automatic patterns are checked first.
   - `auto_multiline_extra_patterns` are supplementary and checked only after the selected built-in group misses.

## Built-In Groups

Automatic built-in patterns are split by the first byte of the line:

- `GlobalDigitPatterns`: first byte is `0-9`.
- `GlobalLetterPatterns`: first byte is `A-Z` or `a-z`.
- `GlobalSymbolPatterns`: non-alphanumeric first bytes, such as `[`, `(`, `<`, `#`, `=`, and `{`.

There is no `GlobalPatterns` catch-all list. Do not reintroduce one for automatic matching.

## Whitespace

Space and tab have a separate fast path because they are usually stack-trace or exception continuations.

Runtime flow in automatic mode:

```text
line
  -> if first byte is space or tab:
       match extraPatterns only, then return
  -> otherwise:
       choose digit / letter / symbol group by first byte
       match selected built-in group
       if built-in group misses, match extraPatterns
```

This keeps common stack traces from scanning the symbol group. If a deployment uses space- or tab-prefixed first lines, configure `auto_multiline_extra_patterns` explicitly.

## Adding Rules

- Add built-in first-line rules directly to one or more arrays in `patterns.go`.
- Keep rules anchored with `^`.
- Put each rule in the group selected by its possible first byte. A rule that can
  only match a letter-starting line does not belong in `GlobalSymbolPatterns`.
- These rules detect record starts only; they are not full log parsers.
- Be conservative with stack or exception keywords. Lines such as `Traceback`, `Exception`, `Caused by`, and indented stack frames are usually continuations.
