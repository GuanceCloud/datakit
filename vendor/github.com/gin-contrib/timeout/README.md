# Timeout

[![Run Tests](https://github.com/gin-contrib/timeout/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/gin-contrib/timeout/actions/workflows/go.yml)
[![codecov](https://codecov.io/gh/gin-contrib/timeout/branch/master/graph/badge.svg)](https://codecov.io/gh/gin-contrib/timeout)
[![Go Report Card](https://goreportcard.com/badge/github.com/gin-contrib/timeout)](https://goreportcard.com/report/github.com/gin-contrib/timeout)
[![GoDoc](https://godoc.org/github.com/gin-contrib/timeout?status.svg)](https://pkg.go.dev/github.com/gin-contrib/timeout?tab=doc)
[![Join the chat at https://gitter.im/gin-gonic/gin](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/gin-gonic/gin)

Timeout wraps a handler and aborts the process of the handler if the timeout is reached.

## Example

```go
package main

import (
  "log"
  "net/http"
  "time"

  "github.com/gin-contrib/timeout"
  "github.com/gin-gonic/gin"
)

func emptySuccessResponse(c *gin.Context) {
  time.Sleep(200 * time.Microsecond)
  c.String(http.StatusOK, "")
}

func main() {
  r := gin.New()

  r.GET("/", timeout.New(
    timeout.WithTimeout(100*time.Microsecond),
    timeout.WithHandler(emptySuccessResponse),
  ))

  // Listen and Server in 0.0.0.0:8080
  if err := r.Run(":8080"); err != nil {
    log.Fatal(err)
  }
}

```

### custom error response

Add new error response func:

```go
func testResponse(c *gin.Context) {
  c.String(http.StatusRequestTimeout, "test response")
}
```

Add `WithResponse` option.

```go
  r.GET("/", timeout.New(
    timeout.WithTimeout(100*time.Microsecond),
    timeout.WithHandler(emptySuccessResponse),
    timeout.WithResponse(testResponse),
  ))
```
