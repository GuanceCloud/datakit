package cmds

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/dustin/go-humanize"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/geo"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ip2isp"
)

var (
	suggestions = []prompt.Suggest{
		{Text: "exit", Description: "exit cmd"},
		{Text: "Q", Description: "exit cmd"},
		{Text: "flushall", Description: "k8s interactive command to generate deploy file"},
	}

	l = logger.DefaultSLogger("cmds")

	curDownloading string
)

type completer struct{}

func newCompleter() (*completer, error) {
	return &completer{}, nil
}

func (c *completer) Complete(d prompt.Document) []prompt.Suggest {
	w := d.GetWordBeforeCursor()
	switch w {
	case "":
		return []prompt.Suggest{}
	default:
		return prompt.FilterFuzzy(suggestions, w, true)
	}
}

func ipInfo(ip string) (map[string]string, error) {

	datadir := datakit.DataDir

	if err := geo.LoadIPLib(filepath.Join(datadir, "iploc.bin")); err != nil {
		return nil, err
	}

	if err := ip2isp.Init(filepath.Join(datadir, "ip2isp.txt")); err != nil {
		return nil, err
	}

	x, err := geo.Geo(ip)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"city":     x.City,
		"province": x.Region,
		"country":  x.Country_short,
		"isp":      ip2isp.SearchIsp(ip),
		"ip":       ip,
	}, nil
}

func setCmdRootLog(rl string) {

	if err := logger.InitRoot(&logger.Option{Path: rl, Flags: logger.OPT_DEFAULT, Level: logger.DEBUG}); err != nil {
		l.Error(err)
		return
	}

	// setup config module logger, redirect to @rl
	config.SetLog()

	l = logger.SLogger("cmds")
	l.Infof("root log path set to %s", rl)
}

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
		fmt.Printf("\r%s", strings.Repeat(" ", 36))
		fmt.Printf("\rDownloading(% 7s)... %s/%s", curDownloading, humanize.Bytes(wc.current), humanize.Bytes(wc.total))
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

		target := filepath.Join(to, hdr.Name)

		switch hdr.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					l.Error(err)
					return err
				}
			}

		case tar.TypeReg:

			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
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
