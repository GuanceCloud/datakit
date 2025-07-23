// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package build

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"

	humanize "github.com/dustin/go-humanize"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/upgrader/upgrader"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type versionDesc struct {
	Version  string `json:"version"`
	Date     string `json:"date_utc"`
	Uploader string `json:"uploader"`
	Branch   string `json:"branch"`
	Commit   string `json:"commit"`
	Go       string `json:"go"`
}

type tarFileOpt uint32

const (
	// Option to include version information in filename.
	tarReleaseVerMask tarFileOpt = 0b1
	tarNoReleaseVer   tarFileOpt = 0b0
	tarWithReleaseVer tarFileOpt = 0b1

	archARM64       = "arm64"
	archAMD64       = "amd64"
	lambdaArchAMD64 = "x86_64"
)

func tarFiles(pubPath, buildPath, appName, goos, goarch string, opt tarFileOpt) (string, string) {
	l.Debugf("tarFiles entry, pubPath = %s, buildPath = %s, appName = %s", pubPath, buildPath, appName)
	var gzFileName, gzFilePath string

	switch opt & tarReleaseVerMask {
	case tarWithReleaseVer:
		gzFileName = fmt.Sprintf("%s-%s-%s-%s.tar.gz",
			appName, goos, goarch, ReleaseVersion)
	case tarNoReleaseVer:
		gzFileName = fmt.Sprintf("%s-%s-%s.tar.gz",
			appName, goos, goarch)
	}

	gzFilePath = filepath.Join(pubPath, ReleaseType, gzFileName)

	args := []string{
		`czf`,
		gzFilePath,
		`-C`,
		// the whole basePath/appName-<goos>-<goarch> dir
		filepath.Join(buildPath, fmt.Sprintf("%s-%s-%s", appName, goos, goarch)), `.`,
	}

	cmd := exec.Command("tar", args...) //nolint:gosec

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	l.Debugf("tar %s...", gzFilePath)
	if err := cmd.Run(); err != nil {
		l.Fatal(err)
	}
	return gzFileName, gzFilePath
}

func addOSSFiles(ossPath string, files []ossFile) []ossFile {
	var res []ossFile
	for _, x := range files {
		res = append(res, ossFile{path.Join(ossPath, x.remote), x.local})
	}
	return res
}

type ossFile struct {
	remote, local string
}

func pubDCA() error {
	start := time.Now()
	basics := []ossFile{
		{"version", filepath.Join(DistDir, "version.dca")},
		{"dca.yaml", filepath.Join(DistDir, "dca.yaml")},
		{fmt.Sprintf("dca-%s.yaml", DCAVersion), filepath.Join(DistDir, "dca.yaml")},
	}

	if ossCli == nil {
		l.Warnf("ossCli not set")
		return nil
	}

	ossfiles := addOSSFiles(ossCli.WorkDir, basics)

	// test if all file ok before uploading
	for _, k := range ossfiles {
		if _, err := os.Stat(k.local); err != nil {
			return err
		}
	}

	for _, x := range ossfiles {
		fi, err := os.Stat(x.local)
		if err != nil {
			l.Errorf("os.Stat(%s): %s", x.local, err)
			return err
		}

		l.Debugf("%s => %s(%s)...", x.local, x.remote, humanize.Bytes(uint64(fi.Size())))

		if err := ossCli.Upload(x.local, x.remote); err != nil {
			return err
		}
	}

	l.Infof("Done!(elapsed: %v)", time.Since(start))
	return nil
}

