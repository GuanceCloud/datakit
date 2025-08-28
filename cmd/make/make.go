// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package main

import (
	"flag"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/mdcheck/check"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/make/build"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	mexport "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/all"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/version"
)

var (
	mdNoAutofix,
	mdCheck,
	mdMetaDir string

	mdNoSectionCheck = false
	sampleConfCheck  = false
	doPub            = false
	doPubeBPF        = false
	pkgEBPF          = 0
	downloadEBPF     = 0
	buildISP         = false
	ut               = false
	dca              = false
	export           = false
	dwURL            = "not-set"
	mdSkip           = ""
	logLevel         = "info"

	l = logger.DefaultSLogger("make")
)

func init() { //nolint:gochecknoinits
	flag.StringVar(&build.AppBin, "binary", build.AppBin, "binary name to build")
	flag.StringVar(&build.AppName, "name", build.AppName, "same as -binary")
	flag.StringVar(&build.MainEntry, "main", "", "binary build entry")

	flag.StringVar(&build.DistDir, "dist-dir", "pub", "")
	flag.StringVar(&build.Archs, "archs", "local", "os archs")
	flag.StringVar(&build.AWSRegions, "aws-regions", "cn-north-1,cn-northwest-1", "aws regions")
	flag.BoolVar(&build.EnableUploadAWS, "enable-upload-aws", false, "enable upload aws")
	flag.StringVar(&build.RaceDetection, "race", "off", "enable race deteciton")
	flag.StringVar(&build.ReleaseType, "release", "", "build for local/testing/production")

	flag.StringVar(&build.Brand, "brand", build.ValueNotSet, "brand URL")

	flag.StringVar(&dwURL, "dataway-url", "", "set dataway URL(https://dataway.com/v1/write/logging?token=xxx) to push testing metrics")

	flag.BoolVar(&build.NotifyOnly, "notify-only", false, "notify CI process")
	flag.BoolVar(&doPub, "pub", false, `publish binaries to OSS: local/testing/production`)
	flag.BoolVar(&doPubeBPF, "pub-ebpf", false, `publish datakit-ebpf to OSS: local/testing/production`)

	flag.IntVar(&pkgEBPF, "pkg-ebpf", 0, `add datakit-ebpf to datakit tar.gz`)
	flag.IntVar(&downloadEBPF, "download-ebpf", 0, `download datakit-ebpf from OSS: local/testing/production`)

	flag.BoolVar(&buildISP, "build-isp", false, "generate ISP data")

	flag.BoolVar(&ut, "ut", false, "test all DataKit code")
	flag.IntVar(&build.Parallel, "ut-parallel", runtime.NumCPU(), "specify concurrent worker on unit testing")
	flag.StringVar(&build.UTExclude, "ut-exclude", "", "exclude packages for testing")
	flag.StringVar(&build.UTOnly, "ut-only", "", "select packages for testing")
	flag.IntVar(&build.OnlyExternalInputs, "only-external-inputs", 0, "only build external inputs")

	flag.StringVar(&build.HelmChartDir, "helm-chart-dir", build.ValueNotSet, "set Helm workdir")
	flag.IntVar(&build.SkipHelm, "skip-helm", 0, "skip build helm package")

	flag.StringVar(&build.DockerImageRepo, "docker-image-repo", build.ValueNotSet, "set docker image repo URL")
	flag.StringVar(&mdCheck, "mdcheck", "", "check markdown docs")
	flag.BoolVar(&sampleConfCheck, "sample-conf-check", false, "check input's sample conf")
	flag.StringVar(&mdNoAutofix, "mdcheck-no-autofix", "", "check markdown docs with autofix")
	flag.BoolVar(&mdNoSectionCheck, "mdcheck-no-section-check", false, "do not check markdown sections")
	flag.StringVar(&mdMetaDir, "meta-dir", "", "metadir used to check markdown meta")

	flag.BoolVar(&dca, "dca", false, "build DCA only")
	flag.StringVar(&build.DCAVersion, "dca-version", build.ValueNotSet, "specify DCA version string")

	//
	// export related flags.
	//
	flag.BoolVar(&export, "export", false, "Export used to output all resource related to Datakit.")

	flag.StringVar(&build.ExportDocDir, "export-doc-dir", "", "export all inputs and related docs to specified path")
	flag.StringVar(&build.ExportIntegrationDir, "export-integration-dir", "", "export all integration related resource to specified path")

	flag.StringVar(&build.ExportIgnore, "ignore", "", "disable list, i.e., --ignore nginx,redis,mem")
	flag.StringVar(&build.ExportTODO, "TODO", "TODO", "set TODO placeholder")
	flag.StringVar(&build.ExportVersion, "version", datakit.Version, "set DataKit version string in related documents")

	flag.StringVar(&logLevel, "log-level", "info", "set log level of building log")
}

