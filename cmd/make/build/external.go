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
			"linux/arm64": true,
		},

		buildArgs: nil,
		envs: []string{
			"CGO_ENABLED=1",
		},
	},
	{
		// requirement: apt-get install gcc-multilib
		name: "db2",
		lang: "go",

		entry: "db2.go",
		osarchs: map[string]bool{
			"linux/amd64": true,
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
			"linux/arm64": true,
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
		// NEVER using ex.envs for appending,
		//       it would be modified and poisoned in the future use.
		envs := make([]string, len(ex.envs))
		buildArgs := make([]string, len(ex.buildArgs))
		copy(envs, ex.envs)
		copy(buildArgs, ex.buildArgs)

		var tags string

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

		if ex.name == "db2" {
			// "CGO_CFLAGS=-I/opt/ibm/clidriver/include",
			// "CGO_LDFLAGS=-L/opt/ibm/clidriver/lib",
			// "LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/opt/ibm/clidriver/lib",

			envs = append(envs, "CGO_CFLAGS=-I"+os.Getenv("IBM_CLI_DRIVER")+"/include") //nolint:makezero
			envs = append(envs, "CGO_LDFLAGS=-L"+os.Getenv("IBM_CLI_DRIVER")+"/lib")    //nolint:makezero

			ldLibraryFullPath := os.Getenv("IBM_CLI_DRIVER") + "/lib"
			ldLibraryPath := os.Getenv("LD_LIBRARY_PATH")
			if len(ldLibraryPath) > 0 {
				ldLibraryFullPath = ldLibraryPath + ":" + ldLibraryFullPath
			}

			envs = append(envs, "LD_LIBRARY_PATH="+ldLibraryFullPath) //nolint:makezero

			tags = "db2"
		}

		if goarch != runtime.GOARCH {
			switch ex.name {
			case "ebpf":
				l.Warnf("skip, " + ex.name + " does not support cross compilation")
				continue
			case "oracle":
				l.Infof("building " + ex.name + " by cross compilation...")

				envs = append(envs, "CC="+os.Getenv("CROSS_GCC"))  //nolint:makezero
				envs = append(envs, "CXX="+os.Getenv("CROSS_GPP")) //nolint:makezero
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

		l.Info("lang = ", ex.lang)
		switch strings.ToLower(ex.lang) {
		case "go", "golang":

			switch osarch {
			case "windows/amd64", "windows/386":
				out += ".exe"
			default: // pass
			}

			args := []string{"go", "build"}
			if len(tags) > 0 {
				args = append(args, "-tags")
				args = append(args, "tags")
			}

			moreBuild := []string{
				"-o", filepath.Join(outdir, out),
				"-ldflags",
				"-w -s",
				filepath.Join("internal", "plugins", "externals", ex.name, ex.entry),
			}
			args = append(args, moreBuild...)

			envs = append(envs, "GOOS="+goos, "GOARCH="+goarch) //nolint:makezero

			msg, err := runEnv(args, envs)
			if err != nil {
				return fmt.Errorf("failed to run %v, envs: %v: %w, msg: %s",
					args, envs, err, string(msg))
			}

		case "makefile", "Makefile":
			args := []string{
				"make",
				"--file=" + filepath.Join("internal", "plugins", "externals", ex.name, ex.entry),
				"SRCPATH=" + "internal/plugins/externals/" + ex.name,
				"OUTPATH=" + filepath.Join(outdir, out),
				"ARCH=" + runtime.GOARCH,
			}

			envs = append(envs, "GOOS="+goos, "GOARCH="+goarch) //nolint:makezero
			msg, err := runEnv(args, envs)
			if err != nil {
				return fmt.Errorf("failed to run %v, envs: %v: %w, msg: %s",
					args, envs, err, string(msg))
			}

		default: // for python, just copy source code into build dir
			buildArgs = append(buildArgs, filepath.Join(outdir, "externals")) //nolint:makezero
			cmd := exec.Command(ex.buildCmd, buildArgs...)                    //nolint:gosec
			if envs != nil {
				cmd.Env = append(os.Environ(), envs...)
			}

			res, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("failed to build python(%s %s): %s, err: %w",
					ex.buildCmd, strings.Join(buildArgs, " "), res, err)
			}
		}
	}

	return nil
}
