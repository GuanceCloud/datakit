// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	influxdb "github.com/influxdata/influxdb1-client/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/snmp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/snmp/snmpmeasurement"
)

func testSNMP(snmpConfFile string) error {
	x, err := config.LoadSingleConfFile(snmpConfFile, inputs.Inputs, false)
	if err != nil {
		cp.Errorf("[LoadSingleConfFile (%s) failed: (%v)\n", *flagCheckSNMP, err)
		return err
	}

	// only sinlge innstance will pass.
	iptsArr, ok := x[snmpmeasurement.InputName]
	if !ok {
		return fmt.Errorf("invalid_conf_file_snmp")
	}
	if length := len(iptsArr); length != 1 {
		cp.Errorf("Instance in conf file has to be ONE: %d\n", length)
		return fmt.Errorf("conf_over_count")
	}

	ipt, ok := iptsArr[0].(*snmp.Input)
	if !ok {
		return fmt.Errorf("invalid_conf_file_input")
	}

	snmp.SetLog()

	if err := ipt.CheckTestSNMP(); err != nil {
		return err
	}

	if err := ipt.ValidateConfig(); err != nil {
		return err
	}

	if err := ipt.Initialize(); err != nil {
		return err
	}

	cp.Infof("Start collecting snmp...\n")

	specificDevices := ipt.GetSpecificDevices()
	for deviceIP, deviceInfo := range specificDevices {
		tn := time.Now().UTC()
		points := ipt.CollectingMeasurements(deviceIP, deviceInfo, tn, true)
		cp.Infof("\n>>>>>>>>>>>>>>>>>> Below is SNMP object (IP: %s) <<<<<<<<<<<<<<<<<<\n", deviceIP)
		if err := printMeasurements(points); err != nil {
			return err
		}

		tn = time.Now().UTC()
		points = ipt.CollectingMeasurements(deviceIP, deviceInfo, tn, false)
		cp.Infof("\n>>>>>>>>>>>>>>>>>> Below is SNMP metrics (IP: %s) <<<<<<<<<<<<<<<<<<\n", deviceIP)
		if err := printMeasurements(points); err != nil {
			return err
		}
	}

	return nil
}

func printMeasurements(pts []*point.Point) error {
	if len(pts) == 0 {
		return fmt.Errorf("measurements_empty")
	}

	return printResult(point2dkpt(pts...))
}

// point2dkpt convert point.Point to old io/point.Point.
func point2dkpt(pts ...*point.Point) (res []*dkpt.Point) {
	for _, pt := range pts {
		pt, err := influxdb.NewPoint(string(pt.Name()), pt.InfluxTags(), pt.InfluxFields(), pt.Time())
		if err != nil {
			continue
		}

		res = append(res, &dkpt.Point{Point: pt})
	}

	return res
}