func applyFlags() {
	build.LoadENVs()

	// set global log root
	lopt := &logger.Option{
		Level: strings.ToLower(logLevel),
		Flags: logger.OPT_DEFAULT,
	}

	if err := logger.InitRoot(lopt); err != nil {
		l.Errorf("set root log(options: %+#v) failed: %s", lopt, err.Error())
	} else {
		l.Infof("set root logger(options: %+#v)ok", lopt)
		l = logger.SLogger("make")
		build.SetLog()
		mexport.SetLog()
	}

	if mdCheck != "" {
		skips := strings.Split(mdSkip, ",")
		l.Infof("skip files %+#v", skips)

		autofix := false
		if x, err := strconv.ParseBool(mdNoAutofix); err == nil {
			autofix = x
		}

		res, err := check.Check(
			check.WithMarkdownDir(mdCheck),
			check.WithMetaDir(mdMetaDir),
			check.WithAutofix(autofix),
			check.WithExcludeFiles(skips...),
			check.WithCheckSection(!mdNoSectionCheck),
		)
		if err != nil {
			l.Errorf("markdown check: %s", err.Error())
			os.Exit(-1)
		}

		for _, r := range res {
			switch {
			case r.Err != "":
				l.Errorf("%s: %q | Err: %s", r.Path, r.Text, r.Err)
			case r.Warn != "":
				l.Warnf("%s: Warn: %s", r.Path, r.Err)
			}
		}

		if len(res) > 0 {
			os.Exit(-1)
		}

		return
	}

	if sampleConfCheck {
		if err := build.CheckSampleConf(inputs.AllInputs); err != nil {
			l.Errorf("sample conf check: %s", err.Error())
			os.Exit(-1)
		}

		return
	}

	l.Infof("download-cdn: %s", build.DownloadCDN)

	if ut {
		testutils.DatawayURL = dwURL

		if err := build.UnitTestDataKit(); err != nil {
			l.Errorf("build.UnitTestDataKit: %s", err)
			os.Exit(-1)
		}
		return
	}

	if dca {
		if err := build.CompileDCA(); err != nil {
			l.Error(err)
			os.Exit(-1)
		}
		return
	}

	if export {
		if err := build.BuildExport(); err != nil {
			l.Errorf("build.BuildExport: %s", err)
			os.Exit(-1)
		}
		return
	}

	if buildISP {
		build.BuildISP()
		return
	}

	switch build.ReleaseType {
	case build.ReleaseProduction, build.ReleaseLocal, build.ReleaseTesting:
	default:
		l.Errorf("invalid release type: %q", build.ReleaseType)
	}

	// override git.Version
	if x := os.Getenv("VERSION"); x != "" {
		build.ReleaseVersion = x
	}

	vi := version.VerInfo{VersionString: build.ReleaseVersion}
	if err := vi.Parse(); err != nil {
		l.Fatalf("invalid version %s", build.ReleaseVersion)
	}

	switch build.ReleaseType {
	case build.ReleaseProduction:
		l.Debug("under release, only checked inputs released")
		build.InputsReleaseType = "checked"
		if !version.IsValidReleaseVersion(build.ReleaseVersion) {
			l.Fatalf("invalid releaseVersion: %s, expect format 1.2.3", build.ReleaseVersion)
		}

	default:
		l.Debug("under non-release, all inputs released")
		build.InputsReleaseType = "all"
	}

	l.Infof("use version %s", build.ReleaseVersion)

	if build.NotifyOnly {
		build.NotifyStartBuild()
		return
	}

	if doPubeBPF {
		build.NotifyStartPubEBpf()
		if err := build.PubDatakitEBpf(); err != nil {
			l.Errorf("build.PubDatakiteBPF: %s", err)
			build.NotifyFail(err.Error())
		} else {
			build.NotifyPubEBpfDone()
		}
		return
	}

	if doPub {
		l.Infof("under publishing...")
		build.NotifyStartPub()
		if downloadEBPF != 0 {
			if err := build.PackageEBPF(); err != nil {
				l.Errorf("build.PackageeBPF: %s", err)
				return
			}
		}

		if err := build.PubDatakit(); err != nil {
			l.Errorf("build.PubDatakit: %s", err)
			build.NotifyFail(err.Error())
		} else {
			build.NotifyPubDone()
		}
		return
	} else {
		l.Infof("under compiling...")
		if err := build.Compile(); err != nil {
			l.Errorf("build.Compile: %s", err)
			build.NotifyFail(err.Error())
		} else if pkgEBPF != 0 {
			if err := build.PackageEBPF(); err != nil {
				l.Errorf("build.PackageeBPF: %s", err)
				return
			}
		}
		return
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()
	applyFlags()
}
