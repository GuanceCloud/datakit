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
	},
	{
		name: "oceanbase",
		lang: "go",

		entry: "oceanbase.go",
		osarchs: map[string]bool{
			"linux/amd64": true,
			"linux/arm64": true,
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

		switch ex.name {
		case "oceanbase", "db2", "oracle":
			if env := os.Getenv("ENABLE_DOCKER_BUILD_INPUTS"); len(env) == 0 {
				l.Warnf("WARNING: skip build %s because env not specified!", ex.name)
				continue
			}

			str, err := exec.LookPath("docker")
			if err != nil {
				l.Warnf("WARNING: skip build %s because docker is NOT exist!", ex.name)
				continue
			}

			l.Infof("Found docker in %s", str)
		}

		if goarch != runtime.GOARCH {
			switch ex.name { //nolint:gocritic
			case "ebpf":
				l.Warnf("skip, " + ex.name + " does not support cross compilation")
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
				args = append(args, tags)
			}

			entryPath := filepath.Join("internal", "plugins", "externals", ex.name, ex.entry)
			l.Infof("entryPath = %s", entryPath)

			outPath := filepath.Join(outdir, out)
			l.Infof("outPath = %s", outPath)

			moreBuild := []string{
				"-o", outPath,
				"-ldflags",
				"-w -s",
				entryPath,
			}
			args = append(args, moreBuild...)

			envs = append(envs, "GOOS="+goos, "GOARCH="+goarch) //nolint:makezero

			switch ex.name {
			case "oceanbase", "db2", "oracle":
				// x86
				// docker run --rm \
				//   --name $DOCKER_IMAGE_NAME \
				//   -e BUILD_PROJECT=oceanbase \
				//   -e BUILD_ARCH=amd64 \
				//   -e BUILD_SOURCE=internal/plugins/externals/oceanbase/oceanbase.go \
				//   -e BUILD_DEST=dist/datakit-linux-amd64/externals/oceanbase \
				//   -v /root/gopath/src:/tmp/gopath/src \
				//   $DOCKER_IMAGE_NAME:$DOCKER_IMAGE_TAG

				// arm
				//  docker run --rm \
				//    --name $DOCKER_IMAGE_NAME \
				//    -e BUILD_PROJECT=oceanbase \
				//    -e BUILD_ARCH=arm64 \
				//    -e BUILD_SOURCE=internal/plugins/externals/oceanbase/oceanbase.go \
				//    -e BUILD_DEST=dist/datakit-linux-arm64/externals/oceanbase \
				//    -v /root/gopath/src:/tmp/gopath/src \
				//    $DOCKER_IMAGE_NAME:$DOCKER_IMAGE_TAG

				wd, err := os.Getwd()
				if err != nil {
					l.Errorf("os.Getwd() failed: %v", err)
					return err
				}
				l.Infof("current directory: %s", wd)

				projectPrefix := getProjectPrefix(wd)
				if len(projectPrefix) == 0 {
					l.Errorf("projectPrefix emptry")
					return fmt.Errorf("path error")
				}

				distOut := getProjectSuffix(outPath)
				if len(distOut) == 0 {
					l.Errorf("distOut emptry")
					return fmt.Errorf("path error")
				}

				if err := cleanDocker(); err != nil {
					return err
				}

				args := []string{
					"docker", "run", "--rm",
					"--name", "builder-plus",
					"-e", "BUILD_PROJECT=" + ex.name,
					"-e", "BUILD_ARCH=" + goarch,
					"-e", "BUILD_SOURCE=" + entryPath,
					"-e", "BUILD_DEST=" + distOut,
					"-v", projectPrefix + ":" + "/tmp/gopath/src",
					"pubrepo.jiagouyun.com/image-repo-for-testing/builder-plus:1.1",
				}
				envs := []string{}
				msg, err := runEnv(args, envs)
				if err != nil {
					return fmt.Errorf("failed to run %v, envs: %v: %w, msg: %s",
						args, envs, err, string(msg))
				}

			default:
				msg, err := runEnv(args, envs)
				if err != nil {
					return fmt.Errorf("failed to run %v, envs: %v: %w, msg: %s",
						args, envs, err, string(msg))
				}
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
			if len(envs) > 0 {
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

func getProjectPrefix(str string) string {
	nIdx := strings.Index(str, "gitlab.jiagouyun.com/")
	if nIdx == -1 {
		return ""
	}

	return str[:nIdx]
}

func getProjectSuffix(str string) string {
	projectName := "/datakit/"
	nIdx := strings.Index(str, projectName)
	if nIdx == -1 {
		return ""
	}

	return str[nIdx+len(projectName):]
}

func cleanDocker() error {
	stopDocker()

	rmDocker()

	return nil
}

func stopDocker() {
	args := []string{
		"docker", "stop", "builder-plus",
	}

	msg, err := runEnv(args, os.Environ())
	if err != nil && !strings.Contains(string(msg), "No such container") {
		l.Info(string(msg))
		l.Warnf("stop docker failed: %v", err)
	}
}

func rmDocker() {
	args := []string{
		"docker", "rm", "builder-plus",
	}

	msg, err := runEnv(args, os.Environ())
	if err != nil && !strings.Contains(string(msg), "No such container") {
		l.Info(string(msg))
		l.Warnf("rm docker failed: %v", err)
	}
}
