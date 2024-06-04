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
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
	"go.uber.org/atomic"
)

// hugePackages is those packages that whose testing so much performance consumption.
var (
	hugePackages = map[string]bool{
		"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/db2":     true,
		"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/kafkamq": true,
		"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/mysql":   true,
		"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/oracle":  true,

		// disalbe parallel running
		"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway": true,
	}

	UTExclude, UTOnly string
	percentCoverage   *regexp.Regexp
	coverTotal        = atomic.NewFloat64(0.0)
	excludes          = map[string]bool{
		// There are multiple-main() within these modules.
		"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/proxy/bench": true,

		// root package got no test to run.
		"gitlab.jiagouyun.com/cloudcare-tools/datakit":         true,
		"gitlab.jiagouyun.com/cloudcare-tools/datakit/scripts": true,
	}

	// only run these test.
	only = map[string]bool{}

	noTestPkgs       = make([]string, 0)
	noTestPkgsLocker sync.RWMutex

	pkgCoverage = make(map[float64][]string, 0)

	npassed  = 0
	nskipped = 0
	nhuge    = 0

	failedPkgs       = make(map[string]string, 0)
	failedPkgsLocker sync.RWMutex
)

const (
	pkgPrefix                        = "gitlab.jiagouyun.com/cloudcare-tools/"
	envExcludeHugeIntegrationTesting = "UT_EXCLUDE_HUGE_INTEGRATION_TESTING"
)

