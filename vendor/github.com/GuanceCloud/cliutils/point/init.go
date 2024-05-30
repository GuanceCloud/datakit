// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//nolint:gochecknoinits
package point

import "github.com/influxdata/influxdb1-client/models"

func init() {
	loadEnvs()

	// Enable unsigned int for line-protocol, we do not use
	// influxdb 1.x any more.
	models.EnableUintSupport()

	// add more...
}
