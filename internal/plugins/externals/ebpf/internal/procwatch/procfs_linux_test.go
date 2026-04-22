//go:build linux
// +build linux

package procwatch

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNormalizeProcfsBinaryPath(t *testing.T) {
	cases := map[string]string{
		"":                         "",
		"relative/path":            "",
		"[vdso]":                   "",
		"/tmp/demo (deleted)":      "/tmp/demo",
		" /usr/bin/bash ":          "/usr/bin/bash",
		"/usr/bin/../bin/python3":  "/usr/bin/python3",
		string(filepath.Separator): "",
	}

	for input, want := range cases {
		if got := normalizeProcfsBinaryPath(input); got != want {
			t.Fatalf("normalizeProcfsBinaryPath(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestIsRegularFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "app")
	if err := os.WriteFile(file, []byte("bin"), 0o755); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	if !isRegularFile(file) {
		t.Fatalf("expected regular file for %s", file)
	}
	if isRegularFile(dir) {
		t.Fatalf("expected directory %s to be rejected", dir)
	}
	if isRegularFile(filepath.Join(dir, "missing")) {
		t.Fatal("expected missing file to be rejected")
	}
}

func TestParseMountInfoLine(t *testing.T) {
	line := []byte("36 25 0:32 /docker/containers /var/lib/docker rw,relatime - overlay overlay rw")

	mount, ok := parseMountInfoLine(line)
	if !ok {
		t.Fatal("expected mountinfo line to parse")
	}
	if mount.dev != "0:32" || mount.root != "/docker/containers" || mount.mountPoint != "/var/lib/docker" {
		t.Fatalf("unexpected mountinfo parse result: dev=%q root=%q mountPoint=%q", mount.dev, mount.root, mount.mountPoint)
	}
}

func TestParseMountInfoLineUnescapesPaths(t *testing.T) {
	line := []byte(`36 25 0:32 /docker\040root /var/lib/container\040mount rw,relatime - overlay overlay rw`)

	mount, ok := parseMountInfoLine(line)
	if !ok {
		t.Fatal("expected escaped mountinfo line to parse")
	}
	if mount.dev != "0:32" || mount.root != "/docker root" || mount.mountPoint != "/var/lib/container mount" {
		t.Fatalf("unexpected escaped mountinfo parse result: dev=%q root=%q mountPoint=%q", mount.dev, mount.root, mount.mountPoint)
	}
}

func TestResolveMountPath(t *testing.T) {
	rootMounts := &mountSnapshot{
		mounts: []*mountEntry{
			{dev: "0:32", root: "/docker/containers/abcd/rootfs", mountPoint: "/var/lib/docker/containers/abcd/rootfs"},
			{dev: "8:1", root: "/", mountPoint: "/"},
		},
	}
	nsMounts := &mountSnapshot{
		mounts: []*mountEntry{
			{dev: "0:32", root: "/docker/containers/abcd/rootfs", mountPoint: "/"},
		},
	}

	got := resolveMountPath(rootMounts, indexMountsByDev(rootMounts), nsMounts, "/usr/bin/app")
	want := "/var/lib/docker/containers/abcd/rootfs/usr/bin/app"
	if got != want {
		t.Fatalf("resolveMountPath() = %q, want %q", got, want)
	}
}

func TestMountSnapshotFindUsesPathBoundary(t *testing.T) {
	snapshot := &mountSnapshot{
		mounts: []*mountEntry{
			{mountPoint: "/var/lib/docker"},
			{mountPoint: "/var/lib/docker2"},
			{mountPoint: "/"},
		},
	}

	if got := snapshot.find("/var/lib/docker2/overlay2/merged/usr/bin/app"); got == nil || got.mountPoint != "/var/lib/docker2" {
		t.Fatalf("expected longest boundary-safe match, got %+v", got)
	}
	if got := snapshot.find("/var/lib/dockerx/usr/bin/app"); got == nil || got.mountPoint != "/" {
		t.Fatalf("expected root mount fallback for sibling prefix, got %+v", got)
	}
}

func TestRootMountCandidatesUsesDeviceIndex(t *testing.T) {
	rootMounts := &mountSnapshot{
		mounts: []*mountEntry{
			{dev: "0:32", root: "/docker", mountPoint: "/var/lib/docker"},
			{dev: "8:1", root: "/", mountPoint: "/"},
		},
	}

	candidates := rootMountCandidates(rootMounts, indexMountsByDev(rootMounts), "0:32")
	if len(candidates) != 1 || candidates[0].mountPoint != "/var/lib/docker" {
		t.Fatalf("unexpected candidates: %+v", candidates)
	}
}

func TestReadProcessEnvironMapForKeys(t *testing.T) {
	procRoot := t.TempDir()
	t.Setenv("HOST_PROC", procRoot)

	pidDir := filepath.Join(procRoot, "123")
	if err := os.MkdirAll(pidDir, 0o755); err != nil {
		t.Fatal(err)
	}

	payload := []byte("DD_SERVICE=checkout\x00IGNORED=value\x00OTEL_SERVICE_NAME=frontend\x00")
	if err := os.WriteFile(filepath.Join(pidDir, "environ"), payload, 0o644); err != nil {
		t.Fatal(err)
	}

	env := readProcessEnvironMapForKeys(123, map[string]struct{}{
		"DD_SERVICE":        {},
		"OTEL_SERVICE_NAME": {},
	})
	if len(env) != 2 {
		t.Fatalf("unexpected env size: %d", len(env))
	}
	if env["DD_SERVICE"] != "checkout" {
		t.Fatalf("unexpected DD_SERVICE value: %q", env["DD_SERVICE"])
	}
	if env["OTEL_SERVICE_NAME"] != "frontend" {
		t.Fatalf("unexpected OTEL_SERVICE_NAME value: %q", env["OTEL_SERVICE_NAME"])
	}
	if _, ok := env["IGNORED"]; ok {
		t.Fatal("expected unselected env var to be skipped")
	}
}

func TestScanSharedLibrariesPreservesSpacesInPath(t *testing.T) {
	pidPath := t.TempDir()
	mapsPath := filepath.Join(pidPath, "maps")
	line := "7f0000000000-7f0000001000 r-xp 00000000 08:02 12345 /tmp/lib with space.so\n"
	if err := os.WriteFile(mapsPath, []byte(line), 0o644); err != nil {
		t.Fatal(err)
	}

	libs := scanSharedLibraries(pidPath, nil)
	if len(libs) != 1 || libs[0] != "/tmp/lib with space.so" {
		t.Fatalf("unexpected libs: %+v", libs)
	}
}

func TestUnescapeProcPathField(t *testing.T) {
	raw := []byte(`/var/lib/docker\040merged/foo\134bar\012baz`)
	if got := unescapeProcPathField(raw); got != "/var/lib/docker merged/foo\\bar\nbaz" {
		t.Fatalf("unexpected unescaped path: %q", got)
	}
}

func TestResolveHostBinaryPathUsesProcRootFirst(t *testing.T) {
	procRoot := t.TempDir()
	t.Setenv("HOST_PROC", procRoot)

	const pid = 4321
	pidDir := filepath.Join(procRoot, "4321")
	rootfs := filepath.Join(procRoot, "rootfs")
	binPath := filepath.Join(rootfs, "usr", "bin", "app")

	if err := os.MkdirAll(filepath.Join(pidDir, "ns"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(binPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(binPath, []byte("bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pidDir, "ns", "mnt"), []byte("mntns"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(rootfs, filepath.Join(pidDir, "root")); err != nil {
		t.Fatal(err)
	}

	got := resolveHostBinaryPath(pid, "/usr/bin/app")
	if got != binPath {
		t.Fatalf("resolveHostBinaryPath() = %q, want %q", got, binPath)
	}
}

func TestResolveHostBinaryPathUsesProcRootPortal(t *testing.T) {
	procRoot := t.TempDir()
	t.Setenv("HOST_PROC", procRoot)

	const pid = 5321
	binPath := filepath.Join(procRoot, "5321", "root", "usr", "bin", "app")
	if err := os.MkdirAll(filepath.Dir(binPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(binPath, []byte("bin"), 0o755); err != nil {
		t.Fatal(err)
	}

	got := resolveHostBinaryPath(pid, "/usr/bin/app")
	if got != binPath {
		t.Fatalf("resolveHostBinaryPath() = %q, want %q", got, binPath)
	}
}

func TestResolveMountPathOverlayReturnsHostLayerFile(t *testing.T) {
	layer1 := t.TempDir()
	layer2 := t.TempDir()
	binPath := filepath.Join(layer2, "usr", "local", "bin", "smrepro")
	if err := os.MkdirAll(filepath.Dir(binPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(binPath, []byte("bin"), 0o755); err != nil {
		t.Fatal(err)
	}

	nsMounts := &mountSnapshot{
		mounts: []*mountEntry{
			{
				dev:        "0:44",
				root:       "/",
				mountPoint: "/",
				fsType:     "overlay",
				options: map[string]string{
					"lowerdir": layer1 + ":" + layer2,
				},
			},
		},
	}

	got := resolveMountPath(&mountSnapshot{}, nil, nsMounts, "/usr/local/bin/smrepro")
	if got != binPath {
		t.Fatalf("resolveMountPath() = %q, want %q", got, binPath)
	}
}

func TestSplitEscapedMountPathListPreservesEscapedSeparator(t *testing.T) {
	raw := `/var/lib/containerd/layer\072one:/var/lib/containerd/layer-two`
	got := splitEscapedMountPathList(raw, ':')
	want := []string{
		"/var/lib/containerd/layer:one",
		"/var/lib/containerd/layer-two",
	}
	if len(got) != len(want) {
		t.Fatalf("splitEscapedMountPathList() len = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("splitEscapedMountPathList()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestResolveMountPathOverlayHandlesEscapedLowerdirPath(t *testing.T) {
	base := t.TempDir()
	layer1 := filepath.Join(base, "layer:one")
	layer2 := filepath.Join(base, "layer-two")
	binPath := filepath.Join(layer1, "usr", "local", "bin", "smrepro")
	if err := os.MkdirAll(filepath.Dir(binPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(layer2, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(binPath, []byte("bin"), 0o755); err != nil {
		t.Fatal(err)
	}

	escapedLayer1 := strings.ReplaceAll(layer1, ":", `\072`)
	nsMounts := &mountSnapshot{
		mounts: []*mountEntry{
			{
				dev:        "0:44",
				root:       "/",
				mountPoint: "/",
				fsType:     "overlay",
				options: map[string]string{
					"lowerdir": escapedLayer1 + ":" + layer2,
				},
			},
		},
	}

	got := resolveMountPath(&mountSnapshot{}, nil, nsMounts, "/usr/local/bin/smrepro")
	if got != binPath {
		t.Fatalf("resolveMountPath() = %q, want %q", got, binPath)
	}
}

func TestNormalizeProcessLinkTargetAllowsContainerAbsolutePath(t *testing.T) {
	got, err := normalizeProcessLinkTarget(1234, "executable", "/usr/local/bin/smrepro")
	if err != nil {
		t.Fatalf("normalizeProcessLinkTarget() error = %v", err)
	}
	if got != "/usr/local/bin/smrepro" {
		t.Fatalf("normalizeProcessLinkTarget() = %q, want %q", got, "/usr/local/bin/smrepro")
	}
}

func TestNormalizeProcessLinkTargetRejectsDeletedBinary(t *testing.T) {
	got, err := normalizeProcessLinkTarget(1234, "executable", "/usr/local/bin/smrepro (deleted)")
	if got != "" {
		t.Fatalf("normalizeProcessLinkTarget() = %q, want empty", got)
	}
	if !isNonRegularExecutablePathError(err) {
		t.Fatalf("expected non-regular executable error, got %v", err)
	}
}

func TestResolveDoesNotCacheEmptyPath(t *testing.T) {
	dir := t.TempDir()
	pidPath := filepath.Join(dir, "200")
	if err := os.MkdirAll(filepath.Join(pidPath, "ns"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pidPath, "ns", "mnt"), []byte("pid-ns"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pidPath, "mountinfo"), []byte("36 25 0:32 /container / rw - overlay overlay rw\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	rootMounts := &mountSnapshot{
		mounts: []*mountEntry{
			{dev: "8:1", root: "/", mountPoint: "/"},
		},
	}
	cache := &pathResolverCache{
		rootMounts: rootMounts,
		rootNS:     mountNamespace{dev: 1, ino: 1},
		rootByDev:  indexMountsByDev(rootMounts),
		nsMounts:   make(map[mountNamespace]*mountSnapshot),
		resolved:   make(map[mountNamespace]map[string]resolvedPathCacheEntry),
		lastAccess: make(map[mountNamespace]time.Time),
	}

	if got := cache.Resolve(pidPath, "/usr/bin/app"); got != "" {
		t.Fatalf("Resolve() = %q, want empty path", got)
	}

	ns := readMountNamespace(pidPath)
	if ns == (mountNamespace{}) {
		t.Fatal("expected test namespace to be readable")
	}
	if cache.resolved[ns] != nil {
		t.Fatalf("expected empty resolve result to avoid cache entry, got %+v", cache.resolved[ns])
	}
}

func TestMaybeCleanupLockedDropsExpiredNamespaceEntries(t *testing.T) {
	ns := mountNamespace{dev: 2, ino: 3}
	now := time.Now()
	cache := &pathResolverCache{
		nsMounts: map[mountNamespace]*mountSnapshot{
			ns: {mounts: []*mountEntry{{dev: "0:32", root: "/", mountPoint: "/"}}},
		},
		resolved: map[mountNamespace]map[string]resolvedPathCacheEntry{
			ns: {
				"/usr/bin/app": {path: "/host/app", ts: now.Add(-resolvedPathCacheTTL - time.Second)},
			},
		},
		lastAccess: map[mountNamespace]time.Time{
			ns: now.Add(-resolvedPathNamespaceTTL - time.Second),
		},
	}

	cache.maybeCleanupLocked(now)

	if _, ok := cache.nsMounts[ns]; ok {
		t.Fatal("expected expired namespace mounts to be removed")
	}
	if _, ok := cache.resolved[ns]; ok {
		t.Fatal("expected expired resolve cache to be removed")
	}
	if _, ok := cache.lastAccess[ns]; ok {
		t.Fatal("expected expired namespace access record to be removed")
	}
}
