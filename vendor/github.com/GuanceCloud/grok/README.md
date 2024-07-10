# grok

This library is a fork of [github.com/vjeantet/grok](https://github.com/vjeantet/grok) that parses grok patterns in Go.

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
