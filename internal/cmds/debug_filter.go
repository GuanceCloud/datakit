// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	fp "github.com/GuanceCloud/cliutils/filter"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/influxdata/influxdb1-client/models"
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/filter"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
)

// https://stackoverflow.com/questions/7427262/how-to-read-a-file-into-a-variable-in-shell
// https://stackoverflow.com/a/7977295/342348

func readFileOrData(s []byte) []byte {
	if fi, err := os.Stat(string(s)); err != nil {
		return s
	} else {
		if fi.IsDir() {
			return s
		}

		if data, err := ioutil.ReadFile(string(s)); err != nil {
			return s
		} else {
			return data
		}
	}
}

func debugFilter(filterConf, data []byte) error {
	fp.Init()

	if len(data) == 0 {
		return fmt.Errorf("debug data(line protocol not set)")
	}

	var f filter.Filters

	if err := json.Unmarshal(readFileOrData(filterConf), &f); err != nil {
		return fmt.Errorf("invalid filter rule(json required): %w", err)
	}

	pts, err := models.ParsePointsWithPrecision(readFileOrData(data), time.Now(), "n")
	if err != nil {
		cp.Errorf("ParsePoints: %s\n", err.Error())
		return err
	}

	start := time.Now()
	for k, v := range f.Filters {
		conds, err := filter.GetConds(v)
		if err != nil {
			return fmt.Errorf("invalid filter rule: %w", err)
		}

		var tfdatas []*filter.TFData
		for _, pt := range pts {
			tfdata := filter.NewTFData(
				point.CatString(k),
				&dkpt.Point{
					Point: point.FromModelsLP(pt).LPPoint(),
				})
			if err != nil {
				return err
			}

			tfdatas = append(tfdatas, tfdata)
		}

		for i, tfdata := range tfdatas {
			if j := conds.Eval(tfdata); j >= 0 {
				cp.Infof("Dropped\n\n")
				cp.Output("\t%s\n\n", pts[i].String())
				cp.Infof("By %dth rule(cost %s) from category %q:\n\n", j+1, time.Since(start), point.CatString(k))
				cp.Output("\t%+s\n", v[j])
			}
		}
	}

	return nil
}
