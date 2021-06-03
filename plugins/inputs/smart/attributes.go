package smart

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	// Device Model:     APPLE SSD SM256E
	// Product:              HUH721212AL5204
	// Model Number: TS128GMTE850
	modelInfo = regexp.MustCompile("^(Device Model|Product|Model Number):\\s+(.*)$")
	// Serial Number:    S0X5NZBC422720
	serialInfo = regexp.MustCompile("(?i)^Serial Number:\\s+(.*)$")
	// LU WWN Device Id: 5 002538 655584d30
	wwnInfo = regexp.MustCompile("^LU WWN Device Id:\\s+(.*)$")
	// User Capacity:    251,000,193,024 bytes [251 GB]
	userCapacityInfo = regexp.MustCompile("^User Capacity:\\s+([0-9,]+)\\s+bytes.*$")
	// SMART support is: Enabled
	smartEnabledInfo = regexp.MustCompile("^SMART support is:\\s+(\\w+)$")
	// SMART overall-health self-assessment test result: PASSED
	// SMART Health Status: OK
	// PASSED, FAILED, UNKNOWN
	smartOverallHealth = regexp.MustCompile("^(SMART overall-health self-assessment test result|SMART Health Status):\\s+(\\w+).*$")

	// sasNvmeAttr is a SAS or NVME SMART attribute
	sasNvmeAttr = regexp.MustCompile(`^([^:]+):\s+(.+)$`)

	// ID# ATTRIBUTE_NAME          FLAGS    VALUE WORST THRESH FAIL RAW_VALUE
	//   1 Raw_Read_Error_Rate     -O-RC-   200   200   000    -    0
	//   5 Reallocated_Sector_Ct   PO--CK   100   100   000    -    0
	// 192 Power-Off_Retract_Count -O--C-   097   097   000    -    14716
	attribute = regexp.MustCompile("^\\s*([0-9]+)\\s(\\S+)\\s+([-P][-O][-S][-R][-C][-K])\\s+([0-9]+)\\s+([0-9]+)\\s+([0-9-]+)\\s+([-\\w]+)\\s+([\\w\\+\\.]+).*$")

	//  Additional Smart Log for NVME device:nvme0 namespace-id:ffffffff
	//	key                               normalized raw
	//	program_fail_count              : 100%       0
	intelExpressionPattern = regexp.MustCompile(`^([\w\s]+):([\w\s]+)%(.+)`)

	//	vid     : 0x8086
	//	sn      : CFGT53260XSP8011P
	nvmeIDCtrlExpressionPattern = regexp.MustCompile(`^([\w\s]+):([\s\w]+)`)
)

var deviceFieldIds = map[string]string{
	"1":   "read_error_rate",
	"7":   "seek_error_rate",
	"190": "temp_c",
	"194": "temp_c",
	"199": "udma_crc_errors",
}

// to obtain metrics from smartctl
var sasNvmeAttributes = map[string]struct {
	ID    string
	Name  string
	Parse func(key string, fields map[string]interface{}, str string) error
}{
	"Accumulated start-stop cycles": {
		ID:   "4",
		Name: "start_stop_count",
	},
	"Accumulated load-unload cycles": {
		ID:   "193",
		Name: "load_cycle_count",
	},
	"Current Drive Temperature": {
		ID:    "194",
		Name:  "temperature_celsius",
		Parse: parseTemperature,
	},
	"Temperature": {
		ID:    "194",
		Name:  "temperature_celsius",
		Parse: parseTemperature,
	},
	"Power Cycles": {
		ID:   "12",
		Name: "power_cycle_count",
	},
	"Power On Hours": {
		ID:   "9",
		Name: "power_on_hours",
	},
	"Media and Data Integrity Errors": {
		Name: "media_and_data_integrity_errors",
	},
	"Error Information Log Entries": {
		Name: "error_information_log_entries",
	},
	"Critical Warning": {
		Name: "critical_warning",
		Parse: func(key string, fields map[string]interface{}, str string) error {
			var value int64
			if _, err := fmt.Sscanf(str, "0x%x", &value); err != nil {
				return err
			}

			fields["raw_value"] = value

			return nil
		},
	},
	"Available Spare": {
		Name:  "available_spare",
		Parse: parsePercentageInt,
	},
	"Available Spare Threshold": {
		Name:  "available_spare_threshold",
		Parse: parsePercentageInt,
	},
	"Percentage Used": {
		Name:  "percentage_used",
		Parse: parsePercentageInt,
	},
	"Data Units Read": {
		Name:  "data_units_read",
		Parse: parseDataUnits,
	},
	"Data Units Written": {
		Name:  "data_units_written",
		Parse: parseDataUnits,
	},
	"Host Read Commands": {
		Name:  "host_read_commands",
		Parse: parseCommaSeparatedInt,
	},
	"Host Write Commands": {
		Name:  "host_write_commands",
		Parse: parseCommaSeparatedInt,
	},
	"Controller Busy Time": {
		Name:  "controller_busy_time",
		Parse: parseCommaSeparatedInt,
	},
	"Unsafe Shutdowns": {
		Name:  "unsafe_shutdowns",
		Parse: parseCommaSeparatedInt,
	},
	"Warning  Comp. Temperature Time": {
		Name:  "warning_temperature_time",
		Parse: parseCommaSeparatedInt,
	},
	"Critical Comp. Temperature Time": {
		Name:  "critical_temperature_time",
		Parse: parseCommaSeparatedInt,
	},
	"Thermal Temp. 1 Transition Count": {
		Name:  "thermal_management_t1_trans_count",
		Parse: parseCommaSeparatedInt,
	},
	"Thermal Temp. 2 Transition Count": {
		Name:  "thermal_management_t2_trans_count",
		Parse: parseCommaSeparatedInt,
	},
	"Thermal Temp. 1 Total Time": {
		Name:  "thermal_management_t1_total_time",
		Parse: parseCommaSeparatedInt,
	},
	"Thermal Temp. 2 Total Time": {
		Name:  "thermal_management_t2_total_time",
		Parse: parseCommaSeparatedInt,
	},
	"Temperature Sensor 1": {
		Name:  "temperature_sensor_1",
		Parse: parseTemperatureSensor,
	},
	"Temperature Sensor 2": {
		Name:  "temperature_sensor_2",
		Parse: parseTemperatureSensor,
	},
	"Temperature Sensor 3": {
		Name:  "temperature_sensor_3",
		Parse: parseTemperatureSensor,
	},
	"Temperature Sensor 4": {
		Name:  "temperature_sensor_4",
		Parse: parseTemperatureSensor,
	},
	"Temperature Sensor 5": {
		Name:  "temperature_sensor_5",
		Parse: parseTemperatureSensor,
	},
	"Temperature Sensor 6": {
		Name:  "temperature_sensor_6",
		Parse: parseTemperatureSensor,
	},
	"Temperature Sensor 7": {
		Name:  "temperature_sensor_7",
		Parse: parseTemperatureSensor,
	},
	"Temperature Sensor 8": {
		Name:  "temperature_sensor_8",
		Parse: parseTemperatureSensor,
	},
}

