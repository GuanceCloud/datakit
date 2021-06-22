package cmds

import (
	"bytes"
	"fmt"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/all"
	"testing"
)

var expects = []string{
	"redis",
	"mysql",
	"mongodb",
}

func TestBuildK8sConfig(t *testing.T) {
	BuildK8sConfig("datakit-k8s-deploy", "/Users/liushaobo/go/src/gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/datakit/cmds", true)
}

func TestRunBuild(t *testing.T) {
	k := &KubeDeploy{
		DeployName: "deploy.yaml",
		Inputs:     make(map[string]string),
	}

	var testCases = []struct {
		in     string
		expect string
		fail   bool
	}{
		{
			in:     "Q",
			expect: "Bye!",
		},
		{
			in:     "q",
			expect: "Bye!",
		},
		{
			in:     "exit",
			expect: "Bye!",
		},
		{
			in:     "render",
			expect: "start render kubernetes deploy config...",
		},
	}

	for _, item := range testCases {
		buf := bytes.Buffer{}
		k.runCmd(&buf, item.in)
		fmt.Println("hello world!")
		got := buf.String()

		if got != item.expect {
			t.Errorf("got '%s' want '%s'", got, item.expect)
		}
	}
}

func TestBuildConfig(t *testing.T) {
	k := &KubeDeploy{
		DeployName: "deploy.yaml",
		Inputs:     make(map[string]string),
	}

	for _, expect := range expects {
		if err := k.buildConfig(expect, &Option{}); err != nil {
			t.Errorf("got error %v", err)
			return
		}
	}

	for k, cfg := range k.Inputs {
		t.Log(k, "input config", cfg)
	}

	t.Log("ok")
}

func TestRender(t *testing.T) {
	k := &KubeDeploy{
		DeployName: "deploy",
		Inputs:     make(map[string]string),
	}

	for _, expect := range expects {
		if err := k.buildConfig(expect, &Option{}); err != nil {
			t.Errorf("got error %v", err)
			return
		}
	}

	if err := k.render(); err != nil {
		t.Errorf("render error %v", err)
		return
	}

	t.Log("ok")
}
