// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package logstreaming

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const defaultMeasurementName = "default"

type logstreamingMeasurement struct{}

//nolint:lll
func (*logstreamingMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Type: "logging",
		Name: defaultMeasurementName,
		Desc: "Using `source` field in the config file, default is `default`.",
		Tags: map[string]interface{}{
			"service":        inputs.NewTagInfo("Service name. Using the `service` parameter in the URL."),
			"ip_or_hostname": inputs.NewTagInfo("Request IP or hostname."),
		},
		Fields: map[string]interface{}{
			"message": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Message text, existed when default. Could use Pipeline to delete this field."},
			"status":  &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Log status."},
		},
	}
}
