package main

import (
	"testing"
)

func TestFilterEnvs(t *testing.T) {
	cases := []string{
		"tmux.*",
		"^[A-Z]+",
		"^ENV_.*",
	}

	for _, c := range cases {
		envs := filterEnvs(c)
		t.Logf("get %d envs from `%s'", len(envs), c)
		if len(envs) < 10 {
			for k, v := range envs {
				t.Logf("%s = %s", k, v)
			}
		}
	}
}

func TestPromptReferences(t *testing.T) {
	promptReferences()
}
