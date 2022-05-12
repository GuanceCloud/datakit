// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type dkexternal struct {
	name       string
	out        string
	standalone bool

	lang string // go/others

	entry     string
	buildArgs []string

	osarchs map[string]bool
	envs    []string

	buildCmd string
}

var externals = []*dkexternal{
	{
		// requirement: apt-get install gcc-multilib
		name: "oracle",
		lang: "go",

		entry: "oracle.go",
		osarchs: map[string]bool{
			"linux/amd64": true,
			"linux/386":   true,
		},

		buildArgs: nil,
		envs: []string{
			"CGO_ENABLED=1",
		},
	},
	{
		// requirement: apt install clang llvm linux-headers-$(uname -r)
		name:       "ebpf",
		out:        "datakit-ebpf",
		standalone: true,
		lang:       "makefile",

		entry: "Makefile",
		osarchs: map[string]bool{
			"linux/amd64": true,
		},

		buildArgs: nil,
		envs: []string{
			"CGO_ENABLED=1",
		},
	},
	{
		name: "logfwd",
		lang: "go",

		entry: "logfwd.go",
		osarchs: map[string]bool{
			"linux/amd64": true,
			"linux/arm64": true,
		},

		buildArgs: nil,
		envs: []string{
			"CGO_ENABLED=0",
		},
	},
	// &dkexternal{
	// 	// requirement: apt-get install gcc-multilib
	// 	name: "skywalkingGrpcV3",
	// 	lang: "go",

	// 	entry: "main.go",
	// 	osarchs: map[string]bool{
	// 		`linux/386`:     true,
	// 		`linux/amd64`:   true,
	// 		`linux/arm`:     true,
	// 		`linux/arm64`:   true,
	// 		`darwin/amd64`:  true,
	// 		`windows/amd64`: true,
	// 		`windows/386`:   true,
	// 	},

	// 	buildArgs: nil,
	// 	envs: []string{
	// 		"CGO_ENABLED=0",
	// 	},
	// },

	// others...
}

func buildExternals(dir, goos, goarch string, standalone bool) error {
	curOSArch := runtime.GOOS + "/" + runtime.GOARCH
	for _, ex := range externals {
		if ex.standalone != standalone {
			continue
		}
		l.Debugf("building %s-%s/%s", goos, goarch, ex.name)

		if _, ok := ex.osarchs[curOSArch]; !ok {
			l.Warnf("skip build %s under %s", ex.name, curOSArch)
			continue
		}

		osarch := goos + "/" + goarch
		if _, ok := ex.osarchs[osarch]; !ok {
			l.Warnf("skip build %s under %s", ex.name, osarch)
			continue
		}

		if ex.name == "ebpf" {
			if goarch != runtime.GOARCH {
				l.Warnf("skip, ebpf does not support cross compilation")
				continue
			}
		}

		out := ex.name
		if ex.out != "" {
			out = ex.out
		}

		var outdir string
		if ex.standalone {
			outdir = filepath.Join(dir, "standalone", fmt.Sprintf("%s-%s-%s", out, goos, goarch))
		} else {
			outdir = filepath.Join(dir, "externals")
		}

		switch strings.ToLower(ex.lang) {
		case "go", "golang":

			switch osarch {
			case "windows/amd64", "windows/386":
				out += ".exe"
			default: // pass
			}

			args := []string{
				"go", "build",
				"-o", filepath.Join(outdir, out),
				"-ldflags",
				"-w -s",
				filepath.Join("plugins", "externals", ex.name, ex.entry),
			}

			ex.envs = append(ex.envs, "GOOS="+goos, "GOARCH="+goarch)

			msg, err := runEnv(args, ex.envs)
			if err != nil {
				return fmt.Errorf("failed to run %v, envs: %v: %w, msg: %s",
					args, ex.envs, err, string(msg))
			}
		case "makefile", "Makefile":
			args := []string{
				"make",
				"--file=" + filepath.Join("plugins", "externals", ex.name, ex.entry),
				"SRCPATH=" + "plugins/externals/" + ex.name,
				"OUTPATH=" + filepath.Join(outdir, out),
				"ARCH=" + runtime.GOARCH,
			}

			ex.envs = append(ex.envs, "GOOS="+goos, "GOARCH="+goarch)
			msg, err := runEnv(args, ex.envs)
			if err != nil {
				return fmt.Errorf("failed to run %v, envs: %v: %w, msg: %s",
					args, ex.envs, err, string(msg))
			}
		default: // for python, just copy source code into build dir
			ex.buildArgs = append(ex.buildArgs, filepath.Join(outdir, "externals"))
			cmd := exec.Command(ex.buildCmd, ex.buildArgs...) //nolint:gosec
			if ex.envs != nil {
				cmd.Env = append(os.Environ(), ex.envs...)
			}

			res, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("failed to build python(%s %s): %s, err: %w",
					ex.buildCmd, strings.Join(ex.buildArgs, " "), res, err)
			}
		}
	}

	return nil
}
