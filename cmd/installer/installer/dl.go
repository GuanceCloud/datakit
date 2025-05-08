// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package installer implements datakit's install and upgrade
package installer

import (
	"crypto/tls"
	"net/url"
	"runtime"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/upgrader/upgrader"
	apminj "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/apminject/utils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmds"
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	dl "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/downloader"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpcli"
)

func (args *InstallerArgs) DownloadFiles(to string) error {
	dl.CurDownloading = "datakit"

	cliopt := &httpcli.Options{
		// ignore SSL error
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint
	}

	if args.Proxy != "" {
		u, err := url.Parse(args.Proxy)
		if err != nil {
			return err
		}
		cliopt.ProxyURL = u
		l.Infof("set proxy to %s ok", args.Proxy)
	}

	cli := httpcli.Cli(cliopt)

	dkURL := args.DistDatakitURL
	if args.IsLite {
		dkURL = args.DistDatakitLiteURL
	} else if args.IsELinker {
		dkURL = args.DistDatakitELinkerURL
	}

	cp.Infof("Downloading %s => %s\n", dkURL, to)
	if err := dl.Download(cli, dkURL, to, true, args.FlagDownloadOnly); err != nil {
		return err
	}
	cp.Printf("\n")

	dl.CurDownloading = "data"

	cp.Infof("Downloading %s => %s\n", args.DistDataURL, to)
	if err := dl.Download(cli, args.DistDataURL, to, true, args.FlagDownloadOnly); err != nil {
		return err
	}

	// We will not upgrade dk-upgrader default when upgrading Datakit except for setting flagUpgradeManagerService flag
	if !args.FlagDKUpgrade || (args.FlagDKUpgrade && args.FlagUpgraderEnabled == 1) || args.FlagDownloadOnly {
		toUpgrader := to
		if !args.FlagDownloadOnly {
			toUpgrader = upgrader.InstallDir
		}
		dl.CurDownloading = upgrader.BuildBinName
		cp.Infof("Downloading %s => %s\n", args.DistDkUpgraderURL, toUpgrader)
		if err := dl.Download(cli, args.DistDkUpgraderURL, toUpgrader, true, args.FlagDownloadOnly); err != nil {
			l.Warnf("unable to download %s from [%s]: %s", upgrader.BuildBinName, args.DistDkUpgraderURL, err)
		}
	}

	if runtime.GOOS == "linux" &&
		(runtime.GOARCH == "amd64" || runtime.GOARCH == "arm64") {
		opts := []apminj.Opt{
			apminj.WithLauncherURL(cli, args.DistDatakitAPMInjectURL),
			apminj.WithInstallDir(to),
			apminj.WithInstrumentationEnabled(args.InstrumentationEnabled),
		}
		if args.InstrumentationEnabled != "" {
			opts = append(opts,
				apminj.WithJavaLibURL(args.DistDatakitAPMInjJavaLibURL),
				apminj.WithPythonLib(true))
		}

		if err := apminj.Download(l, opts...); err != nil {
			l.Warnf("download apm inject failed: %s", err.Error())
		} else {
			config.Cfg.APMInject.InstrumentationEnabled = args.InstrumentationEnabled
		}
	}

	if args.IPDBType != "" {
		cp.Printf("\n")
		baseURL := "https://" + args.DataKitBaseURL
		if args.DistBaseURL != "" {
			baseURL = "https://" + args.DistBaseURL
		}

		l.Debugf("get ipdb from %s", baseURL)
		if _, err := cmds.InstallIPDB(baseURL, args.IPDBType); err != nil {
			l.Warnf("install IPDB %s failed error: %s, please try later.", args.IPDBType, err.Error())
			time.Sleep(1 * time.Second)
		} else {
			config.Cfg.Pipeline.IPdbType = args.IPDBType
		}
	}

	cp.Printf("\n")
	return nil
}
