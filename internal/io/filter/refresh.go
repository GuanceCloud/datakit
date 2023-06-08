// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package filter

import (
	"crypto/md5" //nolint:gosec
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/parser"
)

func (f *filter) refresh(body []byte) error {
	bodymd5 := fmt.Sprintf("%x", md5.Sum(body)) //nolint:gosec
	if bodymd5 == f.md5 {
		return nil
	}

	l.Infof("try refresh filter %q", body)

	// try refresh conditions
	var filters Filters
	if err := json.Unmarshal(body, &filters); err != nil {
		l.Errorf("json.Unmarshal: %v", err)
		return err
	}

	defer func() {
		filtersUpdateCount.Inc()
		lastUpdate.Set(float64(time.Now().Unix()))
	}()

	if filters.PullInterval > 0 && f.pullInterval != filters.PullInterval {
		l.Infof("set pull interval from %s to %s", f.pullInterval, filters.PullInterval)
		f.pullInterval = filters.PullInterval
		f.tick.Reset(f.pullInterval)
	}

	f.md5 = bodymd5
	// Clear old conditions: we refresh all conditions if any changed(new/delete
	// conditons or refresh old conditions)
	f.conditions = map[string]parser.WhereConditions{}
	for k, v := range filters.Filters {
		conds, err := GetConds(v)
		if err != nil {
			l.Errorf("GetConds failed: %v", err)
			return err
		}
		f.conditions[k] = conds

		l.Debugf("set raw filter conditions %v on %s", v, k)
		f.rawConditions[k] = strings.Join(v, " ")
	}

	if err := dump(body, f.dumpDir); err != nil {
		l.Warnf("dump: %s, ignored", err)
	}
	return nil
}

func dump(rules []byte, dir string) error {
	return ioutil.WriteFile(filepath.Join(dir, ".pull"), rules, os.ModePerm)
}
