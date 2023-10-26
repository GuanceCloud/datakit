// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"encoding/json"
	"os"
	"time"

	"github.com/influxdata/influxdb1-client/models"
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
)

func parseLineProto() error {
	x, err := os.ReadFile(*flagToolParseLineProtocol)
	if err != nil {
		cp.Errorf("os.ReadFile: %s\n", err.Error())
		return err
	}

	pts, err := models.ParsePointsWithPrecision(x, time.Now(), "n")
	if err != nil {
		cp.Errorf("ParsePoints: %s\n", err.Error())
		return err
	}

	type parseMeasurement struct {
		Points     int `json:"points"`
		TimeSeries int `json:"time_series"`
	}

	measurements := map[string]*parseMeasurement{}
	hashids := map[uint64]bool{}

	for _, pt := range pts {
		id := pt.HashID()

		name := string(pt.Name())

		// get each measurement's time-series count
		if m, ok := measurements[name]; ok {
			m.Points++
		} else {
			measurements[name] = &parseMeasurement{
				Points: 1,
			}
		}

		if _, ok := hashids[id]; !ok {
			hashids[id] = true
			measurements[name].TimeSeries++
		}
	}

	if *flagToolJSON {
		j, err := json.MarshalIndent(map[string]any{
			"point":        len(pts),
			"time_serial":  len(hashids),
			"measurements": measurements,
		}, "", "  ")
		if err != nil {
			cp.Errorf("Marshal: %s\n", err.Error())
			return err
		}

		cp.Output("%s\n", string(j))
	} else {
		cp.Infof("Parse %d points OK, with %d measurements and %d time series.\n",
			len(pts), len(measurements), len(hashids))
	}

	return nil
}
