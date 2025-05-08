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
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
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
	Parallel          = runtime.NumCPU()
	percentCoverage   *regexp.Regexp
)

const (
	pkgPrefix                        = "gitlab.jiagouyun.com/cloudcare-tools/"
	envExcludeHugeIntegrationTesting = "UT_EXCLUDE_HUGE_INTEGRATION_TESTING"
)

type job struct {
	UTID    string // unit test ID.
	index   int
	lenPkgs int
	pkg     string
}

type unitTest struct {
	only, exclude map[string]bool

	noTestPkgs  []string
	pkgCoverage map[float64][]string

	npassed,
	nskipped,
	nhuge atomic.Int64
	coverTotal atomic.Float64

	failedPkgs map[string]string
	mtx        sync.RWMutex
}

func defaultUnitTest() *unitTest {
	return &unitTest{
		only: map[string]bool{},
		exclude: map[string]bool{
			// There are multiple-main() within these modules.
			"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/proxy/bench": true,

			// root package got no test to run.
			"gitlab.jiagouyun.com/cloudcare-tools/datakit":         true,
			"gitlab.jiagouyun.com/cloudcare-tools/datakit/scripts": true,
		},
		pkgCoverage: map[float64][]string{},
		failedPkgs:  map[string]string{},
	}
}

func (ut *unitTest) jobWorker(ch chan *job) {
	for j := range ch {
		ut.doWork(j)
	}
}

func UnitTestDataKit() error {
	ut := defaultUnitTest()

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
			l.Debugf("package %q excluded", ex)
			ut.exclude[ex] = true
		}
	}

	if len(UTOnly) > 0 && UTOnly != "-" {
		for _, item := range strings.Split(UTOnly, ",") {
			l.Debugf("package %q selected", item)
			ut.only[item] = true
		}
	}

	start := time.Now()

	var wg sync.WaitGroup
	if Parallel < 0 {
		Parallel = 1
	} else if Parallel == 0 {
		Parallel = runtime.NumCPU()
	}

	wg.Add(Parallel)
	jobCh := make(chan *job, Parallel)

	for i := 0; i < Parallel; i++ {
		go func() {
			defer wg.Done()
			ut.jobWorker(jobCh)
		}()
	}

	lenPkgs := len(pkgs)
	for i, p := range pkgs {
		i++
		if hugePackages[p] {
			l.Debugf("%s is HUGE package, testing it later, skip...", p)
			ut.nhuge.Add(1)
			continue
		}

		jobCh <- &job{
			UTID:    utID,
			index:   i,
			lenPkgs: lenPkgs,
			pkg:     p,
		}
	}

	close(jobCh)

	wg.Wait()

	costNormal := time.Now()
	l.Debugf("Normal tests completed, costs = %v", costNormal.Sub(start))

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
			if ut.exclude[pkg] {
				l.Debugf("Skip huge test %q", pkg)
				continue
			}

			nIdx++
			l.Debugf("[%s] run huge test %q", time.Now(), pkg)
			ut.doWork(&job{
				UTID:    utID,
				index:   nIdx,
				lenPkgs: lenHugePkgs,
				pkg:     pkg,
			})
		}

		l.Debugf("Huge tests completed, elapsed: %v", time.Since(costNormal))
	}

	l.Debugf("All tests done, elapsed: %v", time.Since(start))

	mr := &testutils.ModuleResult{
		Name:      "datakit-ut",
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		GoVersion: runtime.Version(),
		Branch:    git.Branch,
		TestID:    utID,
		Coverage:  ut.coverTotal.Load() / float64(ut.npassed.Load()),
		Message:   fmt.Sprintf("done, total cost: %s", time.Since(start)),
	}

	if err := testutils.Flush(mr); err != nil {
		l.Debugf("[E] flush metric failed: %s", err)
	}

	ut.show()

	if len(ut.failedPkgs) > 0 {
		return fmt.Errorf("%d package failed", len(ut.failedPkgs))
	}

	return nil
}

func (ut *unitTest) show() {
	cp.Printf("============ %d package passed(avg %.2f%%) ================\n",
		ut.npassed.Load(), ut.coverTotal.Load()/float64(ut.npassed.Load()))
	ut.showTopNCoveragePkgs()

	cp.Printf("============= %d package got no test ===============\n", len(ut.noTestPkgs))
	sort.Strings(ut.noTestPkgs)
	ut.showNoTestPkgs()

	cp.Printf("============= %d pakage failed ===============\n", len(ut.failedPkgs))
	ut.showFailedPkgs()
}

