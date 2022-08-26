package refertable

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	ihttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/http"
)

var (
	_plReferTables = &PlReferTables{}
	_runner        = &Runner{
		initFinished: make(chan struct{}),
	}

	l = logger.DefaultSLogger("refer-table")
)

const (
	SchemeHTTP  = "http"
	SchemeHTTPS = "https"
)

func QueryReferTable(tableName string, colName []string, colValue []any,
	selected []string,
) (map[string]any, bool) {
	defer func() {
		if err := recover(); err != nil {
			l.Error(fmt.Errorf("run pl: %s", err))
		}
	}()

	if _plReferTables == nil {
		return nil, false
	}

	return _plReferTables.query(tableName, colName, colValue, selected)
}

func Stats() *ReferTableStats {
	return _plReferTables.stats()
}

func InitFinished(interval time.Duration) bool {
	return _runner.InitFinished(interval)
}

type InConfig struct {
	URL      string        `toml:"url"`
	Interval time.Duration `toml:"interval"`
}

type Runner struct {
	inConfig InConfig

	cli *retryablehttp.Client
	g   *goroutine.Group

	initFinished chan struct{}
}

func (r *Runner) InitFinished(interval time.Duration) bool {
	ticker := time.NewTicker(interval)

	if r.initFinished == nil {
		return false
	}

	select {
	case <-r.initFinished:
		return true
	case <-ticker.C:
		return false
	}
}

func InitLog() {
	l = logger.SLogger("refer-table")
}

func InitReferTableRunner(tableURL string, interval time.Duration) error {
	return initReferTableRunner(_runner, _plReferTables, tableURL, interval)
}

func initReferTableRunner(runner *Runner, plRefTables *PlReferTables, tableURL string, interval time.Duration) error {
	if tableURL == "" {
		return nil
	}
	if runner == nil {
		return fmt.Errorf("runner == nil")
	}

	if interval < time.Second*10 {
		interval = time.Second * 10
	}

	runner.inConfig.Interval = interval

	scheme, err := checkURL(tableURL)
	if err != nil {
		l.Error(err)
		return err
	}

	runner.inConfig.URL = tableURL
	runner.g = goroutine.NewGroup(goroutine.Option{Name: "refer-table"})

	switch scheme {
	case SchemeHTTP, SchemeHTTPS:
		opt := &ihttp.Options{
			DialTimeout:         time.Second * 30,
			MaxIdleConnsPerHost: 64,
		}
		runner.cli = newRetryCli(opt, time.Minute)
		runner.g.Go(func(ctx context.Context) error {
			return httpGetWkr(plRefTables, runner, datakit.Exit.Wait())
		})
	}

	return nil
}

func checkURL(tableURL string) (string, error) {
	u, err := url.Parse(tableURL)
	if err != nil {
		return "", fmt.Errorf("parse url: %s, error: %w",
			tableURL, err)
	}
	scheme := strings.ToLower(u.Scheme)
	switch scheme {
	case SchemeHTTP, SchemeHTTPS:
	default:
		return "", fmt.Errorf("url: %s, unsupported scheme %s",
			tableURL, scheme)
	}
	return scheme, nil
}

func httpGetWkr(plRefTable *PlReferTables, runner *Runner, ch <-chan any) error {
	ticker := time.NewTicker(runner.inConfig.Interval)
	for {
		getAndUpdate(plRefTable, runner)
		select {
		case <-ticker.C:
		case <-ch:
			return nil
		}
	}
}

func getAndUpdate(plRefTable *PlReferTables, runner *Runner) {
	if tables, err := httpGet(runner.cli, runner.inConfig.URL); err != nil {
		l.Error(err)
	} else {
		if err := plRefTable.updateAll(tables); err != nil {
			l.Error(err)
		}
	}

	select {
	case <-runner.initFinished:
	default:
		if runner.initFinished != nil {
			close(runner.initFinished)
		}
	}
}

func httpGet(cli *retryablehttp.Client, url string) ([]referTable, error) {
	resp, err := cli.Get(url)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("url: %s, status: %s", url, resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	tables, err := decodeJSONData(data)
	if err != nil {
		return nil, err
	}

	return tables, nil
}
