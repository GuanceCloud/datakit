# grok

This library is a fork of [github.com/vjeantet/grok](https://github.com/vjeantet/grok) that parses grok patterns in Go.

## Project Notes

- Performance results: [docs/perf/BENCHMARKS.md](./docs/perf/BENCHMARKS.md)
- Performance conclusions, exact-literal prefilter notes, and the global dispatcher draft: [docs/perf/PERFORMANCE_NOTES.md](./docs/perf/PERFORMANCE_NOTES.md)
- Project layout and runtime layering: [docs/architecture/PROJECT_LAYOUT.md](./docs/architecture/PROJECT_LAYOUT.md)
- Local handoff context for ongoing work: [docs/context/CURRENT_CONTEXT.md](./docs/context/CURRENT_CONTEXT.md)

## Usage

### Denormalize and Compile

```go
denormalized, errs := DenormalizePatternsFromMap(CopyDefalutPatterns())
if len(errs) == 0 {
  g, err := CompilePattern("%{DAY:day}", grok.PatternStorage{denormalized})
  if err == nil {
    ret, _ := g.Run("Tue qds", false)
  }
}
```

## Example

```go
package main

import (
  "fmt"

  "github.com/GuanceCloud/grok"
)

func main() {
  de, errs := grok.DenormalizePatternsFromMap(grok.CopyDefalutPatterns())
  if len(errs) != 0 {
    fmt.Print(errs)
    return
  }
  g, err := grok.CompilePattern("%{COMMONAPACHELOG}", grok.PatternStorage{de})
  if err != nil {
    fmt.Print(err)
  }
  ret, err := g.Run(`127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`, true)
  if err != nil {
    fmt.Print(err)
  }
  for k, name := range g.MatchNames() {
    fmt.Printf("%+15s: %s\n", name, ret[k])
  }
}

```

### Reusable Entry Points

For callers such as `pipeline-go` that want to reuse result buffers, use
`RunTo` or `RunWithTypeInfoTo`:

```go
buf := make([]any, 0, g.MatchCount())
ret, err := g.RunWithTypeInfoTo(line, true, buf)
```

For multi-pattern pipelines, compile a `MatcherSet` and reuse one buffer sized
from the largest matcher:

```go
buf := make([]string, 0, set.MatchCount())
id, ret, err := set.RunFirstTo(line, true, buf)
```

output:

```txt
     clientip: 127.0.0.1
        ident: -
         auth: -
    timestamp: 23/Apr/2014:22:58:32 +0200
         verb: GET
      request: /index.php
  httpversion: 1.1
   rawrequest:
     response: 404
        bytes: 207
```
