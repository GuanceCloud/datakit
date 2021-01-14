
<p align="center" >
    <img width="500" src ="https://i.imgur.com/AqveQES.png" />
</p>

# Kair
> Date and Time - Golang Formatting Library

[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)  [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) [![Go Report Card](https://goreportcard.com/badge/github.com/GuilhermeCaruso/kair)](https://goreportcard.com/report/github.com/GuilhermeCaruso/kair) [![codecov](https://codecov.io/gh/GuilhermeCaruso/kair/branch/master/graph/badge.svg)](https://codecov.io/gh/GuilhermeCaruso/kair) [![Build Status](https://travis-ci.com/GuilhermeCaruso/kair.svg?branch=master)](https://travis-ci.com/GuilhermeCaruso/kair) ![GitHub](https://img.shields.io/badge/golang%20->=1.6.3-blue.svg) [![GoDoc](https://godoc.org/github.com/GuilhermeCaruso/kair?status.svg)](https://godoc.org/github.com/GuilhermeCaruso/kair) 

## Setup

To get Kair

##### > Go CLI
```sh
go get github.com/GuilhermeCaruso/kair
```
##### > Go DEP
```sh
dep ensure -add github.com/GuilhermeCaruso/kair
```
##### > Govendor
```sh
govendor fetch github.com/GuilhermeCaruso/kair
```

## Example
```go
package main

import (
	"fmt"

	k "github.com/GuilhermeCaruso/kair"
)

func main() {
	now := k.Now()

    fmt.Printf("Right now is %s \n", now.CustomFormat("dd/MM/YYYY hh:mm:ss"))

	date := k.Date(29, 05, 1980)

	fmt.Printf("The %s was a %s in %s\n",
		date.Format("L"),
		date.CustomFormat("DD"),
        date.CustomFormat("MMMM")) //The 29/05/1980 was a Thursday in May 
}

```

## Formatters
- Standard
```sh
    "LT":   10:20 AM,
    "LTS":  10:20:00 AM,
    "L":    20/05/2018,
    "l":    20/5/2018,
    "LL":   May 20, 2018,
    "ll":   May 20, 2018,
    "LLL":  May 20, 2018 10:20 AM,
    "lll":  May 20, 2018 10:20 AM,
    "LLLL": Sunday, May 20, 2018 10:20 AM,
    "llll": Sun, May 20, 2018 10:20 AM,
    "":     2018-05-20 10:20:00 +0000 UTC,
```

- Custom
```sh
    "MMMM": Long Month,
    "MMM":  Month,
    "MM":   Zero Number Month,
    "M":    Number Month,
    "YYYY": Long Year,
    "YY":   Year,
    "DD":   Long Day,
    "D":    Day,
    "dd":   Long Number Day,
    "d":    Number Day,
    "hh":   Long Hour,
    "h":    Hour,
    "mm":   Long Minute,
    "m":    Minute,
    "ss":   Long Second,
    "s":    Second
```

## Contributing
Please feel free to make suggestions, create issues, fork the repository and send pull requests!

## To do:
- [X] Implement Standard Format
- [X] Implement Custom Format
- [X] Implement Now(), Date() and DateTime() initializers
- [ ] Implement Relative Time (FromNow, StartOf ...)
- [ ] Implement CalendarTime (add, subtract, calendar)

## License

MIT License © Guilherme Caruso
