# AGENTS_COMMON.md - Common Coding Agent Instructions for DataKit

This document provides essential information for AI coding agents working on the DataKit repository. It contains build commands, test commands, linting rules, and coding conventions.

> NOTE: **Read this entirely before suggesting or modifying any code.**

## Agent Role

You are a Senior Software Engineer specializing in [Go/C/Python/Linux/Windows]. You write clean, modular, and highly optimized code. You prioritize security, readability, and performance. You never leave `TODO` comments unless explicitly asked.

## Project Overview

DataKit is an open-source data collection agent for system/application observability (metrics, logs, tracing). It's primarily written in Go 1.19+ with C code for eBPF components.

## Build Commands

### Main Build Targets
- `make` or `make local` - Build for local platform
- `make testing` - Build for all platforms (Linux/Mac/Windows)
- `make production` - Build production release
- `make clean` - Clean build artifacts

### Architecture & Platform Support
- Architectures: amd64, arm64, 386
- Platforms: Linux, macOS, Windows
- Binaries: `datakit`, `datakit-ebpf`, `logfwd`, `flameshot`, `installer`, `upgrader`

### Prerequisites
- Go 1.19+
- Build tools: make, gcc, gcc-multilib, tree, goyacc
- eBPF: clang 10.0+, llvm 10.0+, kernel headers
- Linting: golangci-lint v1.46.2

## Test Commands

### Unit Testing
- `make ut` - Run unit tests with configurable parameters:
  - `UT_EXCLUDE` - Exclude specific tests (regex pattern)
  - `UT_ONLY` - Run only specific tests (regex pattern)
  - `UT_PARALLEL` - Parallel execution level (integer)
  - `DATAWAY_URL` - Test dataway endpoint
- `go test ./...` - Standard Go test (via `test.sh` script)
- `make coverage` - Generate test coverage report

### Running Single Tests
```bash
# Run specific test file
go test ./internal/plugins/inputs/cpu

# Run specific test function
go test -v ./internal/plugins/inputs/cpu -run TestCPU

# Run with specific test pattern
make ut UT_ONLY="TestCPU"

# Run tests excluding certain patterns
make ut UT_EXCLUDE="integration|e2e"
```

### Test Configuration
- Target: 80%+ test coverage
- Framework: Standard `testing` package with `github.com/stretchr/testify`
- Pattern: Table-driven tests encouraged
- Location: Test files (`*_test.go`) in same package as code

## Linting & Formatting Commands

### Code Quality Checks
- `make lint` - Run all linting (code, docs, sample configs)
- `make code_lint` - Code-only linting with golangci-lint
- `make copyright_check` - Check copyright headers
- `make disable_funcs` - Check for prohibited functions

### Formatting
- `make gofmt` - Format Go code with gofmt
- `make copyright_check_auto_fix` - Auto-fix copyright headers

### Linter Configuration
- Uses `.golangci.toml` for golangci-lint configuration
- Custom script: `scripts/disable-funcs/conf.toml` prohibits `fmt.Println`, `fmt.Printf`, `log.Println`, `log.Printf`
- Skipped directories: vendor/, .git/, generated files, some driver packages

## Code Style Guidelines

### Go Conventions
1. **Error Handling**: Always handle errors explicitly
   ```go
   // Good
   if err := doSomething(); err != nil {
       return fmt.Errorf("context: %w", err)
   }
   
   // Bad - ignored error
   _, _ = doSomething()
   ```

2. **Logging**: Use module-specific logger
   ```go
   import "github.com/GuanceCloud/cliutils/logger"
   
   l = logger.DefaultSLogger("module-name")
   l.Infof("message: %s", value)
   l.Debugf("debug: %s", value)
   ```

3. **Formatting**: Must comply with `gofmt` and `goimports`

4. **Concurrency**: Prefer channels over mutexes, no goroutine leaks

5. **Comments**: Explain "why" not "what"

6. **Dependencies**: Minimize third-party packages

### C Code Standards (eBPF)
- Format with `clang-format` using `.clang-format` in repo root
- Zero warnings with `clang-tidy` and `cppcheck`
- Compiler flags: `-Wall -Wextra -Werror -pedantic`
- Memory safety: Check `NULL` pointers, match `malloc/free`
- No unsafe functions: Use `strncpy`, `snprintf`, `strncat`
- No variable-length arrays (VLAs)
- No magic numbers for buffer sizes

### Import Order (Go)
1. Standard library
2. Third-party packages
3. Local/internal packages

### Naming Conventions
- Use camelCase for variables and functions
- Use PascalCase for exported identifiers
- Use UPPER_SNAKE_CASE for constants
- Use descriptive names, avoid abbreviations

## Anti-Patterns to Avoid

### General
- Never ignore errors (`_, err := doSomething()`)
- No global variables for configuration
- No `panic` for control flow
- No `TODO` comments unless explicitly asked
- No unused variables or imports
- No dead code

### C-Specific
- No variable-length arrays (VLAs)
- No unsafe pointer casts without clear reason
- No magic numbers for buffer sizes
- No `void*` magic unless implementing generic data structures

## GitLab CI Integration

These global settings are applied for all Gitlab related topics.

