# Pipeline-GO

The pipeline-go project serves as the data pipeline processor and arbiter (SIEM) service.

## Arbiter

Arbiter is the data analysis engine of SIEM (Security Information and Event Management).

It aggregates and analyzes log and event data from different systems (such as servers, network devices, cloud services, and applications) based on the system's built-in query functions.

## Pipeline script check

The unified `pipeline-go` binary contains both Pipeline and Arbiter tools. `pipeline-go pipeline check` validates pipeline scripts by compiling them with the real pipeline runtime and, optionally, running them against a single `message` value.

```sh
go run ./cmd/pipeline-go pipeline check --script ./example.p --message '{"service":"api","status":200}' --require-key service
```

Use `--require-key` and `--expect key=value` to make extraction failures visible, especially for grok patterns that can run without matching.

The JSON output includes both the full transformed point in `output` and an AI-friendly execution summary in `result`. Use `result.extracted_fields.message` for the original `message` input, and use `result.extracted_fields`, `result.extracted_tags`, and `result.time` to quickly inspect what the script produced; use `output.fields`/`output.tags` when the full point state is needed.

The same tool embeds pipeline function docs for AI-assisted script authoring:

```sh
go run ./cmd/pipeline-go pipeline check --list-functions
go run ./cmd/pipeline-go pipeline check --search-functions json --function-lang all
go run ./cmd/pipeline-go pipeline check --function-doc grok
```

## Arbiter script check

`pipeline-go arbiter check` validates Arbiter scripts by parsing and running them with the real Arbiter runtime. It captures `stdout`, `trigger()` output, and all `dql()` calls. DQL is mocked by default so scripts can be checked without a live OpenAPI key.

```sh
go run ./cmd/pipeline-go arbiter check --script ./rule.p --dql-result-file ./sample-dql-result.json --check-dql --require-trigger --expect-status high
```

Use `--live-dql` when the DQL itself must be verified against GuanceCloud OpenAPI:

```sh
go run ./cmd/pipeline-go arbiter check --script ./rule.p --live-dql --guance-key "$GUANCE_API_KEY" --check-dql --duration 15m
```

For syntax-only checks:

```sh
go run ./cmd/pipeline-go arbiter check --script ./rule.p --parse-only
```

The same tool embeds Arbiter function docs:

```sh
go run ./cmd/pipeline-go arbiter check --search-functions dql --function-lang all
go run ./cmd/pipeline-go arbiter check --function-doc trigger
```

## AI skills

AI-facing skills are kept in `skills/`:

- `skills/pipeline-script`
- `skills/arbiter-script`
- `skills/dql`

Each skill is self-contained with `SKILL.md`, optional references, and bundled helper binaries under `bin/`. Release artifacts package these directories and inject the matching static `pipeline-go` binary for the target platform.

### Arbiter command-line tool

1. Download the executable file of Arbiter from [Github Releases](https://github.com/GuanceCloud/pipeline-go/releases)

2. Execute the help command

Commands:

```sh
$ chmod +x ./arbiter-linux-amd64
$ ./arbiter-linux-amd64 help

Arbiter command line tool

Usage:
  arbiter run -e https://openapi.guance.com -k xxxxxx script.p [flags]
  arbiter [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  fn          Arbiter built-in functions
  help        Help about any command
  run         Run aribter program

Flags:
  -h, --help   help for arbiter

Use "arbiter [command] --help" for more information about a command.
```

Command `run`:

```sh
$ ./arbiter-linux-amd64  help run
Run aribter program

Usage:
  arbiter run [flags]

Flags:
  -c, --cmd string          program passed in as string
  -d, --duration string     query time range, such as 1h, 15m, 60s (default "15m")
  -e, --guance string       GuanceCloud openapi endpoint (default "https://openapi.guance.com")
  -k, --guance-key string   GuanceCloud openapi key
  -h, --help                help for run

```


3. Create an [OpenAPI Key](https://console.guance.com/workspace/apiManage) on the platform and obtain the [Endpoint](https://docs.guance.com/open-api/#endpoint) of the corresponding site to query data of various categories.


4. Run Arbiter Script

Script example(test.p):


```txt
data = dql("M::cpu:(usage_total) BY host")

v, ok = dump_json(data, "    ")
if ok {
    printf("%v\n", v)
}


printf("%v, %v\n", 
    dql_series_get(data, "host"),
    dql_series_get(data, "usage_total"))
```

Run script:

```sh
$ ./arbiter-linux-amd64 run -e https://openapi.guance.com  -k 0SOD9gNM*****tBTbsd7x test.p
```

Result:

```json
=== stdout:
{
    "series": [
        [
            {
                "columns": {
                    "time": 1755583811966,
                    "usage_total": 7.10327456
                },
                "tags": {
                    "host": "www",
                    "name": "cpu"
                }
            }
        ],
        [
            {
                "columns": {
                    "time": 1755583803912,
                    "usage_total": 1.29961363
                },
                "tags": {
                    "host": "u22",
                    "name": "cpu"
                }
            }
        ]
    ],
    "status_code": 200
}

[["www"],["u22"]], [[7.10327456],[1.29961363]]

=== program run result:
trigger output:
null
```

5. Arbiter Function Doc

[pkg/arbiter/builtin-funcs/docs/function_doc.md](pkg/arbiter/builtin-funcs/docs/function_doc.md)
