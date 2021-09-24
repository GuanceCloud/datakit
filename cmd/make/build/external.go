package build

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type dkexternal struct {
	name string

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
		name: "net_ebpf",
		lang: "makefile",

		entry: "Makefile",
		osarchs: map[string]bool{
			"linux/amd64": true,
		},

		buildArgs: nil,
		envs: []string{
			"CGO_ENABLED=1",
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

func buildExternals(outdir, goos, goarch string) {
	curOSArch := runtime.GOOS + "/" + runtime.GOARCH

	for _, ex := range externals {
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

		out := ex.name

		switch strings.ToLower(ex.lang) {
		case "go", "golang":

			switch osarch {
			case "windows/amd64", "windows/386":
				out += ".exe"
			default: // pass
			}

			args := []string{
				"go", "build",
				"-o", filepath.Join(outdir, "externals", out),
				"-ldflags",
				"-w -s",
				filepath.Join("plugins", "externals", ex.name, ex.entry),
			}

			ex.envs = append(ex.envs, "GOOS="+goos, "GOARCH="+goarch)

			msg, err := runEnv(args, ex.envs)
			if err != nil {
				l.Fatalf("failed to run %v, envs: %v: %v, msg: %s",
					args, ex.envs, err, string(msg))
			}
		case "makefile", "Makefile":
			args := []string{
				"make",
				"--file=" + filepath.Join("plugins", "externals", ex.name, ex.entry),
				"OUTPATH=" + filepath.Join(outdir, "externals", out),
				"BASEPATH=" + "plugins/externals/" + ex.name,
			}

			ex.envs = append(ex.envs, "GOOS="+goos, "GOARCH="+goarch)
			msg, err := runEnv(args, ex.envs)
			if err != nil {
				l.Fatalf("failed to run %v, envs: %v: %v, msg: %s",
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
				l.Fatalf("failed to build python(%s %s): %s, err: %s",
					ex.buildCmd, strings.Join(ex.buildArgs, " "), res, err.Error())
			}
		}
	}
}
