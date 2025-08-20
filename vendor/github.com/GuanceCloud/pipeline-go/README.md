# Pipeline-GO

The pipeline-go project serves as the data pipeline processor and arbiter (SIEM) service.

## Arbiter

Arbiter is the data analysis engine of SIEM (Security Information and Event Management).

It aggregates and analyzes log and event data from different systems (such as servers, network devices, cloud services, and applications) based on the system's built-in query functions.


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
