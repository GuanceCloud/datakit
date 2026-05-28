//go:build linux
// +build linux

package run

import "testing"

func TestGoMemoryLimitRatio(t *testing.T) {
	t.Setenv(envGoMemoryLimitRatio, "")
	if got := goMemoryLimitRatio(); got != defaultGoMemoryLimitRatio {
		t.Fatalf("default ratio = %v, want %v", got, defaultGoMemoryLimitRatio)
	}

	t.Setenv(envGoMemoryLimitRatio, "0.55")
	if got := goMemoryLimitRatio(); got != 0.55 {
		t.Fatalf("configured ratio = %v, want 0.55", got)
	}

	t.Setenv(envGoMemoryLimitRatio, "1.5")
	if got := goMemoryLimitRatio(); got != defaultGoMemoryLimitRatio {
		t.Fatalf("invalid ratio = %v, want %v", got, defaultGoMemoryLimitRatio)
	}
}