func (ut *unitTest) addNoTestPkgs(pkg string) {
	ut.mtx.Lock()
	defer ut.mtx.Unlock()
	ut.noTestPkgs = append(ut.noTestPkgs, pkg)
}

func (ut *unitTest) addFailedPkgs(pkg, detail string) {
	ut.mtx.Lock()
	defer ut.mtx.Unlock()
	ut.failedPkgs[pkg] = detail
}

func (ut *unitTest) doWork(j *job) {
	start := time.Now()

	if ut.exclude[j.pkg] {
		l.Debugf("[%s] package(%03d/%03d) %s excluded...",
			j.UTID, j.index, j.lenPkgs, j.pkg)
		ut.nskipped.Add(1)
		return
	}

	if len(ut.only) > 0 && !ut.only[j.pkg] {
		l.Debugf("[%s] package(%03d/%03d) %s not selected, selected: %+#v",
			j.UTID, j.index, j.lenPkgs, j.pkg, ut.only)
		ut.nskipped.Add(1)
		return
	}

	mr := &testutils.ModuleResult{
		Name:      strings.TrimPrefix(j.pkg, pkgPrefix), // remove prefix for human readable.
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		GoVersion: runtime.Version(),
		Branch:    git.Branch,
		TestID:    j.UTID,
	}

	tcmd := exec.Command("go", "test", "-count=1", "-timeout", "1h", "-cover", j.pkg) //nolint:gosec
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
			ut.addFailedPkgs(j.pkg, string(res))

			mr.Status = testutils.TestFailed
			mr.FailedMessage = err.Error()
			if err := testutils.Flush(mr); err != nil {
				l.Errorf("flush metric failed: %s", err)
			}

			l.Errorf("package %s failed: %s", j.pkg, string(res))
			return
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

	switch {
	case strings.HasPrefix(coverageLine, "?"),
		strings.Contains(coverageLine, "[no tests to run]"):
		ut.addNoTestPkgs(j.pkg)
		mr.NoTest = true

	case strings.HasPrefix(coverageLine, "ok"):
		mr.Status = testutils.TestPassed
		ut.npassed.Add(1)

		coverage := percentCoverage.FindString(coverageLine)
		if len(coverage) != 0 {
			f, err := strconv.ParseFloat(coverage[0:len(coverage)-1], 64)
			if err != nil {
				l.Errorf("invalid coverage %q: %s: %s", j.pkg, coverage, err)
				return
			}

			ut.addCoveragePkgs(f, j.pkg)
			ut.coverTotal.Add(f)
			mr.Coverage = f
		} else {
			l.Warnf("test ok, but no coverage: %q", j.pkg)
			return
		}

	default: // pass
		l.Warnf("unknown coverage line in package %q: %s", j.pkg, coverageLine)
		return
	}

	if err := testutils.Flush(mr); err != nil {
		l.Errorf("flush metric failed: %s", err)
		return
	}

	j.show(ut, mr)
}

func (j *job) show(ut *unitTest, mr *testutils.ModuleResult) {
	// here will access ut.failedPkgs, we lock it to avoid map concurrent access.
	ut.mtx.Lock()
	defer ut.mtx.Unlock()
	cp.Printf("%s | %d | passed:%d/notest:%d/skipped:%d/huge:%d/failed:%d | %03d/%03d | %s | %f%% | %v\n",
		j.UTID,
		Parallel,
		ut.npassed.Load(),
		len(ut.noTestPkgs),
		ut.nskipped.Load(),
		ut.nhuge.Load(),
		len(ut.failedPkgs),
		j.index,
		j.lenPkgs,
		j.pkg,
		mr.Coverage,
		mr.Cost,
	)
}

func (ut *unitTest) addCoveragePkgs(cov float64, pkg string) {
	ut.mtx.Lock()
	defer ut.mtx.Unlock()

	ut.pkgCoverage[cov] = append(ut.pkgCoverage[cov], pkg)
}

func (ut *unitTest) showTopNCoveragePkgs() {
	topn := []float64{}
	for k := range ut.pkgCoverage {
		topn = append(topn, k)
	}

	sort.Float64s(topn)
	for _, c := range topn {
		cp.Printf("%.2f%%\n\t%s\n", c, strings.Join(ut.pkgCoverage[c], "\n\t"))
	}
}

func (ut *unitTest) showFailedPkgs() {
	for k, v := range ut.failedPkgs {
		cp.Printf("%s\n%s\n", k, v)
		cp.Println("----------------------------")
	}
}

func (ut *unitTest) showNoTestPkgs() {
	for _, p := range ut.noTestPkgs {
		cp.Printf("%s\n", p)
	}
}
