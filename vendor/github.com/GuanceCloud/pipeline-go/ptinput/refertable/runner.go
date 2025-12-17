// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package refertable

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
)

const (
	SchemeHTTP  = "http"
	SchemeHTTPS = "https"

	PullDuration = time.Second
)

var (
	l                = logger.DefaultSLogger("refer-table")
	defaultTransport = &http.Transport{
		DialContext: ((&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext),

		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
)

func InitLog() {
	l = logger.SLogger("refer-table")
}

type InConfig struct {
	URL      string        `toml:"url"`
	Interval time.Duration `toml:"interval"`
}

type RefTbCfg struct {
	// table data pull config
	URL      string
	Interval time.Duration

	// table store config
	UseSQLite     bool
	SQLiteMemMode bool
	DBPath        string
}

type ReferTable struct {
	inConfig     InConfig
	initFinished chan struct{}
	tables       PlReferTables
}

func NewReferTable(cfg RefTbCfg) (*ReferTable, error) {
	ref := &ReferTable{
		initFinished: make(chan struct{}),
	}

	if cfg.URL == "" {
		return nil, fmt.Errorf("empty url")
	}

	if cfg.UseSQLite {
		if cfg.SQLiteMemMode {
			l.Infof("using in-memory SQLite for refer-table")
			d, err := sql.Open("sqlite", ":memory:")
			if err != nil {
				return nil, fmt.Errorf("open in-memory SQLite failed: %w", err)
			}
			ref.tables = &PlReferTablesSqlite{db: d}
		} else {
			l.Infof("using on-disk SQLite for refer-table")
			d, err := sql.Open("sqlite", cfg.DBPath)
			if err != nil {
				return nil, fmt.Errorf("open SQLite at %s failed: %w", cfg.DBPath, err)
			}
			ref.tables = &PlReferTablesSqlite{db: d}
		}
	} else {
		l.Infof("using memory mode for refer-table")
		ref.tables = &PlReferTablesInMemory{}
	}

	if cfg.Interval < PullDuration {
		cfg.Interval = PullDuration
	}

	ref.inConfig.Interval = cfg.Interval
	ref.inConfig.URL = cfg.URL

	scheme, err := checkURL(ref.inConfig.URL)
	if err != nil {
		l.Error(err)
		return nil, err
	}

	switch scheme {
	case SchemeHTTP, SchemeHTTPS:
	default:
		return nil, fmt.Errorf("unsupported scheme: %s", scheme)
	}

	return ref, nil
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

// InitFinished used to check init status.
func (refT *ReferTable) InitFinished(waitTime time.Duration) bool {
	ticker := time.NewTicker(waitTime)

	if refT.initFinished == nil {
		return false
	}

	select {
	case <-refT.initFinished:
		return true
	case <-ticker.C:
		return false
	}
}

func (refT *ReferTable) Tables() PlReferTables {
	return refT.tables
}

func (refT *ReferTable) PullWorker(ctx context.Context) {
	ticker := time.NewTicker(refT.inConfig.Interval)
	for {
		if err := refT.getAndUpdate(); err != nil {
			l.Error(err)
		}
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}
	}
}

func (refT *ReferTable) getAndUpdate() error {
	if tables, err := httpGet(context.Background(), refT.inConfig.URL); err != nil {
		return fmt.Errorf("get table data from URL: %w", err)
	} else {
		if refT.tables == nil {
			return nil
		}
		if err := refT.tables.updateAll(tables); err != nil {
			l.Errorf("failed to update tables: %w", err)
		}
	}

	select {
	case <-refT.initFinished:
	default:
		if refT.initFinished != nil {
			close(refT.initFinished)
		}
	}
	return nil
}

func httpGet(ctx context.Context, url string) ([]referTable, error) {
	resp, err := DoRequestWithRetry(ctx,
		"GET", url, nil, nil,
		&DefaultRetryConfig, defaultTransport)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("url: %s, status: %s", url, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	tables, err := decodeJSONData(data)
	if err != nil {
		return nil, err
	}

	return tables, nil
}
