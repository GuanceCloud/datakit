// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package build

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const itFuncName = "TestIntegrate"

var errMarks = []string{
	"FAIL",

	"PANIC",
	"Panic",
	"panic",

	"not expected",
}

var errDispute = map[string][]string{
	// "error": {
	// 	"auth_errors",
	// },
}

func IntegrationTestingDataKit() error {
	fmt.Println("Integration testing start...")
	fmt.Printf("REMOTE_HOST = %s\n", os.Getenv("REMOTE_HOST"))
	fmt.Printf("DOCKER_REMOTE_HOST = %s\n", os.Getenv("DOCKER_REMOTE_HOST"))

	start := time.Now()

	tcmd := exec.Command("go", "test", "-v", "-timeout", "1h", "-run", itFuncName, "./...") //nolint:gosec
	tcmd.Env = append(os.Environ(), []string{
		"GO111MODULE=off",
		"CGO_ENABLED=1",
		"LOGGER_PATH=nul", // disable logging
	}...)

	res, err := tcmd.CombinedOutput()
	str := string(res)

	cost := time.Since(start)

	var (
		fail     = false
		failword string
		failMsgs []string
		rtnErr   error
	)

	// check regex.
	re := regexp.MustCompile(`(?m)check measurement .* failed`)
	found := re.FindAllString(str, -1)
	if len(found) > 0 {
		fail = true
	}

	if !fail {
		// check marks.
		for _, v := range errMarks {
			if strings.Contains(str, v) {
				// check dispute.
				cntGet := strings.Count(str, v)
				cntDefinedTotal := 0
				if definedArr, ok := errDispute[v]; ok {
					for _, defined := range definedArr {
						cntDefinedGet := strings.Count(str, defined)
						cntDefinedTotal += cntDefinedGet
					}

					if cntGet == cntDefinedTotal {
						continue
					}

					failMsgs = append(failMsgs, fmt.Sprintf("cntGet = %d, cntDefinedTotal = %d\n", cntGet, cntDefinedTotal))
				}

				fail = true
				failword = v
				break
			}
		}
	}

	if fail {
		// report then.
		fmt.Println(str)
		fmt.Println()
		fmt.Println("Result contains " + failword)
		for _, v := range failMsgs {
			fmt.Println(v)
		}
		fmt.Println()

		rtnErr = fmt.Errorf("failed")
	} else {
		fmt.Println()
		fmt.Println(`  /////`)
		fmt.Println(` |     |`)
		fmt.Println(` | " " |`)
		fmt.Println(` | o o |`)
		fmt.Println(`(|  ^  |)`)
		fmt.Println(` | \_/ |`)
		fmt.Println(`  -----`)
		fmt.Println("    Integration testing all PASSED!")
		fmt.Println()
	}

	fmt.Printf("CombinedOutput, err = %v\n", err)
	fmt.Printf("Time costs: %v\n", cost)

	return rtnErr
}