- Repository: `https://gitlab.jiagouyun.com/cloudcare-tools/datakit`
- The {{.Version}} comes from *gitlab-ci.yml* on variable `CI_VERSION`, you can grep on `cat gitlab-ci.yml | grep -w "CI_VERSION"`
- Read Gitlab milestone `https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/milestones`, select the milestone same as {{.Version}}
- Read the Gitlab access token from ENV `GITLAB_ACCESS_TOKEN`

### Pipeline Commands

- `ci_lint_and_ut` - Main CI job (linting + unit tests)
- Configuration in `gitlab-ci.yml`
- Uses `GITLAB_ACCESS_TOKEN` from environment

### For Changelog

Export changelog with the following requirements:

- Read history change log in *internal/export/doc/zh/changelog-{{.Year}}.md*
- Read Gitlab milestone
- For all closed/merged issue, add new changelog to the top of changelog-{{.Year}}.md above. do not use emoji
- Also add english changelog to *internal/export/doc/en/changelog-{{.Year}}.md*

### Local Code Review

We'll ask your to review local code, within the review result, if there any issue found, please 

- Fire a issue to Gitlab, and assign to specific owner.
- No milestone need to assign
- You can list all available labels within Gitlab repository, and add specific label for the new issue(such as bug/feat-request and so on)

### Merge Request Review

Review milestone code with the following requirements:

- Read Gitlab milestone `https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/milestones`, select the milestone same as {{.Version}}
- For all issue:
    - Read the issue and it's related merge request(closed or open), one issue may have one or more merge request(ignore the closed merge request)
    - Review the code for the issue, give advice to code lines if any.
    - Add or update a comment(conclusion) for the issue's merge request

### Version Management

- Version from `git describe --always --tags`
- Git info embedded in `internal/git/git.go`
- DCA version separate from main version

## Directory Structure Reference

- `charts/`: Helm package build for DataKit
- `cmd`: Mainly binary `main.go` entry.
    - `awslambda`: For DataKit running within AWS lambda
    - `datakit`: For DataKit running on host and kubernetes
    - `dca`: For DataKit web manager backend(talk with web on websocket)
    - `flameshot`: Profiling collector for various language(such as Java/Go/Python/...)
    - `installer`: DataKit host(on Linux/MacOS/Windows) binary installer
    - `make`: DataKit dist builder and release tools writting in Go
    - `upgrader`: DataKit upgradder on host, we need to upgrader daemon for DataKit
- `dca`: DataKit web manager frontend(DataKit Controller Application)
- `dockerfiles`: Various Dockerfiles for binaries on multiple platforms and brands
- `internal`: All internal packages, some of them are:
    - `cmds`: All DataKit command line are defined, all flags are defined within [parse_flags.go](internal/cmds/parse_flags.go)
    - `config`: Main configure(`datakit.conf`) loader
    - `datakit`: Global variables used among DataKit
    - `export`: Every DataKit release will build huge number of document, we render and export [these documents](internal/export/doc) for multiple brands.
    - `httpapi`: DataKit exports multiple HTTP APIs
    - `httpcli`: DataKit itself is a HTTP client that upload metrics/logs/tracing to DataWay
    - `io`: All collector are feeding their collected data(metrics/logs/tracing) to `io` module, within `io` module, we will run some filters(to drop unneed data) and pipelines(such as split raw log to structured log), then build these data(in protobuf) and upload them to DataWay via HTTP
    - `plugins`: All collectors are implemented here. [Some of these collectors](internal/plugins/externals) are using CGO, [most of them](internal/plugins/inputs) are build statically. Every collector are imported into [all](internal/plugins/inputs/all/) with define a `init()`
- `k8s-profilers`: **Deprecated** Profiler builder for Java/Golang/Python
- `scripts`: Scripts & tools used among Datakit daily development(spell-checker/linters and so on)
- `templates`: Various templates(yaml/markdown) for Helm/k8s yaml/shell installer
- `vender`: All packages are vendered in repository
- `export.sh`: We run this to export all docs to other repository
- `Makefile`: Build and release entry
- `gitlab-ci.yml`: CI for GitLab

## Quick Reference Commands

### Development Workflow

```bash
# 1. Build
make

# 3. Lint code
make lint

# 4. Format code
make gofmt

# 5. Check for prohibited functions
make disable_funcs
```

### Common Tasks

```bash
# Build specific binary
make local

# Run single test file
go test ./internal/plugins/inputs/cpu

# Lint only Go code
make code_lint

# Auto-fix copyright headers
make copyright_check_auto_fix
```

## Notes for AI Agents

1. **Always read AGENTS.md first** - Comprehensive agent instructions exist at `AGENTS.md`
2. **Check for updates** - This document may be updated; re-read if needed
3. **Follow existing patterns** - Look at similar code for style guidance
4. **Run lint/tests** - Verify changes pass all checks
5. **No TODOs** - Unless explicitly requested by user

## File Location Conventions

- Go files: `*.go`
- Test files: `*_test.go`
- C files: `*.c`, `*.h`
- Configuration: `*.toml`, `*.conf`
- Documentation: `*.md`

This document should be referenced by all AI agents working on the DataKit repository to ensure consistent code quality and adherence to project standards.

