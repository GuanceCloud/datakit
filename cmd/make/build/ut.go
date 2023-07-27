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
var hugePackages = map[string]bool{
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/kafkamq": true,
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/mysql":   true,
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/oracle":  true,
}

func UnitTestDataKit() error {
	pkgsListCmd := exec.Command("go", "list", "./...") //nolint:gosec
	res, err := pkgsListCmd.CombinedOutput()
	if err != nil {
		return err
	}

	percentCoverage = regexp.MustCompile(`\d+\.\d+\%`)

	pkgs := strings.Split(string(res), "\n")
	utID := cliutils.XID("ut_")

	if len(UTExclude) > 0 {
		for _, ex := range strings.Split(UTExclude, ",") {
			excludes[ex] = true
		}
	}

	numWorkers := runtime.NumCPU()
	for i := 1; i <= numWorkers; i++ {
		go work(i)
	}

	start := time.Now()

	lenPkgs := len(pkgs)
	addPassedPkgsJobs = make(chan passedPkgsOpt, lenPkgs)
	lenHugePkgs := len(hugePackages)
	lenPkgs -= lenHugePkgs
	go workAddPassedPkgs()

	wg.Add(lenPkgs)
	for i, p := range pkgs {
		if hugePackages[p] {
			fmt.Printf("%s is HUGE package, testing it later, skip...\n", p)
			continue
		}
		jobs <- Job{
			UTID:    utID,
			Index:   i,
			LenPkgs: lenPkgs,
			Pkg:     p,
			wg:      &wg,
		}
	}
	wg.Wait()

	costNormal := time.Now()
	fmt.Printf("Normal tests completed, costs = %v\n", costNormal.Sub(start))

	close(jobs)

	nIdx := 0
	for pkg := range hugePackages {
		nIdx++
		doWork(1, Job{
			UTID:    utID,
			Index:   nIdx,
			LenPkgs: lenHugePkgs,
			Pkg:     pkg,
			wg:      nil,
		})
	}

	costHuge := time.Now()
	fmt.Printf("Huge tests completed, costs = %v, total = %v\n", costHuge.Sub(costNormal), costHuge.Sub(start))

	close(addPassedPkgsJobs)

	mr := &testutils.ModuleResult{
		Name:      "datakit-ut",
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		GoVersion: runtime.Version(),
		Branch:    git.Branch,
		TestID:    utID,
		Coverage:  coverTotal.Load() / float64(len(passedPkgs)),
		Message:   fmt.Sprintf("done, total cost: %s", time.Since(start)),
	}

	if err := testutils.Flush(mr); err != nil {
		fmt.Printf("[E] flush metric failed: %s\n", err)
	}

	fmt.Printf("============ %d package passed(avg %.2f%%) ================\n",
		len(passedPkgs), coverTotal.Load()/float64(len(passedPkgs)))
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

const (
	pkgPrefix = "gitlab.jiagouyun.com/cloudcare-tools/"
)

var (
	UTExclude       string
	jobs            = make(chan Job)
	wg              sync.WaitGroup
	percentCoverage *regexp.Regexp
	coverTotal      = atomic.NewFloat64(0.0)
	excludes        = map[string]bool{}

	noTestPkgs       = make([]string, 0)
	noTestPkgsLocker sync.RWMutex

	passedPkgs        = make(map[float64][]string, 0)
	addPassedPkgsJobs chan passedPkgsOpt

	failedPkgs       = make(map[string]string, 0)
	failedPkgsLocker sync.RWMutex
)

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

////////////////////////////////////////////////////////////////////////////////

func addPassedPkgs(f float64, pkg string) {
	addPassedPkgsJobs <- passedPkgsOpt{
		F:   f,
		Pkg: pkg,
	}
}

type passedPkgsOpt struct {
	F   float64
	Pkg string
}

func workAddPassedPkgs() {
	for j := range addPassedPkgsJobs {
		passedPkgs[j.F] = append(passedPkgs[j.F], j.Pkg)
	}
}

////////////////////////////////////////////////////////////////////////////////

type Job struct {
	UTID    string // unit test ID.
	Index   int
	LenPkgs int
	Pkg     string
	wg      *sync.WaitGroup
}

func work(id int) {
	for j := range jobs {
		doWork(id, j)
	}
}

func doWork(id int, j Job) {
	defer func() {
		if j.wg != nil {
			j.wg.Done()
		}
	}()

	fmt.Printf("=======================\n")
	start := time.Now()
	fmt.Printf("[Worker %d] [%s][%02d:%02d:%02d] testing(%03d/%03d) %s...\n",
		id,
		j.UTID,
		start.Hour(),
		start.Minute(),
		start.Second(), j.Index, j.LenPkgs, j.Pkg)

	if excludes[j.Pkg] {
		fmt.Printf("[Worker %d] [%s] excluded(%03d/%03d) %s...\n", id, j.UTID, j.Index, j.LenPkgs, j.Pkg)

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

		coverage := percentCoverage.FindString(coverageLine)
		if len(coverage) != 0 {
			f, err := strconv.ParseFloat(coverage[0:len(coverage)-1], 64)
			if err != nil {
				fmt.Printf("[E] invalid coverage: %s: %s\n", coverage, err)
				return
			}

			addPassedPkgs(f, j.Pkg)
			coverTotal.Add(f)
			mr.Coverage = f
		} else {
			fmt.Printf("[W] test ok, but no coverage: %s\n", j.Pkg)
		}

	default: // pass
		fmt.Printf("[W] unknown coverage line: %s\n", coverageLine)
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
