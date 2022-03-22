# grok

This library is a fork of [github.com/vjeantet/grok](https://github.com/vjeantet/grok) that parses grok patterns in Go.

## Usage

### Denormalize and Compile

```go
denormalized, errs := DenormalizePatternsFromMap(CopyDefalutPatterns())
if len(errs) == 0 {
  g, err := CompilePattern("%{DAY:day}", denormalized)
  if err == nil {
    ret, _ := g.Run("Tue qds")
  }
}
```

## Example

```go
package main

import (
  "fmt"

  "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/grok"
)

func main() {
  de, errs := grok.DenormalizePatternsFromMap(grok.CopyDefalutPatterns())
  if len(errs) != 0 {
    fmt.Print(errs)
    return
  }
  g, err := grok.CompilePattern("%{COMMONAPACHELOG}", de)
  if err != nil {
    fmt.Print(err)
  }
  ret, err := g.Run(`127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`)
  if err != nil {
    fmt.Print(err)
  }
  for k, v := range ret {
    fmt.Printf("%+15s: %s\n", k, v)
  }
}

```

output:

```txt
      timestamp: 23/Apr/2014:22:58:32 +0200
           verb: GET
        request: /index.php
    httpversion: 1.1
          bytes: 207
       response: 404
               : 207
       clientip: 127.0.0.1
          ident: -
           auth: -
     rawrequest: 
```
