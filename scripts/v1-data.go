// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package main

import (
	"flag"
	"os"

	"github.com/GuanceCloud/cliutils"
)

var (
	flagLen   = flag.Int64("len", 32, "generated data length(kb)")
	flagCount = flag.Int("count", 1, "generated data count")
	flagFile  = flag.String("output", "v1.data", "data output to file")
)

func genLargeLog() {
	fd, err := os.OpenFile(*flagFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err.Error())
	}

	for i := 0; i < *flagCount; i++ {
		n := *flagLen * 1024

		if i > 0 { // append \n to previous data
			if _, err := fd.WriteString("\n"); err != nil {
				panic(err.Error())
			}
		}

		for {
			if n/1024 > 0 { // each time generate 1kb data
				if _, err := fd.WriteString(cliutils.CreateRandomString(1024)); err != nil {
					panic(err.Error())
				}
				n -= 1024
			} else {
				if _, err := fd.WriteString(cliutils.CreateRandomString(int(n % 1024))); err != nil {
					panic(err.Error())
				}

				break
			}
		}
	}
}

// nolint: typecheck
func main() {
	flag.Parse()
	genLargeLog()
}
