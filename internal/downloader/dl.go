package downloader

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/dustin/go-humanize"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

var (
	CurDownloading string
	l              = logger.DefaultSLogger("downloader")
)

type writeCounter struct {
	total   uint64
	current uint64
	last    float64
}

func (wc *writeCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.current += uint64(n)
	wc.last += float64(n)
	wc.PrintProgress()
	return n, nil
}

func (wc *writeCounter) PrintProgress() {
	if wc.last > float64(wc.total)*0.01 || wc.current == wc.total { // update progress-bar each 1%
		fmt.Printf("\r%s", strings.Repeat(" ", 36)) //nolint:gomnd
		fmt.Printf("\rDownloading(% 7s)... %s/%s", CurDownloading, humanize.Bytes(wc.current), humanize.Bytes(wc.total))
		wc.last = 0.0
	}
}

func doExtract(r io.Reader, to string) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		l.Error(err)
		return err
	}

	defer gzr.Close()
	tr := tar.NewReader(gzr)
	for {
		hdr, err := tr.Next()
		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			l.Error(err)
			return err
		case hdr == nil:
			continue
		}

		target := filepath.Join(to, hdr.Name) //nolint:gosec

		switch hdr.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, os.ModePerm); err != nil {
					l.Error(err)
					return err
				}
			}

		case tar.TypeReg:

			if err := os.MkdirAll(filepath.Dir(target), os.ModePerm); err != nil {
				l.Error(err)
				return err
			}

			// TODO: lock file before extracting, to avoid `text file busy` error
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(hdr.Mode))
			if err != nil {
				l.Error(err)
				return err
			}

			if _, err := io.Copy(f, tr); err != nil { //nolint:gosec
				l.Error(err)
				return err
			}

			if err := f.Close(); err != nil {
				l.Warnf("f.Close(): %v, ignored", err)
			}

		default:
			l.Warnf("unexpected file %s", target)
		}
	}
}

func Download(cli *http.Client, from, to string, progress, downloadOnly bool) error {
	req, err := http.NewRequest("GET", from, nil)
	if err != nil {
		l.Error(err)
		return err
	}

	req.Header.Add("Accept-Encoding", "gzip")

	resp, err := cli.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	progbar := &writeCounter{
		total: uint64(resp.ContentLength),
	}

	if downloadOnly {
		return doDownload(io.TeeReader(resp.Body, progbar), to)
	}

	if !progress {
		return doExtract(resp.Body, to)
	}

	return doExtract(io.TeeReader(resp.Body, progbar), to)
}

func doDownload(r io.Reader, to string) error {
	f, err := os.OpenFile(to, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}

	if _, err := io.Copy(f, r); err != nil { //nolint:gosec
		return err
	}

	return f.Close()
}
