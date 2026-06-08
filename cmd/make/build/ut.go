// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package build

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
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
	UTImpacted        bool
	UTBaseBranch      string
	Parallel          = runtime.NumCPU()
	percentCoverage   *regexp.Regexp
)

const (
	pkgPrefix                        = "gitlab.jiagouyun.com/cloudcare-tools/"
	envExcludeHugeIntegrationTesting = "UT_EXCLUDE_HUGE_INTEGRATION_TESTING"
)

var triggerAllTestsPaths = []string{
	"Makefile",
	"gitlab-ci.yml",
	"go.mod",
	"go.sum",
	"cmd/make/",
	"internal/testutils/",
	"vendor/",
}

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

	pkgs, packagesByDir, err := listPackages()
	if err != nil {
		return err
	}

	percentCoverage = regexp.MustCompile(`\d+\.\d+\%`)

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

	if UTImpacted && len(ut.only) == 0 {
		impacted, err := impactedPackages(pkgs, packagesByDir)
		if err != nil {
			l.Warnf("failed to select impacted packages, fallback to full unit tests: %s", err)
		} else if len(impacted) == 0 {
			l.Warnf("no impacted packages detected, fallback to full unit tests")
		} else if len(impacted) < len(pkgs) {
			l.Infof("run impacted unit tests only: %d/%d packages selected", len(impacted), len(pkgs))
			for _, pkg := range impacted {
				ut.only[pkg] = true
			}
		} else {
			l.Infof("impacted unit tests resolved to all packages")
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
		if p == "" {
			continue
		}

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

type goListPackage struct {
	ImportPath   string
	Dir          string
	Imports      []string
	TestImports  []string
	XTestImports []string
}

func listPackages() ([]string, map[string]string, error) {
	root, err := repoRoot()
	if err != nil {
		return nil, nil, err
	}

	cmd := exec.Command("go", "list", "-json", "./...") //nolint:gosec
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, nil, fmt.Errorf("%w: %s", err, string(exitErr.Stderr))
		}
		return nil, nil, err
	}

	dec := json.NewDecoder(strings.NewReader(string(out)))
	pkgs := []string{}
	packagesByDir := map[string]string{}

	for {
		var pkg goListPackage
		if err := dec.Decode(&pkg); err != nil {
			if err == io.EOF {
				break
			}
			return nil, nil, err
		}
		if pkg.ImportPath == "" {
			continue
		}

		pkgs = append(pkgs, pkg.ImportPath)

		rel, err := filepath.Rel(root, pkg.Dir)
		if err != nil {
			return nil, nil, err
		}
		packagesByDir[filepath.ToSlash(filepath.Clean(rel))] = pkg.ImportPath
	}

	return pkgs, packagesByDir, nil
}

func impactedPackages(allPkgs []string, packagesByDir map[string]string) ([]string, error) {
	base, err := unitTestDiffBase()
	if err != nil {
		return nil, err
	}

	deleted, err := deletedGoFiles(base)
	if err != nil {
		return nil, err
	}
	if len(deleted) > 0 {
		return allPkgs, nil
	}

	files, err := modifiedFiles(base)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, nil
	}

	if shouldRunAllUnitTests(files) {
		return allPkgs, nil
	}

	modifiedPkgs := modifiedPackages(files, packagesByDir)
	if len(modifiedPkgs) == 0 {
		return nil, nil
	}

	deps, err := packageReverseDeps()
	if err != nil {
		return nil, err
	}

	impacted := findImpactedPackages(deps, modifiedPkgs)
	ret := make([]string, 0, len(impacted))
	all := map[string]bool{}
	for _, pkg := range allPkgs {
		all[pkg] = true
	}

	for pkg := range impacted {
		if all[pkg] {
			ret = append(ret, pkg)
		}
	}

	sort.Strings(ret)
	return ret, nil
}

func unitTestDiffBase() (string, error) {
	if sha := strings.TrimSpace(os.Getenv("CI_MERGE_REQUEST_DIFF_BASE_SHA")); sha != "" {
		return sha, nil
	}

	baseBranch := strings.TrimSpace(UTBaseBranch)
	if baseBranch == "" {
		baseBranch = strings.TrimSpace(os.Getenv("CI_MERGE_REQUEST_TARGET_BRANCH_NAME"))
	}
	if baseBranch == "" {
		baseBranch = "testing"
	}

	baseRef := baseBranch
	if !strings.HasPrefix(baseRef, "origin/") {
		baseRef = "origin/" + baseRef
	}

	if err := runGit("rev-parse", "--verify", baseRef); err != nil {
		branch := strings.TrimPrefix(baseRef, "origin/")
		if fetchErr := runGit("fetch", "origin", branch+":"+baseRef); fetchErr != nil {
			return "", fmt.Errorf("verify %s: %w; fetch: %s", baseRef, err, fetchErr)
		}
	}

	out, err := gitOutput("merge-base", "HEAD", baseRef)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(out), nil
}

