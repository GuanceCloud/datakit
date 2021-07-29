package main

import (
	"os"
	"testing"
)

func TestEnviron(t *testing.T) {
	envs := os.Environ()
	for i, x := range envs {
		t.Logf("[%02d] %s", i, x)
	}
}

func TestPromptReferences(t *testing.T) {
	promptReferences()
}
