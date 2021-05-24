package smart

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/influxdata/telegraf"
)

const intelVID = "0x8086"

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

	deviceFieldIds = map[string]string{
		"1":   "read_error_rate",
		"7":   "seek_error_rate",
		"190": "temp_c",
		"194": "temp_c",
		"199": "udma_crc_errors",
	}

	// to obtain metrics from smartctl
	sasNvmeAttributes = map[string]struct {
		ID    string
		Name  string
		Parse func(fields, deviceFields map[string]interface{}, str string) error
	}{
		"Accumulated start-stop cycles": {
			ID:   "4",
			Name: "Start_Stop_Count",
		},
		"Accumulated load-unload cycles": {
			ID:   "193",
			Name: "Load_Cycle_Count",
		},
		"Current Drive Temperature": {
			ID:    "194",
			Name:  "Temperature_Celsius",
			Parse: parseTemperature,
		},
		"Temperature": {
			ID:    "194",
			Name:  "Temperature_Celsius",
			Parse: parseTemperature,
		},
		"Power Cycles": {
			ID:   "12",
			Name: "Power_Cycle_Count",
		},
		"Power On Hours": {
			ID:   "9",
			Name: "Power_On_Hours",
		},
		"Media and Data Integrity Errors": {
			Name: "Media_and_Data_Integrity_Errors",
		},
		"Error Information Log Entries": {
			Name: "Error_Information_Log_Entries",
		},
		"Critical Warning": {
			Name: "Critical_Warning",
			Parse: func(fields, _ map[string]interface{}, str string) error {
				var value int64
				if _, err := fmt.Sscanf(str, "0x%x", &value); err != nil {
					return err
				}

				fields["raw_value"] = value

				return nil
			},
		},
		"Available Spare": {
			Name:  "Available_Spare",
			Parse: parsePercentageInt,
		},
		"Available Spare Threshold": {
			Name:  "Available_Spare_Threshold",
			Parse: parsePercentageInt,
		},
		"Percentage Used": {
			Name:  "Percentage_Used",
			Parse: parsePercentageInt,
		},
		"Data Units Read": {
			Name:  "Data_Units_Read",
			Parse: parseDataUnits,
		},
		"Data Units Written": {
			Name:  "Data_Units_Written",
			Parse: parseDataUnits,
		},
		"Host Read Commands": {
			Name:  "Host_Read_Commands",
			Parse: parseCommaSeparatedInt,
		},
		"Host Write Commands": {
			Name:  "Host_Write_Commands",
			Parse: parseCommaSeparatedInt,
		},
		"Controller Busy Time": {
			Name:  "Controller_Busy_Time",
			Parse: parseCommaSeparatedInt,
		},
		"Unsafe Shutdowns": {
			Name:  "Unsafe_Shutdowns",
			Parse: parseCommaSeparatedInt,
		},
		"Warning  Comp. Temperature Time": {
			Name:  "Warning_Temperature_Time",
			Parse: parseCommaSeparatedInt,
		},
		"Critical Comp. Temperature Time": {
			Name:  "Critical_Temperature_Time",
			Parse: parseCommaSeparatedInt,
		},
		"Thermal Temp. 1 Transition Count": {
			Name:  "Thermal_Management_T1_Trans_Count",
			Parse: parseCommaSeparatedInt,
		},
		"Thermal Temp. 2 Transition Count": {
			Name:  "Thermal_Management_T2_Trans_Count",
			Parse: parseCommaSeparatedInt,
		},
		"Thermal Temp. 1 Total Time": {
			Name:  "Thermal_Management_T1_Total_Time",
			Parse: parseCommaSeparatedInt,
		},
		"Thermal Temp. 2 Total Time": {
			Name:  "Thermal_Management_T2_Total_Time",
			Parse: parseCommaSeparatedInt,
		},
		"Temperature Sensor 1": {
			Name:  "Temperature_Sensor_1",
			Parse: parseTemperatureSensor,
		},
		"Temperature Sensor 2": {
			Name:  "Temperature_Sensor_2",
			Parse: parseTemperatureSensor,
		},
		"Temperature Sensor 3": {
			Name:  "Temperature_Sensor_3",
			Parse: parseTemperatureSensor,
		},
		"Temperature Sensor 4": {
			Name:  "Temperature_Sensor_4",
			Parse: parseTemperatureSensor,
		},
		"Temperature Sensor 5": {
			Name:  "Temperature_Sensor_5",
			Parse: parseTemperatureSensor,
		},
		"Temperature Sensor 6": {
			Name:  "Temperature_Sensor_6",
			Parse: parseTemperatureSensor,
		},
		"Temperature Sensor 7": {
			Name:  "Temperature_Sensor_7",
			Parse: parseTemperatureSensor,
		},
		"Temperature Sensor 8": {
			Name:  "Temperature_Sensor_8",
			Parse: parseTemperatureSensor,
		},
	}

	// to obtain Intel specific metrics from nvme-cli
	intelAttributes = map[string]struct {
		ID    string
		Name  string
		Parse func(acc telegraf.Accumulator, fields map[string]interface{}, tags map[string]string, str string) error
	}{
		"program_fail_count": {
			Name: "Program_Fail_Count",
		},
		"erase_fail_count": {
			Name: "Erase_Fail_Count",
		},
		"end_to_end_error_detection_count": {
			Name: "End_To_End_Error_Detection_Count",
		},
		"crc_error_count": {
			Name: "Crc_Error_Count",
		},
		"retry_buffer_overflow_count": {
			Name: "Retry_Buffer_Overflow_Count",
		},
		"wear_leveling": {
			Name:  "Wear_Leveling",
			Parse: parseWearLeveling,
		},
		"timed_workload_media_wear": {
			Name:  "Timed_Workload_Media_Wear",
			Parse: parseTimedWorkload,
		},
		"timed_workload_host_reads": {
			Name:  "Timed_Workload_Host_Reads",
			Parse: parseTimedWorkload,
		},
		"timed_workload_timer": {
			Name: "Timed_Workload_Timer",
			Parse: func(acc telegraf.Accumulator, fields map[string]interface{}, tags map[string]string, str string) error {
				return parseCommaSeparatedIntWithAccumulator(acc, fields, tags, strings.TrimSuffix(str, " min"))
			},
		},
		"thermal_throttle_status": {
			Name:  "Thermal_Throttle_Status",
			Parse: parseThermalThrottle,
		},
		"pll_lock_loss_count": {
			Name: "Pll_Lock_Loss_Count",
		},
		"nand_bytes_written": {
			Name:  "Nand_Bytes_Written",
			Parse: parseBytesWritten,
		},
		"host_bytes_written": {
			Name:  "Host_Bytes_Written",
			Parse: parseBytesWritten,
		},
	}
)