func deletedGoFiles(base string) ([]string, error) {
	out, err := gitOutput("diff", "--name-only", "--diff-filter=D", base+"...HEAD")
	if err != nil {
		return nil, err
	}

	files := parseDiffFiles(out)
	return goFiles(files), nil
}

func goFiles(files []string) []string {
	goFiles := files[:0]
	for _, file := range files {
		if strings.HasSuffix(file, ".go") {
			goFiles = append(goFiles, file)
		}
	}

	return goFiles
}

func modifiedFiles(base string) ([]string, error) {
	out, err := gitOutput("diff", "--name-only", "--diff-filter=AMR", base+"...HEAD")
	if err != nil {
		return nil, err
	}

	return parseDiffFiles(out), nil
}

func parseDiffFiles(out string) []string {
	files := []string{}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, filepath.ToSlash(line))
		}
	}

	return files
}

func shouldRunAllUnitTests(files []string) bool {
	for _, file := range files {
		for _, triggerPath := range triggerAllTestsPaths {
			if strings.HasSuffix(triggerPath, "/") {
				if strings.HasPrefix(file, triggerPath) {
					return true
				}
				continue
			}

			if file == triggerPath {
				return true
			}
		}
	}

	return false
}

func modifiedPackages(files []string, packagesByDir map[string]string) map[string]bool {
	modified := map[string]bool{}

	for _, file := range files {
		dir := filepath.ToSlash(filepath.Clean(filepath.Dir(file)))
		for {
			if pkg, ok := packagesByDir[dir]; ok {
				modified[pkg] = true
				break
			}

			next := filepath.ToSlash(filepath.Dir(dir))
			if next == "." || next == "/" || next == dir {
				break
			}
			dir = next
		}
	}

	return modified
}

func packageReverseDeps() (map[string][]string, error) {
	root, err := repoRoot()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("go", "list", "-json", "./...") //nolint:gosec
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("%w: %s", err, string(exitErr.Stderr))
		}
		return nil, err
	}

	dec := json.NewDecoder(strings.NewReader(string(out)))
	deps := map[string][]string{}
	for {
		var pkg goListPackage
		if err := dec.Decode(&pkg); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		imports := append([]string{}, pkg.Imports...)
		imports = append(imports, pkg.TestImports...)
		imports = append(imports, pkg.XTestImports...)

		for _, imported := range imports {
			if strings.HasPrefix(imported, "gitlab.jiagouyun.com/cloudcare-tools/datakit") {
				deps[imported] = append(deps[imported], pkg.ImportPath)
			}
		}
	}

	return deps, nil
}

func findImpactedPackages(deps map[string][]string, modified map[string]bool) map[string]bool {
	impacted := map[string]bool{}
	stack := []string{}
	for pkg := range modified {
		stack = append(stack, pkg)
	}

	for len(stack) > 0 {
		pkg := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if impacted[pkg] {
			continue
		}

		impacted[pkg] = true
		stack = append(stack, deps[pkg]...)
	}

	return impacted
}

func gitOutput(args ...string) (string, error) {
	cmd := exec.Command("git", args...) //nolint:gosec
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, string(out))
	}

	return string(out), nil
}

func runGit(args ...string) error {
	_, err := gitOutput(args...)
	return err
}

func repoRoot() (string, error) {
	out, err := gitOutput("rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(out), nil
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
		"GO111MODULE=on",
		"GOFLAGS=-mod=vendor",
		"CGO_ENABLED=1",
		"LOGGER_PATH=nul", // disable logging
	}...)

	res, err := tcmd.CombinedOutput()
	mr.Cost = time.Since(start)
	if len(res) > 0 {
		mr.Message = string(res)
	}

	if err != nil {
		if strings.Contains(mr.Message, "no Go files in") {
			ut.addNoTestPkgs(j.pkg)
			mr.Status = testutils.TestSkipped
			mr.NoTest = true
			if err := testutils.Flush(mr); err != nil {
				l.Errorf("flush metric failed: %s", err)
				return
			}

			j.show(ut, mr)
			return
		}

		ut.addFailedPkgs(j.pkg, string(res))

		mr.Status = testutils.TestFailed
		mr.FailedMessage = err.Error()
		if err := testutils.Flush(mr); err != nil {
			l.Errorf("flush metric failed: %s", err)
		}

		l.Errorf("package %s failed: %s", j.pkg, string(res))
		return
	}

	lines := strings.Split(string(res), "\n")
	coverageLine := strings.TrimSpace(lines[len(lines)-2])

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
		strings.Contains(coverageLine, "[no test files]"),
		strings.Contains(coverageLine, "[no tests to run]"),
		strings.Contains(coverageLine, "[no statements]"):
		ut.addNoTestPkgs(j.pkg)
		mr.NoTest = true

	case strings.HasPrefix(coverageLine, "ok"):
		fallthrough
	case strings.Contains(coverageLine, "coverage:"):
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
