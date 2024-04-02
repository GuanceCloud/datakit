// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package plval store pipeline private values.
package plval

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"runtime"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/pipeline"
	plmanager "github.com/GuanceCloud/cliutils/pipeline/manager"
	"github.com/GuanceCloud/cliutils/pipeline/ptinput/ipdb"
	"github.com/GuanceCloud/cliutils/pipeline/ptinput/plmap"
	"github.com/GuanceCloud/cliutils/pipeline/ptinput/refertable"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/offload"
)

var (
	l = logger.DefaultSLogger("pl-val")

	g = goroutine.NewGroup(goroutine.Option{
		Name: "pipeline",
	})
)

var (
	// pipeline mamager.
	_managerIns *plmanager.Manager

	// refer table.
	_referTable *refertable.ReferTable

	// offload.
	_offloadWkr *offload.OffloadWorker

	// ipdb.
	_ipdb ipdb.IPdb
)

func SetManager(m *plmanager.Manager) {
	_managerIns = m
}

func GetManager() (*plmanager.Manager, bool) {
	if _managerIns == nil {
		return nil, false
	}
	return _managerIns, true
}

func SetIPDB(db ipdb.IPdb) {
	_ipdb = db
}

func SetRefTb(tb *refertable.ReferTable) {
	_referTable = tb
}

func GetRefTb() (*refertable.ReferTable, bool) {
	if _referTable == nil {
		return nil, false
	}
	return _referTable, true
}

func SetOffload(offl *offload.OffloadWorker) {
	_offloadWkr = offl
}

func GetOffload() (*offload.OffloadWorker, bool) {
	if _offloadWkr == nil {
		return nil, false
	}
	return _offloadWkr, true
}

func GetIPDB() (ipdb.IPdb, bool) {
	if _ipdb == nil {
		return nil, false
	}
	return _ipdb, true
}

func SearchISP(ip string) string {
	if _ipdb != nil {
		return _ipdb.SearchIsp(ip)
	}
	return "unknown"
}

func Geo(ip string) (*ipdb.IPdbRecord, error) {
	if _ipdb != nil {
		return _ipdb.Geo(ip)
	}
	return nil, fmt.Errorf("ipdb not ready")
}

const maxCustomer = 16

func InitPlVal(cfg *plmanager.PipelineCfg, upFn plmap.UploadFunc, gTags map[string]string,
	installDir string,
) error {
	l = logger.SLogger("pipeline")

	offload.InitLog()
	pipeline.InitLog()

	// load grok pattern
	if err := plmanager.LoadPatterns(datakit.PipelinePatternDir); err != nil {
		l.Warnf("load pattern from directory failed: %w", err)
	}

	var gTagsLi [][2]string

	for k, v := range gTags {
		gTagsLi = append(gTagsLi, [2]string{k, v})
	}

	// init script manager
	managerIns := plmanager.NewManager(plmanager.NewManagerCfg(upFn, gTagsLi))
	plmanager.InitStore(managerIns, installDir)
	SetManager(managerIns)

	// init ipdb
	if ipdb, err := plmanager.InitIPdb(datakit.DataDir, cfg); err != nil {
		l.Warnf("init ipdb error: %s", err.Error())
	} else {
		SetIPDB(ipdb)
	}

	// init refer-table
	if cfg != nil && cfg.ReferTableURL != "" {
		dur, err := time.ParseDuration(cfg.ReferTablePullInterval)
		if err != nil {
			l.Warnf("refer table pull interval %s, err: %v", dur, err)
			dur = time.Minute * 5
		}

		if referTable, err := refertable.NewReferTable(refertable.RefTbCfg{
			URL:      cfg.ReferTableURL,
			Interval: dur,

			UseSQLite:     cfg.UseSQLite,
			SQLiteMemMode: cfg.SQLiteMemMode,
			DBPath:        filepath.Join(datakit.DataDir, "reftable_sqlite"),
		}); err != nil {
			l.Error("init refer table, error: %v", err)
		} else {
			SetRefTb(referTable)
			g.Go(func(ctx context.Context) error {
				if v, ok := GetRefTb(); ok && v != nil {
					v.PullWorker(ctx)
					return nil
				} else {
					return fmt.Errorf("pipeline refertable not ready")
				}
			})
		}
	}

	// init offload
	if cfg != nil && cfg.Offload != nil && cfg.Offload.Receiver != "" &&
		len(cfg.Offload.Addresses) != 0 {
		offloadCfg := &offload.OffloadConfig{
			Receiver:  cfg.Offload.Receiver,
			Addresses: cfg.Offload.Addresses,
		}

		if wkr, err := offload.NewOffloader(offloadCfg); err != nil {
			l.Errorf("init offload worker, error: %v", err)
		} else {
			SetOffload(wkr)

			for i := 0; i < int(math.Ceil(float64(runtime.NumCPU())*1.5)); i++ {
				if i >= maxCustomer {
					break
				}

				// logging only
				g.Go(func(ctx context.Context) error {
					if v, ok := GetOffload(); ok && v != nil {
						return v.Customer(ctx, point.Logging)
					}
					return fmt.Errorf("pipeline offloader not ready")
				})
			}
		}
	}

	return nil
}
