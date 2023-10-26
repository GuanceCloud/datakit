package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

func gatherGoVer() []string {
	var li []string
	for i := 9; i <= 21; i++ {
		li = append(li,
			"1."+strconv.Itoa(i))
	}
	return li
}

type cmdLine struct {
	binname string
	args    []string
	arch    string
}

func nocgocmd(ver []string, cgoEnabled bool) []cmdLine {
	dir, err := os.Getwd()
	if err != nil {
		return nil
	}

	var cgoStr string
	if cgoEnabled {
		cgoStr = "cgo"
	} else {
		cgoStr = "nocgo"
	}

	var r []cmdLine

	for _, v := range ver {
		r = append(r, cmdLine{
			binname: "docker",
			args: []string{
				"run", "--platform", "arm64", "--rm", "-v", dir + "/" + cgoStr + "/:/usr/src/" + cgoStr,
				"-w", "/usr/src/" + cgoStr, "golang:" + v, "go", "build", "-o", "dist/gobin" + ".arm64.go" + v, "gobin.go",
			},
			arch: "arm64",
		},
		)

		r = append(r, cmdLine{
			binname: "docker",
			args: []string{
				"run", "--platform", "arm64", "--rm", "-v", dir + "/" + cgoStr + "/:/usr/src/" + cgoStr,
				"-w", "/usr/src/" + cgoStr, "golang:" + v, "go", "build", "-o", "dist/gobin" + ".amd64.go" + v, "gobin.go",
			},
			arch: "amd64",
		},
		)
	}

	return r
}

func main() {
	var err error
	cmd := exec.Command("docker", "run", "--privileged", "--rm", "tonistiigi/binfmt", "--install", "all")

	b, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Print(string(b))

	// cmdLi := nocgocmd(gatherGoVer(), false)
	// for _, v := range cmdLi {
	// 	cmd := exec.Command(v.binname, v.args...)
	// 	cmd.Env = append(cmd.Env, "GOOS=linux", "GOARCH="+v.arch)
	// 	b, err = cmd.CombinedOutput()
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// 	fmt.Print(string(b))
	// }

	cmdLi := nocgocmd(gatherGoVer(), true)
	for _, v := range cmdLi {
		cmd := exec.Command(v.binname, v.args...) //nolint:gosec
		cmd.Env = append(cmd.Env, "GOOS=linux", "GOARCH="+v.arch)
		b, err = cmd.CombinedOutput()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Print(string(b))
	}
}
