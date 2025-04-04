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

	"github.com/GuanceCloud/cliutils"
	humanize "github.com/dustin/go-humanize"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/cmd/upgrader/upgrader"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export"
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
	TarRlsVerMask tarFileOpt = 0b1
	TarNoRlsVer   tarFileOpt = 0b0
	TarWithRlsVer tarFileOpt = 0b1
)

const (
	ARM64 = "arm64"
	AMD64 = "amd64"
)

const Linux = "linux"

func tarFiles(pubPath, buildPath, appName, goos, goarch string, opt tarFileOpt) (string, string) {
	l.Debugf("tarFiles entry, pubPath = %s, buildPath = %s, appName = %s", pubPath, buildPath, appName)
	var gzFileName, gzFilePath string

	switch opt & TarRlsVerMask {
	case TarWithRlsVer:
		gzFileName = fmt.Sprintf("%s-%s-%s-%s.tar.gz",
			appName, goos, goarch, ReleaseVersion)
	case TarNoRlsVer:
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

func addOSSFiles(ossPath string, files map[string]string) map[string]string {
	res := map[string]string{}
	for k, v := range files {
		res[path.Join(ossPath, k)] = v
	}
	return res
}

//nolint:funlen,gocyclo
func PubDatakit() error {
	start := time.Now()
	var ak, sk, bucket, ossHost string

	// 在你本地设置好这些 oss-key 环境变量
	switch ReleaseType {
	case ReleaseTesting, ReleaseProduction, ReleaseLocal:
		tag := strings.ToUpper(ReleaseType)
		ak = os.Getenv(tag + "_OSS_ACCESS_KEY")
		sk = os.Getenv(tag + "_OSS_SECRET_KEY")
		bucket = os.Getenv(tag + "_OSS_BUCKET")
		ossHost = os.Getenv(tag + "_OSS_HOST")
	default:
		return fmt.Errorf("unknown release type: %s", ReleaseType)
	}

	if ak == "" || sk == "" {
		return fmt.Errorf("OSS %s/%s not set",
			strings.ToUpper(ReleaseType)+"_OSS_ACCESS_KEY",
			strings.ToUpper(ReleaseType)+"_OSS_SECRET_KEY")
	}

	ossSlice := strings.SplitN(UploadAddr, "/", 2) // at least 2 parts
	if len(ossSlice) != 2 {
		return fmt.Errorf("invalid download addr: %s", UploadAddr)
	}
	OSSPath = ossSlice[1]

	oc := &cliutils.OssCli{
		Host:       ossHost,
		PartSize:   512 * 1024 * 1024,
		AccessKey:  ak,
		SecretKey:  sk,
		BucketName: bucket,
		WorkDir:    OSSPath,
	}

	if err := oc.Init(); err != nil {
		return err
	}

	// upload all build archs
	curArchs = ParseArchs(Archs)

	if err := generateInstallScript(); err != nil {
		return err
	}

	exporter := export.NewIntegration(export.WithTopDir(PubDir))
	if err := exporter.Export(); err != nil {
		return err
	}

	basics := map[string]string{
		"version": path.Join(PubDir, ReleaseType, "version"),

		// NOTE: these will overwrite online files, you should instead use xxx-<version>.
		"datakit.yaml":         "datakit.yaml",
		"datakit-elinker.yaml": "datakit-elinker.yaml",
		"install.sh":           "install.sh",
		"install.ps1":          "install.ps1",
		fmt.Sprintf("datakit-%s.yaml", ReleaseVersion):         "datakit.yaml",
		fmt.Sprintf("datakit-elinker-%s.yaml", ReleaseVersion): "datakit-elinker.yaml",
		fmt.Sprintf("install-%s.sh", ReleaseVersion):           "install.sh",
		fmt.Sprintf("install-%s.ps1", ReleaseVersion):          "install.ps1",

		"measurements-meta.json": filepath.Join(PubDir,
			"datakit",
			inputs.I18nZh.String(), // on Zh version
			"measurements-meta.json"),

		"pipeline-docs.json": filepath.Join(PubDir,
			"datakit",
			inputs.I18nZh.String(),
			"pipeline-docs.json"),

		"en/pipeline-docs.json": filepath.Join(PubDir,
			"datakit",
			inputs.I18nEn.String(),
			"pipeline-docs.json"),

		// only Zh version
		"internal-pipelines.json": filepath.Join(PubDir,
			"datakit",
			inputs.I18nZh.String(),
			"internal-pipelines.json"),
	}

	// tar files and collect OSS upload/backup info
	for _, arch := range curArchs {
		parts := strings.Split(arch, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid arch: %s", arch)
		}
		goos, goarch := parts[0], parts[1]
		gzName, gzPath := tarFiles(PubDir, BuildDir, AppName, parts[0], parts[1], TarWithRlsVer)
		// gzName := fmt.Sprintf("%s-%s-%s.tar.gz", AppName, goos+"-"+goarch, ReleaseVersion)

		if isExtraLite() {
			gzName, gzPath := tarFiles(PubDir, BuildDir, AppName+"_lite", parts[0], parts[1], TarWithRlsVer)
			basics[gzName] = gzPath
		}
		if isExtraELinker() {
			gzName, gzPath := tarFiles(PubDir, BuildDir, AppName+"_elinker", parts[0], parts[1], TarWithRlsVer)
			basics[gzName] = gzPath
		}
		if isExtraAWSLambda() && (goarch == AMD64 || goarch == ARM64) && goos == Linux {
			zipName := fmt.Sprintf("%s-%s-%s-%s.zip", AppName+"_aws_extension", goos, goarch, ReleaseVersion)
			err := uploadAWSLambdaZip(zipName, goos, goarch, basics, true)
			if err != nil {
				l.Fatal(err)
			}
			zipNameLatest := fmt.Sprintf("%s-%s-%s.zip", AppName+"_aws_extension", goos, goarch)
			err = uploadAWSLambdaZip(zipNameLatest, goos, goarch, basics, false)
			if err != nil {
				l.Fatal(err)
			}
		}

		// apm-auto-inject-launcher
		if goos == Linux && (goarch == AMD64 || goarch == ARM64) && runtime.GOOS == "linux" {
			gzName, gzPath := tarFiles(
				PubDir, BuildDir, "datakit-apm-inject", goos, goarch, TarWithRlsVer)
			basics[gzName] = gzPath
		}

		upgraderGZFile, upgraderGZPath := tarFiles(PubDir, BuildDir, upgrader.BuildBinName, parts[0], parts[1], TarNoRlsVer)

		installerExe := fmt.Sprintf("installer-%s-%s", goos, goarch)
		installerExeWithVer := fmt.Sprintf("installer-%s-%s-%s", goos, goarch, ReleaseVersion)
		if parts[0] == datakit.OSWindows {
			installerExe = fmt.Sprintf("installer-%s-%s.exe", goos, goarch)
			installerExeWithVer = fmt.Sprintf("installer-%s-%s-%s.exe", goos, goarch, ReleaseVersion)
		}

		basics[gzName] = gzPath
		basics[upgraderGZFile] = upgraderGZPath
		basics[installerExe] = path.Join(PubDir, ReleaseType, installerExe)
		basics[installerExeWithVer] = path.Join(PubDir, ReleaseType, installerExe)
	}

	// Darwin release not under CI, so disable upload `version' file under darwin,
	// only upload darwin related files.
	if Archs == datakit.OSArchDarwinAmd64 && runtime.GOOS == datakit.OSDarwin {
		delete(basics, "version")
	}

	ossfiles := addOSSFiles(OSSPath, basics)

	// test if all file ok before uploading
	for _, k := range ossfiles {
		if _, err := os.Stat(k); err != nil {
			return err
		}
	}

	l.Infof("upload to %q...", UploadAddr)
	for k, v := range ossfiles {
		fi, err := os.Stat(v)
		if err != nil {
			l.Errorf("os.Stat(%s): %s", v, err)
			return err
		}

		l.Debugf("%s => %s(%s)...", v, k, humanize.Bytes(uint64(fi.Size())))

		if err := oc.Upload(v, k); err != nil {
			return err
		}
	}

	l.Infof("Done!(elapsed: %v)", time.Since(start))
	return nil
}

func uploadAWSLambdaZip(zipName string, goos string, goarch string, basics map[string]string, isUploadAWS bool) error {
	// oss
	targetZipPath := filepath.Join(PubDir, ReleaseType, zipName)
	sourceZipPath := filepath.Join(BuildDir, fmt.Sprintf("%s-%s-%s", AppName+"_aws_lambda", goos, goarch), AppName+"_aws_extension.zip")
	cmd := exec.Command("cp", sourceZipPath, targetZipPath) //nolint:gosec
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run: %w, msg: %s", err, string(output))
	}
	basics[zipName] = targetZipPath
	// aws
	if isUploadAWS && EnableUploadAWS {
		rs := strings.Split(AWSRegions, ",")
		for _, region := range rs {
			err = os.Setenv("AWS_REGION", region)
			if err != nil {
				return err
			}
			var arn string
			switch goarch {
			case AMD64:
				arn, err = uploadAWSLayer(targetZipPath, AppName, "x86_64")
			case ARM64:
				arn, err = uploadAWSLayer(targetZipPath, AppName, ARM64)
			default:
			}
			if err != nil {
				l.Warnf("failed to upload layer to aws %v: %v", region, err)
			}
			l.Infof("aws layer arn: %s", arn)
		}
	}
	return nil
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
