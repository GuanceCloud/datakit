// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package downloader wrap HTTP download function
package downloader

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	humanize "github.com/dustin/go-humanize"

	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
)

var CurDownloading string

type WriteCounter struct {
	Total   uint64
	current uint64
	last    float64
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.current += uint64(n)
	wc.last += float64(n)
	wc.PrintProgress()
	if n > 0 && wc.current >= wc.Total {
		cp.Println()
	}
	return n, nil
}

func (wc *WriteCounter) PrintProgress() {
	if wc.last > float64(wc.Total)*0.01 || wc.current == wc.Total { // update progress-bar each 1%
		cp.Printf("\r%s", strings.Repeat(" ", 100)) //nolint:gomnd
		cp.Printf("\rDownloading(% 7s)... %s/%s", CurDownloading, humanize.Bytes(wc.current), humanize.Bytes(wc.Total))
		wc.last = 0.0
	}
}

// Extract unzip files from @r to directory @to.
//
//nolint:cyclop
func Extract(r io.Reader, to string) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("gzip.NewReader: %w", err)
	}

	defer gzr.Close() //nolint:errcheck

	tr := tar.NewReader(gzr)
	for {
		hdr, err := tr.Next()

		if errors.Is(err, io.EOF) {
			return nil
		}

		if err != nil {
			return fmt.Errorf("tr.Next(): %w", err)
		}

		if hdr == nil {
			continue
		}

		target := filepath.Join(to, hdr.Name) //nolint:gosec

		switch hdr.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, os.ModePerm); err != nil {
					return fmt.Errorf("MkdirAll: %w", err)
				}
			}

		case tar.TypeReg:
			if err := extractRegularFile(tr, target, os.FileMode(hdr.Mode)); err != nil {
				return err
			}

		default:
			return fmt.Errorf("unexpected file %s", target)
		}
	}
}

func extractRegularFile(r io.Reader, target string, mode os.FileMode) error {
	target = filepath.Clean(target)

	if err := os.MkdirAll(filepath.Dir(target), os.ModePerm); err != nil {
		return fmt.Errorf("MkdirAll: %w", err)
	}

	if runtime.GOOS == "windows" {
		_ = os.Remove(target)
		f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, mode) //nolint:gosec
		if err != nil {
			return fmt.Errorf("OpenFile: %w", err)
		}

		if _, err := io.Copy(f, r); err != nil { //nolint:gosec
			_ = f.Close()
			return fmt.Errorf("io.Copy: %w", err)
		}

		if err := f.Close(); err != nil {
			return fmt.Errorf("on Close(): %w", err)
		}
		return nil
	}

	f, err := os.CreateTemp(filepath.Dir(target), ".dl.tmp.")
	if err != nil {
		return fmt.Errorf("CreateTemp: %w", err)
	}
	tmpName := f.Name()
	defer func() {
		_ = os.Remove(tmpName)
	}()

	if _, err := io.Copy(f, r); err != nil { //nolint:gosec
		_ = f.Close()
		return fmt.Errorf("io.Copy: %w", err)
	}

	if err := f.Chmod(mode); err != nil {
		_ = f.Close()
		return fmt.Errorf("chmod: %w", err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("on Close(): %w", err)
	}

	if err := os.Rename(tmpName, target); err != nil {
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}

func Download(cli *http.Client, from, to string, progress, downloadOnly bool) error {
	req, err := http.NewRequest("GET", from, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Accept-Encoding", "gzip")

	resp, err := cli.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close() //nolint:errcheck
	progbar := &WriteCounter{
		Total: uint64(resp.ContentLength),
	}

	if downloadOnly {
		if to == "" {
			to = filepath.Base(from)
		}
		return doDownload(io.TeeReader(resp.Body, progbar), to)
	}

	if !progress {
		return Extract(resp.Body, to)
	}

	return Extract(io.TeeReader(resp.Body, progbar), to)
}

func doDownload(r io.Reader, to string) error {
	to = filepath.Clean(to)

	if _, err := os.Stat(filepath.Dir(to)); err != nil {
		if err := os.MkdirAll(filepath.Dir(to), os.ModePerm); err != nil {
			return fmt.Errorf("MkdirAll: %w", err)
		}
	}

	_ = os.Remove(filepath.Clean(to))
	f, err := os.OpenFile(filepath.Clean(to),
		os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}

	if _, err := io.Copy(f, r); err != nil { //nolint:gosec
		return err
	}

	return f.Close()
}