func UnitTestDataKit() error {
	pkgsListCmd := exec.Command("go", "list", "./...") //nolint:gosec
	res, err := pkgsListCmd.CombinedOutput()
	if err != nil {
		return err
	}

	percentCoverage = regexp.MustCompile(`\d+\.\d+\%`)

	pkgs := strings.Split(string(res), "\n")
	utID := cliutils.XID("ut_")

	if len(UTExclude) > 0 && UTExclude != "-" {
		for _, ex := range strings.Split(UTExclude, ",") {
			fmt.Printf("package %q excluded\n", ex)
			excludes[ex] = true
		}
	}

	if len(UTOnly) > 0 && UTOnly != "-" {
		for _, item := range strings.Split(UTOnly, ",") {
			fmt.Printf("package %q selected\n", item)
			only[item] = true
		}
	}

	start := time.Now()

	lenPkgs := len(pkgs)
	for i, p := range pkgs {
		i++
		if hugePackages[p] {
			fmt.Printf("%s is HUGE package, testing it later, skip...\n", p)
			nhuge++
			continue
		}

		doWork(Job{
			UTID:    utID,
			Index:   i,
			LenPkgs: lenPkgs,
			Pkg:     p,
		})
	}

	costNormal := time.Now()
	fmt.Printf("Normal tests completed, costs = %v\n", costNormal.Sub(start))

	skipHuge := false
	if val := os.Getenv(envExcludeHugeIntegrationTesting); len(val) > 0 {
		lower := strings.ToLower(val)
		if lower == "on" {
			skipHuge = true
		}
	}

	if !skipHuge {
		nIdx := 0
		lenHugePkgs := len(hugePackages)
		for pkg := range hugePackages {
			if excludes[pkg] {
				fmt.Printf("Skip huge test %q\n", pkg)
				continue
			}

			nIdx++
			fmt.Printf("run huge test %q\n", pkg)
			doWork(Job{
				UTID:    utID,
				Index:   nIdx,
				LenPkgs: lenHugePkgs,
				Pkg:     pkg,
			})
		}

		costHuge := time.Now()
		fmt.Printf("Huge tests completed, costs = %v, total = %v\n", costHuge.Sub(costNormal), costHuge.Sub(start))
	} else {
		fmt.Printf("All huge tests skipped\n")
	}

	mr := &testutils.ModuleResult{
		Name:      "datakit-ut",
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		GoVersion: runtime.Version(),
		Branch:    git.Branch,
		TestID:    utID,
		Coverage:  coverTotal.Load() / float64(npassed),
		Message:   fmt.Sprintf("done, total cost: %s", time.Since(start)),
	}

	if err := testutils.Flush(mr); err != nil {
		fmt.Printf("[E] flush metric failed: %s\n", err)
	}

	fmt.Printf("============ %d package passed(avg %.2f%%) ================\n",
		npassed, coverTotal.Load()/float64(npassed))
	showTopNCoveragePkgs(pkgCoverage)

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

func addNoTestPkgs(pkg string) {
	noTestPkgsLocker.Lock()
	defer noTestPkgsLocker.Unlock()
	noTestPkgs = append(noTestPkgs, pkg)
}

func addFailedPkgs(pkg, detail string) {
	failedPkgsLocker.Lock()
	defer failedPkgsLocker.Unlock()
	failedPkgs[pkg] = detail
}

type Job struct {
	UTID    string // unit test ID.
	Index   int
	LenPkgs int
	Pkg     string
}

func doWork(j Job) {
	start := time.Now()
	fmt.Printf("=======================\n")
	fmt.Printf("[%s][%02d:%02d:%02d][passed:%d/notest:%d/skipped:%d/huge:%d/failed:%d] testing(%03d/%03d) %s...\n",
		j.UTID,
		start.Hour(),
		start.Minute(),
		start.Second(),
		npassed,
		len(noTestPkgs),
		nskipped,
		nhuge,
		len(failedPkgs),
		j.Index,
		j.LenPkgs,
		j.Pkg)

	if excludes[j.Pkg] {
		fmt.Printf("[%s] package(%03d/%03d) %s excluded...\n", j.UTID, j.Index, j.LenPkgs, j.Pkg)
		nskipped++
		return
	}

	if len(only) > 0 && !only[j.Pkg] {
		fmt.Printf("[%s] package(%03d/%03d) %s not selected, selected: %+#v\n", j.UTID, j.Index, j.LenPkgs, j.Pkg, only)
		nskipped++
		return
	}

	mr := &testutils.ModuleResult{
		Name:      strings.TrimPrefix(j.Pkg, pkgPrefix), // remove prefix for human readable.
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		GoVersion: runtime.Version(),
		Branch:    git.Branch,
		TestID:    j.UTID,
	}

	tcmd := exec.Command("go", "test", "-count=1", "-timeout", "1h", "-cover", j.Pkg) //nolint:gosec
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
		if (!strings.Contains(mr.Message, "no Go files in")) || strings.Contains(mr.Message, "FAIL") {
			addFailedPkgs(j.Pkg, string(res))

			showFailedPkgs(failedPkgs) // show failed packages ASAP.

			mr.Status = testutils.TestFailed
			mr.FailedMessage = err.Error()
			if err := testutils.Flush(mr); err != nil {
				fmt.Printf("[E] flush metric failed: %s\n", err)
			}

			l.Errorf("package %s failed: %s", j.Pkg, string(res))
		} else {
			mr.Status = testutils.TestSkipped
		}
	}

	lines := strings.Split(string(res), "\n")
	coverageLine := lines[len(lines)-2]

	//nolint
	// go test output example:
	//  ^ok  	gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/promremote	0.652s	coverage: 0.5% of statements [no tests to run]
	//  ^?   	gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/process	[no test files]
	//  ^ok  	gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/postgresql	0.715s	coverage: 52.3% of statements
	// ok  	gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/point	(cached)	coverage: 0.0% of statements [no tests to run]
	// ok  	gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/conntrack	(cached)	coverage: [no statements] [no tests to run]
	// package gitlab.jiagouyun.com/cloudcare-tools/datakit: no Go files in /root/go/src/gitlab.jiagouyun.com/cloudcare-tools/datakit

	fmt.Println(coverageLine)

	switch {
	case strings.HasPrefix(coverageLine, "?"),
		strings.Contains(coverageLine, "[no tests to run]"):
		addNoTestPkgs(j.Pkg)
		mr.NoTest = true

	case strings.HasPrefix(coverageLine, "ok"):
		mr.Status = testutils.TestPassed
		npassed++

		coverage := percentCoverage.FindString(coverageLine)
		if len(coverage) != 0 {
			f, err := strconv.ParseFloat(coverage[0:len(coverage)-1], 64)
			if err != nil {
				fmt.Printf("[E] invalid coverage %q: %s: %s\n", j.Pkg, coverage, err)
				return
			}

			pkgCoverage[f] = append(pkgCoverage[f], j.Pkg)
			coverTotal.Add(f)
			mr.Coverage = f
		} else {
			fmt.Printf("[W] test ok, but no coverage: %q\n", j.Pkg)
		}

	default: // pass
		fmt.Printf("[W] unknown coverage line in package %q: %s\n", j.Pkg, coverageLine)
	}

	if err := testutils.Flush(mr); err != nil {
		fmt.Printf("[E] flush metric failed: %s\n", err)
	}
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
	failedPkgsLocker.Lock()
	for k, v := range pkgs {
		fmt.Printf("%s\n%s\n", k, v)
		fmt.Println("----------------------------")
	}
	failedPkgsLocker.Unlock()
}

func showNoTestPkgs(pkgs []string) {
	for _, p := range pkgs {
		fmt.Printf("%s\n", p)
	}
}
