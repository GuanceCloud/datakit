// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindBestMount(t *testing.T) {
	tests := []struct {
		name    string
		mounts  Mounts
		path    string
		wantDst string
		wantSrc string
		wantOK  bool
	}{
		{
			name: "simple best match",
			mounts: Mounts{
				{Destination: "/tmp/opt", Source: "/mount/container-id/k8s.io/empty-dir/"},
			},
			path:    "/tmp/opt/abc/123.log",
			wantDst: "/tmp/opt",
			wantSrc: "/mount/container-id/k8s.io/empty-dir",
			wantOK:  true,
		},
		{
			name: "longest destination prefix wins",
			mounts: Mounts{
				{Destination: "/tmp/opt", Source: "/mount/a"},
				{Destination: "/tmp/opt/abc", Source: "/mount/b"},
			},
			path:    "/tmp/opt/abc/123.log",
			wantDst: "/tmp/opt/abc",
			wantSrc: "/mount/b",
			wantOK:  true,
		},
		{
			name: "exact match choose dest",
			mounts: Mounts{
				{Destination: "/tmp/opt", Source: "/mount/source"},
			},
			path:    "/tmp/opt",
			wantDst: "/tmp/opt",
			wantSrc: "/mount/source",
			wantOK:  true,
		},
		{
			name: "no matching destination",
			mounts: Mounts{
				{Destination: "/var/lib", Source: "/mount/source"},
			},
			path:   "/tmp/opt/abc/123.log",
			wantOK: false,
		},
		{
			name: "destination with trailing slash is normalized",
			mounts: Mounts{
				{Destination: "/tmp/opt/", Source: "/mount/s"},
			},
			path:    "/tmp/opt/abc/123.log",
			wantDst: "/tmp/opt",
			wantSrc: "/mount/s",
			wantOK:  true,
		},
		{
			name: "path with redundant slashes is cleaned",
			mounts: Mounts{
				{Destination: "/tmp/opt", Source: "/mount/s"},
			},
			path:    "/tmp//opt///abc/123.log",
			wantDst: "/tmp/opt",
			wantSrc: "/mount/s",
			wantOK:  true,
		},
		{
			name: "multiple candidates choose longest prefix even with trailing slash in source",
			mounts: Mounts{
				{Destination: "/tmp/opt", Source: "/mount/a/"},
				{Destination: "/tmp/opt/abc", Source: "/mount/b/"},
			},
			path:    "/tmp/opt/abc/xyz.txt",
			wantDst: "/tmp/opt/abc",
			wantSrc: "/mount/b",
			wantOK:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst, src, ok := tt.mounts.FindBestMount(tt.path)
			assert.Equal(t, tt.wantOK, ok, "ok mismatch")
			if ok {
				assert.Equal(t, tt.wantDst, dst, "destination mismatch")
				assert.Equal(t, tt.wantSrc, src, "source mismatch")
			}
		})
	}
}

