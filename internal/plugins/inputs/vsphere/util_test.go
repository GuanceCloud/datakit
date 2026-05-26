// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package vsphere

import "testing"

func TestAnythingEnabled(t *testing.T) {
	tests := []struct {
		name    string
		exclude []string
		want    bool
	}{
		{
			name:    "empty exclude enables resource",
			exclude: nil,
			want:    true,
		},
		{
			name:    "non wildcard exclude keeps resource enabled",
			exclude: []string{"cpu.*"},
			want:    true,
		},
		{
			name:    "wildcard exclude disables resource",
			exclude: []string{"*"},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := anythingEnabled(tt.exclude); got != tt.want {
				t.Fatalf("anythingEnabled(%v) = %v, want %v", tt.exclude, got, tt.want)
			}
		})
	}
}

func TestIsSimple(t *testing.T) {
	tests := []struct {
		name    string
		include []string
		exclude []string
		want    bool
	}{
		{
			name:    "explicit include is simple",
			include: []string{"cpu.usage.average"},
			want:    true,
		},
		{
			name:    "empty include is complex",
			include: nil,
			want:    false,
		},
		{
			name:    "wildcard include is complex",
			include: []string{"cpu.*"},
			want:    false,
		},
		{
			name:    "exclude makes complex",
			include: []string{"cpu.usage.average"},
			exclude: []string{"mem.*"},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSimple(tt.include, tt.exclude); got != tt.want {
				t.Fatalf("isSimple(%v, %v) = %v, want %v", tt.include, tt.exclude, got, tt.want)
			}
		})
	}
}

func TestNewFilterOrPanic(t *testing.T) {
	filter := newFilterOrPanic([]string{"cpu.*"}, []string{"cpu.ready.*"})
	if filter == nil {
		t.Fatal("newFilterOrPanic() = nil")
	}

	defer func() {
		if recover() == nil {
			t.Fatal("newFilterOrPanic() did not panic for invalid filter")
		}
	}()
	_ = newFilterOrPanic([]string{"["}, nil)
}
