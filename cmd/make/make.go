// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package main

import (
	"flag"
	"os"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/make/build"
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/all"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/version"
)

func init() { //nolint:gochecknoinits
	flag.StringVar(&build.AppBin, "binary", build.AppBin, "binary name to build")
	flag.StringVar(&build.AppName, "name", build.AppName, "same as -binary")
	flag.StringVar(&build.BuildDir, "build-dir", "build", "output of build files")
	flag.StringVar(&build.MainEntry, "main", "", "binary build entry")
	flag.StringVar(&build.UploadAddr, "upload-addr", "", "dist where to upload to")
	flag.StringVar(&build.DownloadCDN, "download-cdn", "", "dist where to download from")
	flag.StringVar(&build.PubDir, "pub-dir", "pub", "")
	flag.StringVar(&build.Archs, "archs", "local", "os archs")
	flag.StringVar(&raceDetection, "race", "off", "enable race deteciton")
	flag.StringVar(&build.ReleaseType, "release", "", "build for local/testing/production")
	flag.StringVar(&dwURL, "dataway-url", "", "set dataway URL(https://dataway.com/v1/write/logging?token=xxx) to push testing metrics")

	flag.BoolVar(&build.NotifyOnly, "notify-only", false, "notify CI process")
	flag.BoolVar(&doPub, "pub", false, `publish binaries to OSS: local/testing/production`)
	flag.BoolVar(&doPubeBPF, "pub-ebpf", false, `publish datakit-ebpf to OSS: local/testing/production`)
	flag.BoolVar(&pkgeBPF, "pkg-ebpf", false, `add datakit-ebpf to datakit tarball`)
	flag.BoolVar(&buildISP, "build-isp", false, "generate ISP data")

	flag.BoolVar(&ut, "ut", false, "test all DataKit code")
	flag.BoolVar(&it, "it", false, "test only DataKit integration testing code")
	flag.StringVar(&build.UTExclude, "ut-exclude", "", "exclude packages for testing")

	flag.BoolVar(&downloadSamples, "download-samples", false, "download samples from OSS to samples/")
	flag.BoolVar(&dumpSamples, "dump-samples", false, "download and dump local samples to OSS")
	flag.StringVar(&build.MarkdownCheck, "mdcheck", "", "check markdown docs")
	flag.StringVar(&build.Autofix, "mdcheck-autofix", "off", "check markdown docs with autofix")
	flag.StringVar(&build.MetaDir, "meta-dir", "", "metadir used to check markdown meta")
}

var (
	doPub         = false
	doPubeBPF     = false
	pkgeBPF       = false
	buildISP      = false
	ut            = false
	it            = false
	raceDetection = "off"
	dwURL         = "not-set"

	downloadSamples = false
	dumpSamples     = false

	l = logger.DefaultSLogger("make")
)

func applyFlags() {
	if build.MarkdownCheck != "" {
		// NOTE: if nothing matched, then we think all docs are valid, but exit -1
		// to co-work with Makefile's shell, see Makefile's entry 'check_man'
		//
		// Here we keep the exit code the same with grep:
		//
		//  	If find something, exit ok, or exit fail.
		//
		if n := build.Match(build.MarkdownCheck); n == 0 {
			cp.Infof("mdcheck format ok.\n")
			os.Exit(0)
		} else {
			cp.Errorf("[E] Got %d invalid docs/erros during markdown checking(Unicode/ASCII/meta).\n"+
				"    See https://docs.guance.com/datakit/mkdocs-howto/#zh-en-mix\n", n)
			os.Exit(-1)
		}
	}

	l.Infof("download-cdn: %s", build.DownloadCDN)

	if it {
		if err := build.IntegrationTestingDataKit(); err != nil {
			l.Errorf("build.IntegrationTestingDataKit: %s", err)
			os.Exit(-1)
		}
		return
	}

	if ut {
		testutils.DatawayURL = dwURL

		if err := build.UnitTestDataKit(); err != nil {
			l.Errorf("build.UnitTestDataKit: %s", err)
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
		l.Fatalf("invalid release type: %s", build.ReleaseType)
	}

	// override git.Version
	if x := os.Getenv("VERSION"); x != "" {
		build.ReleaseVersion = x
	}

	if x := os.Getenv("DINGDING_TOKEN"); x != "" {
		build.NotifyToken = x
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

	if dumpSamples {
		build.DumpSamples()

		l.Infof("upload datakit-conf-samples.tar.gz to OSS successfully")
		return
	}

	if downloadSamples {
		build.DownloadSamples()
		l.Infof("download samples from OSS successfully")
		return
	}

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
		build.NotifyStartPub()
		if pkgeBPF {
			build.PackageeBPF()
		}

		if err := build.PubDatakit(); err != nil {
			l.Error(err)
			build.NotifyFail(err.Error())
		} else {
			build.NotifyPubDone()
		}
		return
	} else {
		if err := build.Compile(); err != nil {
			l.Error(err)
			build.NotifyFail(err.Error())
		} else {
			build.NotifyBuildDone()
		}
		return
	}
}

func main() {
	flag.Parse()
	applyFlags()
}
