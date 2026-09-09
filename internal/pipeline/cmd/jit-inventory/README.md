# Upstream inventory (not a compatibility report)

From the DataKit repository root:

```sh
go test ./internal/pipeline/cmd/jit-inventory
go run ./internal/pipeline/cmd/jit-inventory \
  -upstream "$(go list -m -f '{{.Dir}}' github.com/GuanceCloud/pipeline-go)" \
  -datakit . \
  -out internal/pipeline/testdata/pipeline-go-inventory.json
```

The generated baseline currently corresponds to pipeline-go v1.4.3. It contains
76 `FuncsMap` names (including aliases), 235 Test entry candidates, and 688 literal
case-label occurrences. These are **not** counts of passing tests or fully expanded
table cases. No upstream tests are executed by this command.

The scanner uses Go ASTs, preserves duplicate case labels, and reports relative
file/line locations. It records named table fields and literal `.Run` arguments;
dynamic labels, generated cases, positional table entries, helpers and imported
fixtures still need manual mapping. The literal-label list can include labels
unrelated to actual subtests. Tests under all upstream packages are included;
some are infrastructure tests rather than script behavior tests. Build constraints
are deliberately not filtered, so platform-specific entries remain visible.

Every entry starts `unmapped`. Do not edit this generated file to claim coverage:
the next step is a separate reviewed mapping from source Test/table cases to
reproducible Go/Rust differential cases and their results. Existing partial test
successes have deliberately not been inferred from matching function names.

With `-datakit`, the scanner also records direct `FuncsMap`/`FuncsCheckMap`
assignments/deletions and `SetNetFilter` calls under DataKit's `internal` directory,
excluding tests. Import aliases and nested function literals are supported;
dynamic keys are explicitly flagged. These are source observations, not proof of
runtime reachability, active configuration, or compatibility. Registry reads
are not overrides. Dot imports fail rather than silently disappear. Alias/dataflow
mutation, package-level initializer effects and code outside `internal` still
require manual review; build constraints are not evaluated.

The separately imported platypus interpreter's syntax tests, expanded upstream
table cases and reviewed mappings of these observations remain part of G0.
This file alone is insufficient to establish the full coverage denominator.
