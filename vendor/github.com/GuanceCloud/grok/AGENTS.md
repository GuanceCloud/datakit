# Repository Guidelines

## Project Structure & Module Organization
This repository is a small Go module named `github.com/GuanceCloud/grok` targeting Go `1.18`. Core library code lives at the repository root in files such as `grok.go`, `pattern.go`, and `tree.go`, all under package `grok`. Tests also live at the root in `*_test.go` files. Built-in pattern definitions are stored as plain files under `patterns/`; add new pattern sets there, using descriptive filenames such as `patterns/nginx` or `patterns/custom-app`.

## Build, Test, and Development Commands
Use the `Makefile` targets where possible:

- `make fmt`: runs `gofmt -w -s` on all Go files.
- `make lint`: runs `golangci-lint run --fix --allow-parallel-runners`.
- `make test`: runs the full test suite with `go test -v ./...`.
- `make test-cov`: generates `coverage.txt` with `go test -cover -coverprofile=coverage.txt ./...`.

For targeted iteration, `go test -run TestParse ./...` is the quickest way to rerun a single test.

## Coding Style & Naming Conventions
Follow standard Go formatting and let `gofmt` own whitespace, indentation, and import ordering. Keep exported API names in `CamelCase` and internal helpers in lower camel case. Use short, descriptive names that match the existing parser vocabulary: `CompilePattern`, `LoadPatternsFromPath`, `DenormalizePattern`. Keep package scope flat unless a new subpackage is clearly justified.

## Testing Guidelines
Write tests next to the code they cover in `*_test.go` files. Prefer table-driven tests for parser behavior and edge cases, following the existing style in `grok_test.go` and `tree_test.go`. `stretchr/testify/assert` is already used and can be reused for value comparisons. Run `make test` before opening a PR; run `make test-cov` when changing parsing logic, matching behavior, or pattern expansion.

## Commit & Pull Request Guidelines
Recent commits use short, imperative summaries such as `add function CompilePattern2`, `modify test case`, and `update go.mod`. Keep commit subjects concise, focused, and scoped to one change. Pull requests should explain the behavior change, note any added or updated patterns, and include test coverage details. If the change affects parsing performance or memory allocation, include benchmark or measurement notes.
