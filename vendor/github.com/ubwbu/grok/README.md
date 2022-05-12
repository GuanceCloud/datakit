# grok

This library is a fork of [github.com/vjeantet/grok](https://github.com/vjeantet/grok) that parses grok patterns in Go.

## Installation

Make sure you have a working Go environment.

```sh
go get github.com/ubwbu/grok
```

## Use in your project

```go
import "github.com/ubwbu/grok"
```

## Usage

### Denormalize and Compile

```go
denormalized, err := DenormalizePatternsFromMap(CopyDefalutPatterns())
if err == nil {
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

  "github.com/ubwbu/grok"
)

func main() {
  de, err := grok.DenormalizePatternsFromMap(grok.CopyDefalutPatterns())
  if err != nil {
    fmt.Print(err)
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