//nolint:funlen,gocyclo
func PubDatakit() error {
	start := time.Now()

	// upload all build archs
	curArchs = ParseArchs(Archs)

	basics := []ossFile{
		// NOTE: these will overwrite online files, you should instead use xxx-<version>.
		{"datakit.yaml", filepath.Join(DistDir, "datakit.yaml")},
		{"datakit-elinker.yaml", filepath.Join(DistDir, "datakit-elinker.yaml")},
		{"install.sh", filepath.Join(DistDir, "install.sh")},
		{"install.ps1", filepath.Join(DistDir, "install.ps1")},
		{fmt.Sprintf("datakit-%s.yaml", ReleaseVersion), filepath.Join(DistDir, "datakit.yaml")},
		{fmt.Sprintf("datakit-elinker-%s.yaml", ReleaseVersion), filepath.Join(DistDir, "datakit-elinker.yaml")},
		{fmt.Sprintf("install-%s.sh", ReleaseVersion), filepath.Join(DistDir, "install.sh")},
		{fmt.Sprintf("install-%s.ps1", ReleaseVersion), filepath.Join(DistDir, "install.ps1")},

		// on Zh version for measurements meta and internla-pipelines
		{"measurements-meta.json", filepath.Join(DistDir, "datakit", inputs.I18nZh.String(), "measurements-meta.json")},
		{
			fmt.Sprintf("measurements-meta-%s.json", ReleaseVersion),
			filepath.Join(DistDir, "datakit", inputs.I18nZh.String(), "measurements-meta.json"),
		},
		{"internal-pipelines.json", filepath.Join(DistDir, "datakit", inputs.I18nZh.String(), "internal-pipelines.json")},

		// pipeline docs both export en & zh
		{"pipeline-docs.json", filepath.Join(DistDir, "datakit", inputs.I18nZh.String(), "pipeline-docs.json")},
		{"en/pipeline-docs.json", filepath.Join(DistDir, "datakit", inputs.I18nEn.String(), "pipeline-docs.json")},
	}

	// Darwin release not under CI, so disable upload `version' file under darwin,
	// only upload darwin related files.
	if Archs != datakit.OSArchDarwinAmd64 || runtime.GOOS != datakit.OSDarwin {
		basics = append(basics, ossFile{"version", path.Join(DistDir, ReleaseType, "version")})
	}

	// tar files and collect OSS upload/backup info
	for _, arch := range curArchs {
		parts := strings.Split(arch, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid arch: %s", arch)
		}

		goos, goarch := parts[0], parts[1]
		gzName, gzPath := tarFiles(DistDir, DistDir, AppName, parts[0], parts[1], tarWithReleaseVer)

		if isExtraLite() {
			gzName, gzPath := tarFiles(DistDir, DistDir, AppName+"_lite", parts[0], parts[1], tarWithReleaseVer)
			basics = append(basics, ossFile{gzName, gzPath})
		}

		if isExtraELinker() {
			gzName, gzPath := tarFiles(DistDir, DistDir, AppName+"_elinker", parts[0], parts[1], tarWithReleaseVer)
			basics = append(basics, ossFile{gzName, gzPath})
		}

		if isExtraAWSLambda() && (goarch == archAMD64 || goarch == archARM64) && goos == datakit.OSLinux {
			var (
				zipName       = fmt.Sprintf("%s-%s-%s-%s.zip", AppName+"_aws_extension", goos, goarch, ReleaseVersion)
				zipNameLatest = fmt.Sprintf("%s-%s-%s.zip", AppName+"_aws_extension", goos, goarch)
			)

			if of, err := uploadAWSLambdaZip(zipName, goos, goarch, true); err != nil {
				return err
			} else {
				basics = append(basics, *of)
			}

			if of, err := uploadAWSLambdaZip(zipNameLatest, goos, goarch, false); err != nil {
				return err
			} else {
				basics = append(basics, *of)
			}
		}

		// apm-auto-inject-launcher
		if goos == datakit.OSLinux && (goarch == archAMD64 || goarch == archARM64) && runtime.GOOS == datakit.OSLinux {
			gzName, gzPath := tarFiles(DistDir, DistDir, "datakit-apm-inject", goos, goarch, tarWithReleaseVer)
			basics = append(basics, ossFile{gzName, gzPath})
		}

		upgraderGZFile, upgraderGZPath := tarFiles(DistDir, DistDir, upgrader.BuildBinName, parts[0], parts[1], tarNoReleaseVer)
		// upload dk_upgrader-os-arch.tar.gz and dk_upgrader-os-arch-version.tar.gz
		basics = append(basics, ossFile{upgraderGZFile, upgraderGZPath})
		basics = append(basics, ossFile{fmt.Sprintf("dk_upgrader-%s-%s-%s.tar.gz", parts[0], parts[1], ReleaseVersion), upgraderGZPath})

		installerExe := fmt.Sprintf("installer-%s-%s", goos, goarch)
		installerExeWithVer := fmt.Sprintf("installer-%s-%s-%s", goos, goarch, ReleaseVersion)
		if parts[0] == datakit.OSWindows {
			installerExe = fmt.Sprintf("installer-%s-%s.exe", goos, goarch)
			installerExeWithVer = fmt.Sprintf("installer-%s-%s-%s.exe", goos, goarch, ReleaseVersion)
		}

		basics = append(basics, ossFile{gzName, gzPath})
		basics = append(basics, ossFile{installerExe, path.Join(DistDir, ReleaseType, installerExe)})
		basics = append(basics, ossFile{installerExeWithVer, path.Join(DistDir, ReleaseType, installerExe)})
	}

	ossfiles := addOSSFiles(ossCli.WorkDir, basics)

	// test if all file ok before uploading
	for _, k := range ossfiles {
		if _, err := os.Stat(k.local); err != nil {
			return err
		}
	}

	for _, x := range ossfiles {
		fi, err := os.Stat(x.local)
		if err != nil {
			l.Errorf("os.Stat(%s): %s", x.local, err)
			return err
		}

		l.Debugf("%s => %s(%s)...", x.local, x.remote, humanize.Bytes(uint64(fi.Size())))

		if err := ossRetryUpload(x.local, x.remote, 3); err != nil {
			return err
		}
	}

	if runtime.GOOS == datakit.OSLinux { // only publish helm under linux
		if err := pubDatakitHelm(); err != nil {
			return err
		}
	}

	l.Infof("Done!(elapsed: %v)", time.Since(start))
	return nil
}

