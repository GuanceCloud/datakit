// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package build

import (
	"reflect"
	"testing"
)

func TestModifiedPackagesWalksUpToNearestGoPackage(t *testing.T) {
	packagesByDir := map[string]string{
		"internal/foo":     "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/foo",
		"internal/foo/bar": "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/foo/bar",
	}

	got := modifiedPackages([]string{
		"internal/foo/testdata/case.json",
		"internal/foo/bar/bar.go",
		"docs/readme.md",
	}, packagesByDir)

	want := map[string]bool{
		"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/foo":     true,
		"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/foo/bar": true,
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("modifiedPackages() = %#v, want %#v", got, want)
	}
}

func TestListPackagesMapsPackageDirs(t *testing.T) {
	pkgs, packagesByDir, err := listPackages()
	if err != nil {
		t.Fatal(err)
	}
	if len(pkgs) == 0 {
		t.Fatal("expected at least one package")
	}

	const want = "gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/make/build"
	if got := packagesByDir["cmd/make/build"]; got != want {
		t.Fatalf("packagesByDir[cmd/make/build] = %q, want %q", got, want)
	}
}

func TestParseDiffFilesNormalizesPaths(t *testing.T) {
	got := parseDiffFiles("internal/foo/input.go\n\n docs/readme.md \n")
	want := []string{"internal/foo/input.go", "docs/readme.md"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseDiffFiles() = %#v, want %#v", got, want)
	}
}

func TestGoFilesKeepsOnlyGoFiles(t *testing.T) {
	got := goFiles([]string{
		"internal/foo/input.go",
		"internal/foo/input_test.go",
		"docs/readme.md",
		"internal/foo/testdata/case.json",
	})
	want := []string{
		"internal/foo/input.go",
		"internal/foo/input_test.go",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("goFiles() = %#v, want %#v", got, want)
	}
}

func TestFindImpactedPackagesIncludesReverseDeps(t *testing.T) {
	deps := map[string][]string{
		"pkg/a": {"pkg/b", "pkg/c"},
		"pkg/b": {"pkg/d"},
	}

	got := findImpactedPackages(deps, map[string]bool{"pkg/a": true})
	want := map[string]bool{
		"pkg/a": true,
		"pkg/b": true,
		"pkg/c": true,
		"pkg/d": true,
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("findImpactedPackages() = %#v, want %#v", got, want)
	}
}

func TestShouldRunAllUnitTests(t *testing.T) {
	if !shouldRunAllUnitTests([]string{"cmd/make/build/ut.go"}) {
		t.Fatal("cmd/make changes should trigger all unit tests")
	}

	if shouldRunAllUnitTests([]string{"internal/plugins/inputs/cpu/input.go"}) {
		t.Fatal("ordinary input changes should not trigger all unit tests")
	}
}
