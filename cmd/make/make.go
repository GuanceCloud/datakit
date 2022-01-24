// tool to build datakit

package main

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/make/build"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/version"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ip2isp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	_ "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/all"
)

var (
	flagBinary          = flag.String("binary", "", "binary name to build")
	flagName            = flag.String("name", *flagBinary, "same as -binary")
	flagBuildDir        = flag.String("build-dir", "build", "output of build files")
	flagMain            = flag.String("main", `main.go`, `binary build entry`)
	flagDownloadAddr    = flag.String("download-addr", "", "")
	flagPubDir          = flag.String("pub-dir", "pub", "")
	flagArchs           = flag.String("archs", "local", "os archs")
	flagRelease         = flag.String("release", "", `build for local/testing/production`)
	flagPub             = flag.Bool("pub", false, `publish binaries to OSS: local/testing/production`)
	flagBuildISP        = flag.Bool("build-isp", false, "generate ISP data")
	flagDownloadSamples = flag.Bool("download-samples", false, "download samples from OSS to samples/")
	flagDumpSamples     = flag.Bool("dump-samples", false, "download and dump local samples to OSS")

	l = logger.DefaultSLogger("make")
)

func applyFlags() {
	if *flagBuildISP {
		curDir, _ := os.Getwd()

		inputIPDir := filepath.Join(curDir, "china-operator-ip")
		ip2ispFile := filepath.Join(curDir, "pipeline", "ip2isp", "ip2isp.txt")
		if err := os.Remove(ip2ispFile); err != nil {
			l.Warnf("os.Remove: %s, ignored", err.Error())
		}

		if err := ip2isp.MergeIsp(inputIPDir, ip2ispFile); err != nil {
			l.Errorf("MergeIsp failed: %v", err)
		} else {
			l.Infof("merge ip2isp file in `%v`", ip2ispFile)
		}

		inputFile := filepath.Join(curDir, "IP2LOCATION-LITE-DB11.CSV")
		outputFile := filepath.Join(curDir, "pipeline", "ip2isp", "contry_city.yaml")
		if !datakit.FileExist(inputFile) {
			l.Errorf("%v not exist, you can download from `https://lite.ip2location.com/download?id=9`", inputFile)
			os.Exit(0)
		}

		if err := os.Remove(ip2ispFile); err != nil {
			l.Warnf("os.Remove: %s, ignored", err.Error())
		}

		if err := ip2isp.BuildContryCity(inputFile, outputFile); err != nil {
			l.Errorf("BuildContryCity failed: %v", err)
		} else {
			l.Infof("contry and city list in file  `%v`", outputFile)
		}

		os.Exit(0)
	}

	build.AppBin = *flagBinary
	build.BuildDir = *flagBuildDir
	build.PubDir = *flagPubDir
	build.AppName = *flagName
	build.Archs = *flagArchs

	build.MainEntry = *flagMain

	switch *flagRelease {
	case build.ReleaseProduction, build.ReleaseLocal, build.ReleaseTesting:
	default:
		l.Fatalf("invalid release type: %s", *flagRelease)
	}

	build.ReleaseType = *flagRelease

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

	build.DownloadAddr = *flagDownloadAddr
	if !vi.IsStable() {
		build.DownloadAddr = path.Join(*flagDownloadAddr, "rc")

		l.Debugf("under unstable version %s, reset download address to %s",
			build.ReleaseVersion, build.DownloadAddr)
	}

	switch *flagRelease {
	case build.ReleaseProduction:
		l.Debug("under release, only checked inputs released")
		build.InputsReleaseType = "checked"
		if !version.IsValidReleaseVersion(build.ReleaseVersion) {
			l.Fatalf("invalid releaseVersion: %s, expect format 1.2.3-rc0", build.ReleaseVersion)
		}

	default:
		l.Debug("under non-release, all inputs released")
		build.InputsReleaseType = "all"
	}

	l.Infof("use version %s", build.ReleaseVersion)

	if *flagDumpSamples {
		tarPath := "datakit-conf-samples.tar.gz"
		ossPath := "datakit/datakit-conf-samples.tar.gz"
		if err := downloadSamples(ossPath, tarPath); err != nil {
			l.Fatalf("fail to download samples: %v", err)
		}
		if err := extractSamples(tarPath); err != nil {
			l.Fatalf("fail to extract samples: %v", err)
		}
		dirName := getDirName()
		dumpTo := filepath.Join("samples", dirName)
		if err := dumpLocalSamples(dumpTo); err != nil {
			l.Fatalf("fail to dump local samples: %v", err)
		}
		if err := compressSamples("samples", tarPath); err != nil {
			l.Fatalf("fail to compress samples: %v", err)
		}
		if err := uploadSamples(tarPath, ossPath); err != nil {
			l.Fatalf("fail to upload samples: %v", err)
		}
		l.Infof("upload datakit-conf-samples.tar.gz to OSS successfully")
		os.Exit(0)
	}

	if *flagDownloadSamples {
		tarPath := "datakit-conf-samples.tar.gz"
		ossPath := "datakit/datakit-conf-samples.tar.gz"
		if err := downloadSamples(ossPath, tarPath); err != nil {
			l.Fatalf("fail to download samples: %v", err)
		}
		if err := extractSamples(tarPath); err != nil {
			l.Fatalf("fail to extract samples: %v", err)
		}
		l.Infof("download samples from OSS successfully")
		os.Exit(0)
	}
}

