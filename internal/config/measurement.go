// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import "strings"

// NormalizeMeasurementVersion normalizes and validates global measurement version.
func NormalizeMeasurementVersion(version string) string {
	version = strings.ToLower(strings.TrimSpace(version))
	switch version {
	case "v1", "v2":
		return version
	default:
		return ""
	}
}

// NormalizeInputMeasurementVersion normalizes collector measurement version.
func NormalizeInputMeasurementVersion(version string) string {
	if strings.ToLower(strings.TrimSpace(version)) == "v1" {
		return "v1"
	}
	return "v2"
}

func (c *Config) setupMeasurementVersion() {
	version := NormalizeMeasurementVersion(c.MeasurementVersion)
	if strings.TrimSpace(c.MeasurementVersion) != "" && version == "" {
		l.Warnf("invalid measurement_version %q, ignored", c.MeasurementVersion)
	}
	c.MeasurementVersion = version
}

// IsOverrideMeasurement checks if the effective measurement version requires override.
func IsOverrideMeasurement(inputVersion string) bool {
	version := NormalizeInputMeasurementVersion(inputVersion)
	if Cfg != nil {
		if Cfg.MeasurementVersion != "" {
			version = Cfg.MeasurementVersion
		}
	}

	return IsOverrideMeasurementVersion(version)
}

// IsOverrideMeasurementVersion checks if the measurement version requires override.
func IsOverrideMeasurementVersion(version string) bool {
	return strings.ToLower(strings.TrimSpace(version)) == "v2"
}