func TestResolveToSourcePath(t *testing.T) {
	tests := []struct {
		name   string
		mounts Mounts
		path   string
		want   string
		wantOK bool
	}{
		{
			name: "simple match and append suffix",
			mounts: Mounts{
				{Destination: "/tmp/opt", Source: "/mount/container-id/k8s.io/empty-dir/"},
			},
			path:   "/tmp/opt/abc/123.log",
			want:   "/mount/container-id/k8s.io/empty-dir/abc/123.log",
			wantOK: true,
		},
		{
			name: "longest destination prefix wins",
			mounts: Mounts{
				{Destination: "/tmp/opt", Source: "/mount/a"},
				{Destination: "/tmp/opt/abc", Source: "/mount/b"},
			},
			path:   "/tmp/opt/abc/123.log",
			want:   "/mount/b/123.log",
			wantOK: true,
		},
		{
			name: "exact match returns source as-is",
			mounts: Mounts{
				{Destination: "/tmp/opt", Source: "/mount/source"},
			},
			path:   "/tmp/opt",
			want:   "/mount/source",
			wantOK: true,
		},
		{
			name: "destination and source both have trailing slash",
			mounts: Mounts{
				{Destination: "/d1/", Source: "/s1/"},
			},
			path:   "/d1/a/b",
			want:   "/s1/a/b",
			wantOK: true,
		},
		{
			name: "no matching destination",
			mounts: Mounts{
				{Destination: "/var/lib", Source: "/mount/source"},
			},
			path:   "/tmp/opt/abc/123.log",
			want:   "",
			wantOK: false,
		},
		{
			name: "destination with trailing slash is normalized",
			mounts: Mounts{
				{Destination: "/tmp/opt/", Source: "/mount/s"},
			},
			path:   "/tmp/opt/abc/123.log",
			want:   "/mount/s/abc/123.log",
			wantOK: true,
		},
		{
			name: "path with redundant slashes is cleaned",
			mounts: Mounts{
				{Destination: "/tmp/opt", Source: "/mount/s"},
			},
			path:   "/tmp//opt///abc/123.log",
			want:   "/mount/s/abc/123.log",
			wantOK: true,
		},
		{
			name: "multiple candidates choose longest prefix even with trailing slash in source",
			mounts: Mounts{
				{Destination: "/tmp/opt", Source: "/mount/a/"},
				{Destination: "/tmp/opt/abc", Source: "/mount/b/"},
			},
			path:   "/tmp/opt/abc/xyz.txt",
			want:   "/mount/b/xyz.txt",
			wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst, src, ok := tt.mounts.FindBestMount(tt.path)
			assert.Equal(t, tt.wantOK, ok, "find ok mismatch")
			if !ok {
				return
			}
			got, ok2 := ResolveToSourcePath(dst, src, tt.path)
			assert.True(t, ok2, "mapping should succeed after find")
			assert.Equal(t, tt.want, got, "path mismatch")
		})
	}
}

func TestReverseResolveFromSourcePath(t *testing.T) {
	tests := []struct {
		name         string
		destination  string
		source       string
		mapped       string
		wantOriginal string
		wantOK       bool
	}{
		{
			name:         "simple reverse mapping",
			destination:  "/tmp/opt/abc",
			source:       "/mount/a",
			mapped:       "/mount/a/123.go",
			wantOriginal: "/tmp/opt/abc/123.go",
			wantOK:       true,
		},
		{
			name:         "empty source",
			destination:  "",
			source:       "/run/containerd/k8s/rootfs",
			mapped:       "/run/containerd/k8s/rootfs/tmp/opt/123.log",
			wantOriginal: "/tmp/opt/123.log",
			wantOK:       true,
		},
		{
			name:         "exact match returns destination",
			destination:  "/tmp/opt",
			source:       "/mount/source",
			mapped:       "/mount/source",
			wantOriginal: "/tmp/opt",
			wantOK:       true,
		},
		{
			name:         "not under source returns false",
			destination:  "/tmp/opt",
			source:       "/mount/a",
			mapped:       "/mount/b/123.go",
			wantOriginal: "",
			wantOK:       false,
		},
		{
			name:         "normalize trailing slashes",
			destination:  "/tmp/opt",
			source:       "/mount/s/",
			mapped:       "/mount/s/abc/123.log",
			wantOriginal: "/tmp/opt/abc/123.log",
			wantOK:       true,
		},
		{
			name:         "root source special case",
			destination:  "/mnt/dest",
			source:       "/",
			mapped:       "/abc/def",
			wantOriginal: "/mnt/dest/abc/def",
			wantOK:       true,
		},
		{
			name:         "round-trip: forward then reverse",
			destination:  "/tmp/opt/abc",
			source:       "/mount/b",
			mapped:       func() string { m, _ := ResolveToSourcePath("/tmp/opt/abc", "/mount/b", "/tmp/opt/abc/x/y"); return m }(),
			wantOriginal: "/tmp/opt/abc/x/y",
			wantOK:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ReverseResolveFromSourcePath(tt.mapped, tt.destination, tt.source)
			assert.Equal(t, tt.wantOK, ok, "ok mismatch")
			if ok {
				assert.Equal(t, tt.wantOriginal, got, "original path mismatch")
			} else {
				assert.Equal(t, "", got, "expected empty result when ok=false")
			}
		})
	}
}