func getDirName() string {
	var ret string
	idx := strings.Index(build.ReleaseVersion, "-")
	if idx != -1 {
		ret = build.ReleaseVersion[:idx]
	} else {
		ret = build.ReleaseVersion
	}
	return ret
}

func getOSSClient() (*cliutils.OssCli, error) {
	var ak, sk, bucket, ossHost string

	switch build.ReleaseType {
	case build.ReleaseTesting, build.ReleaseProduction, build.ReleaseLocal:
		tag := strings.ToUpper(build.ReleaseType)
		ak = os.Getenv(tag + "_OSS_ACCESS_KEY")
		sk = os.Getenv(tag + "_OSS_SECRET_KEY")
		bucket = os.Getenv(tag + "_OSS_BUCKET")
		ossHost = os.Getenv(tag + "_OSS_HOST")
	default:
		return nil, fmt.Errorf("unknown release type: %s", build.ReleaseType)
	}

	oc := &cliutils.OssCli{
		Host:       ossHost,
		PartSize:   512 * 1024 * 1024,
		AccessKey:  ak,
		SecretKey:  sk,
		BucketName: bucket,
		WorkDir:    "datakit",
	}
	if err := oc.Init(); err != nil {
		return nil, err
	}
	return oc, nil
}

func downloadSamples(from, to string) error {
	oc, err := getOSSClient()
	if err != nil {
		return err
	}
	if err := oc.Download(from, to); err != nil {
		return fmt.Errorf("fail to download from oss, bucket: %s: %w", oc.BucketName, err)
	}
	return nil
}

// extractSamples extracts samples from given datakit-conf-samples.tar.gz to datakit/samples.
// Samples of current version is skipped because neither --dump-samples nor --download-samples
// (it is used to download samples from oss and then check compatibility) needs samples of current version.
// Besides, samples of current version may change before official release.
func extractSamples(from string) error {
	f, err := os.Open(filepath.Clean(from))
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck,gosec
	reader, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer reader.Close() //nolint:errcheck,gosec
	tr := tar.NewReader(reader)
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		// Skip directories and hidden files.
		if h.FileInfo().IsDir() || strings.HasPrefix(h.FileInfo().Name(), ".") {
			continue
		}
		// Skip current version samples.
		if strings.Contains(h.Name, getDirName()) {
			continue
		}
		path := h.Name
		if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
			return err
		}
		dest, err := os.Create(path)
		if err != nil {
			return err
		}
		//nolint:gosec
		if _, err := io.Copy(dest, tr); err != nil {
			return err
		}
	}
	return nil
}

// dumpLocalSamples dumps config samples to given path.
func dumpLocalSamples(to string) error {
	// Remove existing samples in samplesPath.
	if err := os.RemoveAll(to); err != nil {
		return err
	}
	if err := os.Mkdir(to, os.ModePerm); err != nil {
		return err
	}

	for name, creator := range inputs.Inputs {
		input := creator()
		catalog := input.Catalog()
		catalogPath := filepath.Join(to, catalog)
		// Create catalog directory if not exist.
		if _, err := os.Stat(catalogPath); err != nil {
			if err := os.Mkdir(catalogPath, os.ModePerm); err != nil {
				return err
			}
		}
		f, err := os.Create(filepath.Join(catalogPath, name+".conf"))
		if err != nil {
			return err
		}
		defer f.Close() //nolint:errcheck,gosec
		if _, err := f.WriteString(input.SampleConfig()); err != nil {
			return err
		}
	}
	return nil
}

// compressSamples compresses given samples directory.
func compressSamples(from, to string) error {
	fw, err := os.Create(to)
	if err != nil {
		return err
	}
	defer fw.Close() //nolint:errcheck,gosec
	gw := gzip.NewWriter(fw)
	defer gw.Close() //nolint:errcheck,gosec
	tw := tar.NewWriter(gw)
	defer tw.Close() //nolint:errcheck,gosec
	return filepath.Walk(from, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip directories and hidden files.
		if info.IsDir() || strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}
		fr, err := os.Open(filepath.Clean(path))
		if err != nil {
			return err
		}
		defer fr.Close() //nolint:errcheck,gosec
		if h, err := tar.FileInfoHeader(info, path); err != nil {
			return err
		} else {
			h.Name = path
			if err = tw.WriteHeader(h); err != nil {
				return err
			}
		}
		if _, err := io.Copy(tw, fr); err != nil {
			return err
		}
		return nil
	})
}

// uploadSamples uploads given conf.tar.gz to oss.
func uploadSamples(from, to string) error {
	oc, err := getOSSClient()
	if err != nil {
		return err
	}
	return oc.Upload(from, to)
}

func main() {
	flag.Parse()
	applyFlags()

	if *flagPub {
		build.NotifyStartPub()
		if err := build.PubDatakit(); err != nil {
			l.Error(err)
			build.NotifyFail(err.Error())
		} else {
			build.NotifyPubDone()
		}
	} else {
		build.NotifyStartBuild()
		if err := build.Compile(); err != nil {
			l.Error(err)
			build.NotifyFail(err.Error())
		} else {
			build.NotifyBuildDone()
		}
	}
}
