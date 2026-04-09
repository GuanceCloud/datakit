# AGENTS.md - GuanceCloud CLI Utils

This document provides guidelines for AI agents working on the GuanceCloud CLI Utils repository.

## Build, Lint, and Test Commands

### Build Commands
- `go build ./...` - Build all packages
- `go build ./<package>` - Build specific package
- `go install ./...` - Install all packages

### Lint Commands
- `make lint` - Run golangci-lint with auto-fix enabled (default)
- `LINT_FIX=false make lint` - Run golangci-lint without auto-fix
- `make vet` - Run go vet on all packages
- `make gofmt` - Check Go formatting
- `make copyright_check` - Check copyright headers
- `make copyright_check_auto_fix` - Fix copyright headers automatically

### Test Commands
- `make test` - Run all tests with coverage
- `go test ./...` - Run all tests
- `go test ./<package>` - Run tests for specific package
- `go test -v ./<package>` - Run tests with verbose output
- `go test -run TestName ./<package>` - Run specific test
- `CGO_CFLAGS=-Wno-undef-prefix go test -timeout 99999m -cover ./...` - Full test command with coverage

### Other Commands
- `make show_metrics` - Show Prometheus metrics in markdown format

## Code Style Guidelines

### General
- **Go Version**: 1.19+
- **License Header**: All files must include the MIT license header (see existing files)
- **Line Length**: Maximum 150 characters (configured in .golangci.toml)
- **Complexity**: Maximum cyclomatic complexity of 14

### Imports
- Group imports: standard library first, then third-party, then internal
- Use aliases for clarity when needed (e.g., `tu "github.com/GuanceCloud/cliutils/testutil"`)
- Use `// nolint:gosec` for intentional security exceptions (e.g., MD5 usage)
- Avoid `github.com/pkg/errors` (banned via depguard)

### Formatting
- Use `gofmt` for formatting
- Tab width: 2 spaces
- Maximum function length: 230 lines, 150 statements
- Maximum file length: Check .golangci.toml for specific limits

### Naming Conventions
- **Packages**: Lowercase, single word, descriptive
- **Functions**: MixedCaps, descriptive names
- **Variables**: camelCase, descriptive names
- **Constants**: UPPER_CASE with underscores
- **Interfaces**: MixedCaps, often ending with "er" (e.g., `Reader`)
- **Test Files**: `_test.go` suffix
- **Test Functions**: `TestXxx` format

### Error Handling
- Return errors directly: `if err != nil { return err }`
- Use `errors.New()` or `fmt.Errorf()` for new errors
- Avoid `github.com/pkg/errors` (banned)
- Handle errors immediately after function calls
- Use named returns for clarity in complex functions

### Testing
- Use `github.com/stretchr/testify/assert` for assertions
- Use `github.com/stretchr/testify/require` for required assertions
- Test functions should follow pattern: `func TestXxx(t *testing.T)`
- Use table-driven tests for multiple test cases
- Use `t.Run()` for subtests
- Test files should be in same package as code being tested
- Use `t.Helper()` for test helper functions

### Comments and Documentation
- **Package Comments**: First line should describe package purpose
- **Function Comments**: Should describe what function does, not how
- **Export Comments**: Required for exported functions, types, and variables
- **Inline Comments**: Use `//` for brief explanations
- **TODO/FIXME**: Use `// TODO:` or `// FIXME:` comments (godox checks for FIXME)

### Linter Configuration
The project uses golangci-lint with custom configuration (.golangci.toml):
- **Enabled**: Most linters enabled by default
- **Disabled**: See .golangci.toml for full list (includes: testpackage, wrapcheck, tagliatelle, paralleltest, noctx, nlreturn, gomnd, wsl, prealloc, nestif, goerr113, gochecknoglobals, exhaustivestruct, golint, interfacer, scopelint, gocognit, gocyclo, dupl, cyclop, gomoddirectives, nolintlint, revive, exhaustruct, varnamelen, nonamedreturns, forcetypeassert, gci, maintidx, containedctx, ireturn, contextcheck, errchkjson, nilnil)
- **Test Files**: Many linters disabled for test files (see exclude-rules)

### Security
- Avoid `print`, `println`, `spew.Print`, `spew.Printf`, `spew.Println`, `spew.Dump` (forbidigo)
- Use `crypto/rand` for cryptographic randomness
- Be cautious with `unsafe` package usage
- Validate all external inputs

### Performance
- Use field alignment suggestions (fieldalignment linter)
- Avoid unnecessary allocations
- Use appropriate data structures for the task

### Dependencies
- Check go.mod for existing dependencies before adding new ones
- Use `replace` directives only from allowed list (see .golangci.toml)
- Vendor directory exists - use vendored dependencies

## Project Structure
- `aggregate/` - Aggregation utilities
- `cmd/` - Command-line tools
- `dialtesting/` - Dial testing utilities
- `diskcache/` - Disk-based caching
- `filter/` - Filtering utilities
- `lineproto/` - Line protocol handling
- `logger/` - Logging utilities
- `metrics/` - Metrics collection
- `network/` - Network utilities
- `pkg/` - Internal packages
- `point/` - Data point handling
- `pprofparser/` - pprof parsing
- `system/` - System utilities
- `testutil/` - Test utilities
- `tracer/` - Tracing utilities

## Workflow
1. Always run `make lint` before committing
2. Run `make test` to ensure tests pass
3. Check copyright headers with `make copyright_check`
4. Follow existing patterns in the codebase
5. Use the testutil package for test helpers

## Notes for AI Agents
- This is a Go utilities library for GuanceCloud
- Code should be production-ready and well-tested
- Follow the existing code patterns and conventions
- Check .golangci.toml for specific linter configurations
- Use the Makefile commands for standard operations
- Test coverage is important - maintain or improve coverage