// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package postgresql

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
)

const (
	DefaultDiscoveryInterval = time.Minute * 10
	MaxDiscoveryDatabases    = 100
	sqlGetDatabaseList       = `
select datname from pg_catalog.pg_database where datistemplate = false;
`
)

type DatabaseDiscovery struct {
	sync.RWMutex
	Enabled      bool             `toml:"enabled"`
	MaxDatabases uint32           `toml:"max_databases"`
	Include      []string         `toml:"include"`
	Exclude      []string         `toml:"exclude"`
	Interval     datakit.Duration `toml:"interval"`

	ipt    *Input
	filter *Filter
	dbs    []string
}

func (dd *DatabaseDiscovery) init(ipt *Input) error {
	dd.ipt = ipt
	if dd.Interval.Duration <= 0 {
		dd.Interval = datakit.Duration{Duration: DefaultDiscoveryInterval}
	}

	if dd.MaxDatabases == 0 {
		dd.MaxDatabases = MaxDiscoveryDatabases
	}

	filter, err := NewFilter(FilterConfig{
		Include: dd.Include,
		Exclude: dd.Exclude,
	})
	if err != nil {
		return fmt.Errorf("new filter failed: %w", err)
	}
	dd.filter = filter

	return nil
}

func (dd *DatabaseDiscovery) Start() {
	g := goroutine.NewGroup(goroutine.Option{Name: "postgres_database_discovery"})
	g.Go(func(ctx context.Context) error {
		l.Infof("start to discovery databases, every %s", dd.Interval.Duration)
		ticker := time.NewTicker(dd.Interval.Duration)
		defer ticker.Stop()

		for {
			dd.discoverDatabases() // Initial discovery

			select {
			case <-datakit.Exit.Wait():
				return nil
			case <-dd.ipt.semStop.Wait():
				return nil
			case <-ticker.C:
			}
		}
	})
}

func (dd *DatabaseDiscovery) discoverDatabases() {
	ipt := dd.ipt
	dbs := []string{}
	rows, err := ipt.service.QueryByDatabase(sqlGetDatabaseList, "")
	if err != nil {
		l.Errorf("query failed: %s", err.Error())
		return
	}

	defer rows.Close() //nolint:errcheck

	for rows.Next() {
		var db string
		if err := rows.Scan(&db); err != nil {
			l.Errorf("scan failed: %s", err.Error())
			return
		} else if dd.filter.Allow(db) && uint32(len(dbs)) < dd.MaxDatabases {
			dbs = append(dbs, db)
		}
	}

	dd.SetDatabases(dbs)
}

func (dd *DatabaseDiscovery) SetDatabases(dbs []string) {
	dd.Lock()
	defer dd.Unlock()
	dd.dbs = dbs
}

func (dd *DatabaseDiscovery) GetDatabases() []string {
	dd.RLock()
	defer dd.RUnlock()
	dbs := make([]string, len(dd.dbs))
	copy(dbs, dd.dbs)
	return dbs
}
