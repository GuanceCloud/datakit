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
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
)

const (
	pkgPrefix = "gitlab.jiagouyun.com/cloudcare-tools/"
)

var UTExclude string

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
	utID := cliutils.XID("ut_")

	coverTotal := 0.0
	excludes := map[string]bool{}

	if len(UTExclude) > 0 {
		for _, ex := range strings.Split(UTExclude, ",") {
			excludes[ex] = true
		}
	}

	start := time.Now()

	for i, p := range pkgs {
		fmt.Printf("=======================\n")
		fmt.Printf("[%s] testing(%03d/%d) %s...\n", utID, i, len(pkgs), p)

		if excludes[p] {
			fmt.Printf("%s excluded\n", p)
			continue
		}

		mr := &testutils.ModuleResult{
			// remove prefix for human readable
			Name:      strings.TrimPrefix(p, pkgPrefix),
			OS:        runtime.GOOS,
			Arch:      runtime.GOARCH,
			GoVersion: runtime.Version(),
			Branch:    git.Branch,
			TestID:    utID,
		}

		start := time.Now()
		tcmd := exec.Command("go", "test", "-timeout", "1h", "-cover", p) //nolint:gosec
		tcmd.Env = append(os.Environ(), []string{
			"GO111MODULE=off",
			"CGO_ENABLED=1",
			"LOGGER_PATH=nul", // disable logging
		}...)

		res, err := tcmd.CombinedOutput()
		mr.Cost = time.Since(start)
		if len(res) > 0 {
			mr.Message = string(res)
		}

		if err != nil {
			if !strings.Contains(mr.Message, "no Go files in") {
				failedPkgs[p] = string(res)

				mr.Status = testutils.TestFailed
				mr.FailedMessage = err.Error()
				if err := testutils.Flush(mr); err != nil {
					fmt.Printf("[E] flush metric failed: %s\n", err)
				}
			} else {
				mr.Status = testutils.TestSkipped
			}
		}

		lines := strings.Split(string(res), "\n")

		coverageLine := lines[len(lines)-2]

		// go test output example:
		//  ^ok  	gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/promremote	0.652s	coverage: 0.5% of statements [no tests to run]
		//  ^?   	gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/process	[no test files]
		//  ^ok  	gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/postgresql	0.715s	coverage: 52.3% of statements

		switch {
		case strings.HasPrefix(coverageLine, "?"),
			strings.Contains(coverageLine, "[no tests to run]"):
			noTestPkgs = append(noTestPkgs, p)
			mr.NoTest = true

		case strings.HasPrefix(coverageLine, "ok"):
			mr.Status = testutils.TestPassed

			coverage := perc.FindString(coverageLine)
			if len(coverage) != 0 {
				f, err := strconv.ParseFloat(coverage[0:len(coverage)-1], 64)
				if err != nil {
					fmt.Printf("[E] invalid coverage: %s: %s\n", coverage, err)
					continue
				}

				passedPkgs[f] = append(passedPkgs[f], p)
				coverTotal += f
				mr.Coverage = f
			} else {
				fmt.Printf("[W] test ok, but no coverage: %s\n", p)
			}
		default: // pass
			fmt.Printf("[W] unknown coverage line: %s\n", coverageLine)
		}

		if err := testutils.Flush(mr); err != nil {
			fmt.Printf("[E] flush metric failed: %s\n", err)
		}
	}

	mr := &testutils.ModuleResult{
		// remove prefix for human readable
		Name:      "datakit-ut",
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		GoVersion: runtime.Version(),
		Branch:    git.Branch,
		TestID:    utID,
		Coverage:  coverTotal / float64(len(passedPkgs)),
		Message:   fmt.Sprintf("done, total cost: %s", time.Since(start)),
	}

	if err := testutils.Flush(mr); err != nil {
		fmt.Printf("[E] flush metric failed: %s\n", err)
	}

	fmt.Printf("============ %d package passed(avg %.2f%%) ================\n",
		len(passedPkgs), coverTotal/float64(len(passedPkgs)))
	showTopNCoveragePkgs(passedPkgs)

	fmt.Printf("============= %d package got no test ===============\n", len(noTestPkgs))
	sort.Strings(noTestPkgs)
	showNoTestPkgs(noTestPkgs)

	fmt.Printf("============= %d pakage failed ===============\n", len(failedPkgs))
	showFailedPkgs(failedPkgs)
	if len(failedPkgs) > 0 {
		return fmt.Errorf("%d package failed", len(failedPkgs))
	}

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