func ossRetryUpload(local, remote string, retry int) error {
	var err error
	for i := 0; i < retry; i++ {
		if err = ossCli.Upload(local, remote); err == nil {
			return nil
		}
		time.Sleep(time.Second)
	}

	return err
}

func pubDatakitHelm() error {
	// run helm push command
	cmdArgs := []string{
		// TODO: we should switch to the new style: "helm push filepath.Join(DistDir, "datakit-"+ReleaseVersion+".tgz") HelmChartRepo "
		// old-style helm push
		"helm", "cm-push",
		filepath.Join(DistDir, "datakit-"+strings.Split(ReleaseVersion, "-")[0]+".tgz"),
		brand(Brand).chartRepoName(ReleaseType != ReleaseProduction),
	}

	msg, err := runEnv(cmdArgs, nil)
	if err != nil {
		return fmt.Errorf("failed to run %v: %w, msg: %s", cmdArgs, err, string(msg))
	}
	return nil
}

func uploadAWSLambdaZip(zipName string, goos string, goarch string, isUploadAWS bool) (*ossFile, error) {
	var (
		targetZipPath = filepath.Join(DistDir, ReleaseType, zipName)
		sourceZipPath = filepath.Join(DistDir,
			fmt.Sprintf("%s-%s-%s", AppName+"_aws_lambda", goos, goarch),
			AppName+"_aws_extension.zip")
	)

	cmd := exec.Command("cp", sourceZipPath, targetZipPath) //nolint:gosec
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to run: %w, msg: %s", err, string(output))
	}

	of := &ossFile{zipName, targetZipPath}

	// aws
	if isUploadAWS && EnableUploadAWS {
		rs := strings.Split(AWSRegions, ",")
		for _, region := range rs {
			err = os.Setenv("AWS_REGION", region)
			if err != nil {
				return nil, err
			}
			var arn string
			switch goarch {
			case archAMD64:
				arn, err = uploadAWSLayer(targetZipPath, AppName, lambdaArchAMD64)
			case archARM64:
				arn, err = uploadAWSLayer(targetZipPath, AppName, archARM64)
			default:
			}
			if err != nil {
				l.Warnf("failed to upload layer to aws %v: %v", region, err)
			}
			l.Infof("aws layer arn: %s", arn)
		}
	}

	return of, nil
}

// uploadAWSLayer load env:
// AWS_REGION
// AWS_ACCESS_KEY_ID
// AWS_SECRET_ACCESS_KEY
// AWS_SESSION_TOKEN.
func uploadAWSLayer(zipPath string, layerName string, arch string) (string, error) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Read zip file content
	zipBytes, err := os.ReadFile(zipPath) //nolint:gosec
	if err != nil {
		return "", fmt.Errorf("failed to read zip file: %w", err)
	}

	// Initialize Lambda service client
	svc := lambda.New(sess)

	// Upload layer
	resp, err := svc.PublishLayerVersion(&lambda.PublishLayerVersionInput{
		LayerName: aws.String(layerName + "-" + arch),
		Content: &lambda.LayerVersionContentInput{
			ZipFile: zipBytes,
		},
		CompatibleArchitectures: []*string{aws.String(arch)},
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload layer: %w", err)
	}

	return aws.StringValue(resp.LayerVersionArn), nil
}
