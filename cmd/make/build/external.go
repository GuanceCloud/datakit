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

var (
	externals = []*dkexternal{
		&dkexternal{
			// requirement: apt-get install gcc-multilib
			name: "oraclemonitor",
			lang: "go",

			entry: "main.go",
			osarchs: map[string]bool{
				"linux/amd64": true,
				"linux/386":   true,
			},

			buildArgs: nil,
			envs: []string{
				"CGO_ENABLED=1",
			},
		},

		&dkexternal{
			name: "csv",
			osarchs: map[string]bool{
				`linux/386`:     true,
				`linux/amd64`:   true,
				`linux/arm`:     true,
				`linux/arm64`:   true,
				`darwin/amd64`:  true,
				`windows/amd64`: true,
				`windows/386`:   true,
			},
			buildArgs: []string{"plugins/externals/csv/build.sh"},
			buildCmd:  "bash",
		},
		&dkexternal{
			name: "ansible",
			osarchs: map[string]bool{
				`linux/386`:     true,
				`linux/amd64`:   true,
				`linux/arm`:     true,
				`linux/arm64`:   true,
				`darwin/amd64`:  true,
				`windows/amd64`: true,
				`windows/386`:   true,
			},
			buildArgs: []string{"plugins/externals/ansible/build.sh"},
			buildCmd:  "bash",
		},

		&dkexternal{
			// requirement: apt-get install gcc-multilib
			name: "skywalkingGrpcV3",
			lang: "go",

			entry: "main.go",
			osarchs: map[string]bool{
				`linux/386`:     true,
				`linux/amd64`:   true,
				`linux/arm`:     true,
				`linux/arm64`:   true,
				`darwin/amd64`:  true,
				`windows/amd64`: true,
				`windows/386`:   true,
			},

			buildArgs: nil,
			envs: []string{
				"CGO_ENABLED=0",
			},
		},

		// others...
	}
)

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
				out = out + ".exe"
			default: // pass
			}

			args := []string{
				"go", "build",
				"-o", filepath.Join(outdir, "externals", out),
				"-ldflags",
				"-w -s",
				filepath.Join("plugins/externals", ex.name, ex.entry),
			}

			env := append(ex.envs, "GOOS="+goos, "GOARCH="+goarch)

			msg, err := runEnv(args, env)
			if err != nil {
				l.Fatalf("failed to run %v, envs: %v: %v, msg: %s", args, env, err, string(msg))
			}

		default: // for python, just copy source code into build dir
			args := append(ex.buildArgs, filepath.Join(outdir, "externals"))
			cmd := exec.Command(ex.buildCmd, args...)
			if ex.envs != nil {
				cmd.Env = append(os.Environ(), ex.envs...)
			}

			res, err := cmd.CombinedOutput()
			if err != nil {
				l.Fatalf("failed to build python(%s %s): %s, err: %s", ex.buildCmd, strings.Join(args, " "), res, err.Error())
			}
		}
	}
}