func parseTemperature(key string, fields map[string]interface{}, str string) error {
	var temp int64
	if _, err := fmt.Sscanf(str, "%d C", &temp); err != nil {
		return err
	}
	fields[key] = temp

	return nil
}

func parseCommaSeparatedInt(key string, fields map[string]interface{}, str string) error {
	str = strings.Join(strings.Fields(str), "")
	i, err := strconv.ParseInt(strings.Replace(str, ",", "", -1), 10, 64)
	if err != nil {
		return err
	}
	fields[key] = i

	return nil
}

func parsePercentageInt(key string, fields map[string]interface{}, str string) error {
	return parseCommaSeparatedInt(key, fields, strings.TrimSuffix(str, "%"))
}

func parseDataUnits(key string, fields map[string]interface{}, str string) error {
	units := strings.Fields(str)[0]

	return parseCommaSeparatedInt(key, fields, units)
}

func parseTemperatureSensor(key string, fields map[string]interface{}, str string) error {
	var temp int64
	if _, err := fmt.Sscanf(str, "%d C", &temp); err != nil {
		return err
	}

	fields[key] = temp

	return nil
}

// to obtain Intel specific metrics from nvme-cli
var intelAttributes = map[string]struct {
	ID    string
	Name  string
	Parse func(key string, fields map[string]interface{}, str string) error
}{
	"program_fail_count": {
		Name: "Program_fail_count",
	},
	"erase_fail_count": {
		Name: "Erase_fail_count",
	},
	"end_to_end_error_detection_count": {
		Name: "End_to_end_error_detection_count",
	},
	"crc_error_count": {
		Name: "Crc_error_count",
	},
	"retry_buffer_overflow_count": {
		Name: "Retry_buffer_overflow_count",
	},
	"wear_leveling": {
		Name:  "wear_leveling",
		Parse: parseWearLeveling,
	},
	"timed_workload_media_wear": {
		Name:  "timed_workload_media_wear",
		Parse: parseTimedWorkload,
	},
	"timed_workload_host_reads": {
		Name:  "timed_workload_host_reads",
		Parse: parseTimedWorkload,
	},
	"timed_workload_timer": {
		Name: "Timed_workload_timer",
		Parse: func(key string, fields map[string]interface{}, str string) error {
			return parseCommaSeparatedIntWithCache(key, fields, strings.TrimSuffix(str, " min"))
		},
	},
	"thermal_throttle_status": {
		Name:  "thermal_throttle_status",
		Parse: parseThermalThrottle,
	},
	"pll_lock_loss_count": {
		Name: "Pll_lock_loss_count",
	},
	"nand_bytes_written": {
		Name:  "nand_bytes_written",
		Parse: parseBytesWritten,
	},
	"host_bytes_written": {
		Name:  "host_bytes_written",
		Parse: parseBytesWritten,
	},
}

func parseWearLeveling(key string, fields map[string]interface{}, str string) error {
	var min, max, avg int64
	if _, err := fmt.Sscanf(str, "min: %d, max: %d, avg: %d", &min, &max, &avg); err != nil {
		return err
	}
	values := []int64{min, max, avg}
	for i, submetricName := range []string{"Min", "Max", "Avg"} {
		fields[fmt.Sprintf("%s_%s", key, submetricName)] = values[i]
	}

	return nil
}

func parseTimedWorkload(key string, fields map[string]interface{}, str string) error {
	var value float64
	if _, err := fmt.Sscanf(str, "%f", &value); err != nil {
		return err
	}
	fields[key] = value

	return nil
}

func parseThermalThrottle(key string, fields map[string]interface{}, str string) error {
	var (
		percentage float64
		count      int64
	)
	if _, err := fmt.Sscanf(str, "%f%%, cnt: %d", &percentage, &count); err != nil {
		return err
	}
	fields[key+"_Prc"] = percentage
	fields[key+"_Count"] = count

	return nil
}

func parseBytesWritten(key string, fields map[string]interface{}, str string) error {
	var value int64
	if _, err := fmt.Sscanf(str, "sectors: %d", &value); err != nil {
		return err
	}
	fields[key] = value

	return nil
}

func parseCommaSeparatedIntWithCache(key string, fields map[string]interface{}, str string) error {
	i, err := strconv.ParseInt(strings.Replace(str, ",", "", -1), 10, 64)
	if err != nil {
		return err
	}
	fields[key] = i

	return nil
}
