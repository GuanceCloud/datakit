# AI Agent Instructions (AGENT.md)

This document contains instructions for AI agents, coding assistants, and LLMs interacting with this repository(DataKit). **Read this entirely before suggesting or modifying any code.**

We'll update this doc on the fly, before any operation, you should check if there any update of this doc, if any update, you should re-read and apply the updates.

## Agent Role

You are a Senior Software Engineer specializing in [Go/C/Python/Linux/Windows]. You write clean, modular, and highly optimized code. You prioritize security, readability, and performance. You never leave `TODO` comments unless explicitly asked.

## Project Overview

This project is a [collect agent on system/app observerbility with metrics/logs/tracing].

## Tech Stack

- **Language:** [mainly Golang 1.19+/C99 (or C11) Standard for eBPF]
- **Testing:** [github.com/stretchr/testify]
    - We are planing to apply BDD in the future.

## Directory Structure

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
    - `plugins`: All collectors are implemented here. [Some of these collectors](internal/plugins/external) are using CGO, [most of them](internal/plugins/inputs) are build statically. Every collector are imported into [all](internal/plugins/all/) with define a `init()`
- `k8s-profilers`: **Deprecated** Profiler builder for Java/Golang/Python
- `scripts`: Scripts & tools used among Datakit daily development(spell-checker/linters and so on)
- `templates`: Various templates(yaml/markdown) for Helm/k8s yaml/shell installer
- `vender`: All packages are vendered in repository
- `export.sh`: We run this to export all docs to other repository
- `Makefile`: Build and release entry
- `gitlab-ci.yml`: CI for GitLab

## Coding Standards (Strict Go Idioms)

- **Comments:** Explain *why* complex logic exists, not *what* the code is doing.
- **Concurrency:** Prefer communicating over channels rather than sharing memory with mutexes, unless performance dictates otherwise. Do not leak Goroutines.
- **Dependencies:** Use Go modules (`go.mod`). Do not introduce new third-party packages unless absolutely necessary.
- **Error Handling:** DO NOT use `panic` for control flow. ALWAYS return errors as the last return value and handle them explicitly with `if err != nil { ... }`. Wrap errors using `fmt.Errorf("failed to do X: %w", err)`.
- **Formatting:** All code MUST comply with `gofmt` and `goimports`. 
- **Linting:** Code MUST pass `golangci-lint` with default settings. No dead code, no unused variables.
- **Logging:** We prefer use a specific logger for single module

    ```go
	import "github.com/GuanceCloud/cliutils/logger"

    // define moulde global logger
    //
    // this l default log to stdout
	l = logger.DefaultSLogger("module-name")

    // after internal/config load, call Init() to apply logging settings(level/log path and so on)
    func Init() {
        l = logger.SLogger("module-name")
    }

    // log some message
    l.Infof("this is a info: %s", "some-info")
    l.Debugf("this is a debug: %s", "some-debug")
    ```

### Coding Standards for C(Strict C Rules)

- **Formatting:** Code MUST be formatted using `clang-format` based on the `.clang-format` file in the repository root.
- **Linting & Analysis:** Code MUST pass `clang-tidy` and `cppcheck` with zero warnings.
- **Compiler Flags:** Code must compile cleanly with `-Wall -Wextra -Werror -pedantic`. Do not write code that generates compiler warnings.
- **Memory Management:** 
  - Every `malloc`, `calloc`, or `realloc` MUST have a matching `free()`.
  - ALWAYS check if a pointer is `NULL` before dereferencing it.
  - ALWAYS check the return value of memory allocation functions (if `malloc` returns `NULL`, handle the out-of-memory error).
- **Safety First:** DO NOT use unsafe standard functions. Use `strncpy` instead of `strcpy`, `snprintf` instead of `sprintf`, and `strncat` instead of `strcat`.

## Testing Requirements
- Write tests using the standard `testing` package (`go test`).
- Use table-driven tests (slice of structs) for testing multiple scenarios.
- Put test files in the same package as the code being tested (e.g., `user_test.go` next to `user.go`).
- For unit test, reach %80+ coverage.

### Testing Requirements for C
- Tests are written using [e.g., Unity, CMocka, or CTest].
- All new memory allocation logic must be verifiable by `Valgrind` to ensure zero memory leaks.

## Anti-Patterns (DO NOT DO THESE)
- **NEVER** ignore errors (e.g., `_, err := doSomething()`). Always handle or return the error.
- **NEVER** use global variables for configuration; use dependency injection (struct fields).

### Anti-Patterns for C
- **NEVER** use variable-length arrays (VLAs).
- **NEVER** cast a pointer without a clear and necessary reason. Avoid `void*` magic unless implementing generic data structures.
- **NEVER** hardcode buffer sizes with magic numbers. Use `#define` or `const size_t`.

## Gitlab 

These global settings are applied for all Gitlab related topics.

- Repository: `https://gitlab.jiagouyun.com/cloudcare-tools/datakit`
- The {{.Version}} comes from *gitlab-ci.yml* on variable `CI_VERSION`, you can grep on `cat gitlab-ci.yml | grep -w "CI_VERSION"`
- Read Gitlab milestone `https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/milestones`, select the milestone same as {{.Version}}
- Read the Gitlab access token from ENV `GITLAB_ACCESS_TOKEN`

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
