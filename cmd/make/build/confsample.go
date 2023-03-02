// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package build

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func DumpSamples() {
	tarPath := "datakit-conf-samples.tar.gz"
	ossPath := "datakit/datakit-conf-samples.tar.gz"
	if err := doDownloadSamples(ossPath, tarPath); err != nil {
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
}

func DownloadSamples() {
	tarPath := "datakit-conf-samples.tar.gz"
	ossPath := "datakit/datakit-conf-samples.tar.gz"
	if err := doDownloadSamples(ossPath, tarPath); err != nil {
		l.Fatalf("fail to download samples: %v", err)
	}
	if err := extractSamples(tarPath); err != nil {
		l.Fatalf("fail to extract samples: %v", err)
	}
}

// uploadSamples uploads given conf.tar.gz to oss.
func uploadSamples(from, to string) error {
	oc, err := GetOSSClient()
	if err != nil {
		return err
	}
	return oc.Upload(from, to)
}

// compressSamples compresses given samples directory.
func compressSamples(from, to string) error {
	fw, err := os.Create(filepath.Clean(to))
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
		f, err := os.Create(filepath.Clean(filepath.Join(catalogPath, name+".conf")))
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
		dest, err := os.Create(filepath.Clean(path))
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

func doDownloadSamples(from, to string) error {
	oc, err := GetOSSClient()
	if err != nil {
		return err
	}
	if err := oc.Download(from, to); err != nil {
		return fmt.Errorf("fail to download from oss, bucket: %s: %w", oc.BucketName, err)
	}
	return nil
}

func getDirName() string {
	var ret string
	idx := strings.Index(ReleaseVersion, "-")
	if idx != -1 {
		ret = ReleaseVersion[:idx]
	} else {
		ret = ReleaseVersion
	}
	return ret
}
