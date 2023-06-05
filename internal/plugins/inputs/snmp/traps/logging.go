// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package traps

import (
	"github.com/gosnmp/gosnmp"
)

// trapLogger is a GoSNMP logger interface implementation.
type trapLogger struct {
	gosnmp.Logger
}

// NOTE: GoSNMP logs show the full content of decoded trap packets. Logging as DEBUG would be too noisy.
func (logger *trapLogger) Print(v ...interface{}) {
	l.Debug(v...)
}

func (logger *trapLogger) Printf(format string, v ...interface{}) {
	l.Debugf(format, v...)
}
