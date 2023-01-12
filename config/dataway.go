// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
)

func (c *Config) SetupDataway() error {
	if c.DataWayCfg == nil {
		return fmt.Errorf("dataway config is empty")
	}

	// 如果 env 已传入了 dataway 配置, 则不再追加老的 dataway 配置,
	// 避免俩边配置了同样的 dataway, 造成数据混乱
	if c.DataWayCfg.DeprecatedURL != "" && len(c.DataWayCfg.URLs) == 0 {
		c.DataWayCfg.URLs = []string{c.DataWayCfg.DeprecatedURL}
	}

	if len(c.DataWayCfg.URLs) > 0 && c.DataWayCfg.URLs[0] == datakit.DatawayDisableURL {
		c.RunMode = datakit.ModeDev
		return nil
	} else {
		c.RunMode = datakit.ModeNormal
	}

	dataway.ExtraHeaders = map[string]string{
		"X-Datakit-Info": fmt.Sprintf("%s; %s", c.Hostname, datakit.Version),
	}

	c.DataWay = &dataway.DataWayDefault{}

	c.DataWayCfg.Hostname = c.Hostname
	if err := c.DataWay.Init(c.DataWayCfg); err != nil {
		c.DataWay = nil
		return err
	}

	return nil
}
