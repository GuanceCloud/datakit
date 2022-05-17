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
	"sort"
	"strconv"
	"strings"
)

func UnitTestDataKit() error {
	pkgsListCmd := exec.Command("go", "list", "./...") //nolint:gosec
	res, err := pkgsListCmd.CombinedOutput()
	if err != nil {
		return err
	}

	noTestPkgs := []string{}
	passedPkgs := map[float64][]string{}
	failedPkgs := map[string]string{}
	perc := regexp.MustCompile(`\d+\.\d+\%`)

	pkgs := strings.Split(string(res), "\n")

	coverTotal := 0.0

	for _, p := range pkgs {
		fmt.Printf("=======================\n")
		fmt.Printf("testing %s...\n", p)
		tcmd := exec.Command("go", "test", "-timeout", "1m", "-cover", p) //nolint:gosec
		tcmd.Env = append(os.Environ(), []string{
			"GO111MODULE=off",
			"CGO_ENABLED=1",
			"LOGGER_PATH=nul", // disable logging
		}...)

		res, err := tcmd.CombinedOutput()
		if err != nil {
			failedPkgs[p] = string(res)
			continue
		}

		lines := strings.Split(string(res), "\n")

		coverageLine := lines[len(lines)-2]

		// samples:
		//  ^ok  	gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/promremote	0.652s	coverage: 0.5% of statements [no tests to run]
		//  ^?   	gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/process	[no test files]
		//  ^ok  	gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/postgresql	0.715s	coverage: 52.3% of statements

		switch {
		case strings.HasPrefix(coverageLine, "?"),
			strings.Contains(coverageLine, "[no tests to run]"):
			noTestPkgs = append(noTestPkgs, p)

		case strings.HasPrefix(coverageLine, "ok"):
			coverage := perc.FindString(coverageLine)
			if len(coverage) != 0 {
				f, err := strconv.ParseFloat(coverage[0:len(coverage)-1], 64)
				if err != nil {
					fmt.Printf("[E] invalid coverage: %s: %s\n", coverage, err)
					continue
				}

				passedPkgs[f] = append(passedPkgs[f], p)
				coverTotal += f
			} else {
				fmt.Printf("[W] test ok, but no coverage: %s\n", p)
			}
		default:
			// pass
		}
	}

	fmt.Printf("============ %d package passed(avg %.2f%%) ================\n",
		len(passedPkgs), coverTotal/float64(len(passedPkgs)))
	showTopNCoveragePkgs(passedPkgs)

	fmt.Printf("============= %d package got no test ===============\n", len(noTestPkgs))
	sort.Strings(noTestPkgs)
	showNoTestPkgs(noTestPkgs)

	fmt.Printf("============= %d pakage failed ===============\n", len(failedPkgs))
	showFailedPkgs(failedPkgs)
	return nil
}

func showTopNCoveragePkgs(pkgs map[float64][]string) {
	topn := []float64{}
	for k := range pkgs {
		topn = append(topn, k)
	}

	sort.Float64s(topn)
	for _, c := range topn {
		fmt.Printf("%.2f%%\n\t%s\n", c, strings.Join(pkgs[c], "\n\t"))
	}
}

func showFailedPkgs(pkgs map[string]string) {
	for k, v := range pkgs {
		fmt.Printf("%s\n%s\n", k, v)
		fmt.Println("----------------------------")
	}
}

func showNoTestPkgs(pkgs []string) {
	for _, p := range pkgs {
		fmt.Printf("%s\n", p)
	}
}
