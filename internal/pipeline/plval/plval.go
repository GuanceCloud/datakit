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

	_enableAppendRunInfo bool = false
)

func EnableAppendRunInfo() bool {
	return _enableAppendRunInfo
}

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

const maxCustomer = 16

var localDefaultPipeline map[point.Category]string

func GetLocalDefaultPipeline() map[point.Category]string {
	return localDefaultPipeline
}

func PreferLocalDefaultPipeline(m map[point.Category]string) map[point.Category]string {
	result := map[point.Category]string{}
	for k, v := range m {
		result[k] = v
	}
	for k, v := range GetLocalDefaultPipeline() {
		result[k] = v
	}

	return result
}

func InitPlVal(cfg *PipelineCfg, upFn plmap.UploadFunc, gTags map[string]string,
	installDir string,
) error {
	l = logger.SLogger("plval")

	offload.InitOffload()
	pipeline.InitLog()

	// load grok pattern
	if err := LoadPatterns(datakit.PipelinePatternDir); err != nil {
		l.Warnf("load pattern from directory failed: %w", err)
	}

	var gTagsLi [][2]string

	for k, v := range gTags {
		gTagsLi = append(gTagsLi, [2]string{k, v})
	}

	// init script manager
	managerIns := plmanager.NewManager(plmanager.NewManagerCfg(upFn, gTagsLi))
	plmanager.InitStore(managerIns, installDir, nil)
	SetManager(managerIns)

	// init ipdb
	if ipdb, err := InitIPdb(datakit.DataDir, cfg); err != nil {
		l.Warnf("init ipdb error: %s", err.Error())
	} else {
		SetIPDB(ipdb)
	}

	if cfg != nil && cfg.EnableDebugFields {
		_enableAppendRunInfo = true
	}

	if cfg != nil && len(cfg.DefaultPipeline) > 0 {
		mp := map[point.Category]string{}
		for k, v := range cfg.DefaultPipeline {
			mp[point.CatString(k)] = v
		}
		localDefaultPipeline = mp
		managerIns.UpdateDefaultScript(mp)
		l.Infof("set default pipeline: %v", mp)
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

			nworkers := int(math.Ceil(float64(datakit.AvailableCPUs) * 1.5))
			l.Infof("start %d offload workers...", nworkers)
			for i := 0; i < nworkers; i++ {
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
