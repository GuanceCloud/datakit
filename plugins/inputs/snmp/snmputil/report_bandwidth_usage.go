// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmputil

import "fmt"

var bandwidthMetricNameToUsage = map[string]string{
	"ifHCInOctets":  "ifBandwidthInUsage",
	"ifHCOutOctets": "ifBandwidthOutUsage",
}

const ifHighSpeedOID = "1.3.6.1.2.1.31.1.1.1.15"

func trySendBandwidthUsageMetric(symbol SymbolConfig, fullIndex string, values *ResultValueStore, tags []string, outData *MetricDatas) {
	err := sendBandwidthUsageMetric(symbol, fullIndex, values, tags, outData)
	if err != nil {
		l.Debugf("failed to send bandwidth usage metric: %v", err)
	}
}

//nolint:lll
/* sendBandwidthUsageMetric evaluate and report input/output bandwidth usage.
   If any of `ifHCInOctets`, `ifHCOutOctets`  or `ifHighSpeed` is missing then bandwidth will not be reported.

   Bandwidth usage is:

   interface[In|Out]Octets(t+dt) - interface[In|Out]Octets(t)
   ----------------------------------------------------------
                   dt*interfaceSpeed

   Given:
   * ifHCInOctets: the total number of octets received on the interface.
   * ifHCOutOctets: The total number of octets transmitted out of the interface.
   * ifHighSpeed: An estimate of the interface's current bandwidth in Mb/s (10^6 bits
                  per second). It is constant in time, can be overwritten by the system admin.
                  It is the total available bandwidth.
   Bandwidth usage is evaluated as: ifHC[In|Out]Octets/ifHighSpeed and reported as *Rate*
*/
func sendBandwidthUsageMetric(symbol SymbolConfig, fullIndex string, values *ResultValueStore, tags []string, outData *MetricDatas) error {
	usageName, ok := bandwidthMetricNameToUsage[symbol.Name]
	if !ok {
		return nil
	}

	ifHighSpeedValues, err := values.GetColumnValues(ifHighSpeedOID)
	if err != nil {
		return fmt.Errorf("bandwidth usage: missing `ifHighSpeed` metric, skipping metric. fullIndex=%s", fullIndex)
	}

	metricValues, err := getColumnValueFromSymbol(values, symbol)
	if err != nil {
		return fmt.Errorf("bandwidth usage: missing `%s` metric, skipping this row. fullIndex=%s", symbol.Name, fullIndex)
	}

	octetsValue, ok := metricValues[fullIndex]
	if !ok {
		return fmt.Errorf("bandwidth usage: missing value for `%s` metric, skipping this row. fullIndex=%s", symbol.Name, fullIndex)
	}

	ifHighSpeedValue, ok := ifHighSpeedValues[fullIndex]
	if !ok {
		return fmt.Errorf("bandwidth usage: missing value for `ifHighSpeed`, skipping this row. fullIndex=%s", fullIndex)
	}

	ifHighSpeedFloatValue, err := ifHighSpeedValue.ToFloat64()
	if err != nil {
		return fmt.Errorf("failed to convert ifHighSpeedValue to float64: %v", err) //nolint:errorlint
	}
	if ifHighSpeedFloatValue == 0.0 {
		// return fmt.Errorf("bandwidth usage: zero or invalid value for ifHighSpeed, skipping this row. fullIndex=%s, ifHighSpeedValue=%#v", fullIndex, ifHighSpeedValue)
		return nil
	}
	octetsFloatValue, err := octetsValue.ToFloat64()
	if err != nil {
		return fmt.Errorf("failed to convert octetsValue to float64: %v", err) //nolint:errorlint
	}
	usageValue := ((octetsFloatValue * 8) / (ifHighSpeedFloatValue * (1e6))) * 100.0

	sample := MetricSample{
		value:      ResultValue{SubmissionType: "counter", Value: usageValue},
		tags:       tags,
		symbol:     SymbolConfig{Name: usageName + ".rate"},
		forcedType: "counter",
		options:    MetricsConfigOption{},
	}

	sendMetric(sample, outData)
	return nil
}
