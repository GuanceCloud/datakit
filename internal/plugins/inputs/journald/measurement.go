// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package journald

import (
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type journalMeasurement struct{}

//nolint:lll,funlen
func (*journalMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   inputName,
		Cat:    point.Logging,
		Desc:   "Systemd journal logs. Note: Field availability varies by systemd version - refer to version hints (e.g., v188+, v205+) in each field description",
		DescZh: "Systemd 日志。注意：字段的可用性因 systemd 版本而异 - 请参阅每个字段描述中的版本提示（例如 v188+、v205+）",
		Tags: map[string]any{
			"service": inputs.NewTagInfo("Service identifier (from `SYSLOG_IDENTIFIER`, `_SYSTEMD_UNIT`, or `_COMM`)"),
			"host":    inputs.NewTagInfo("Hostname (from `_HOSTNAME`, v188+)"),
		},
		Fields: map[string]any{
			// NOTE: systemd version compatibility varies by distribution and release.
			// Each field includes a version hint (e.g., v188+, v205+) indicating minimum systemd version.
			// Actual availability depends on your system's systemd version (check with: systemd --version).
			// Fields marked with address field (e.g., __CURSOR, __SEQNUM) are export-only and not available during collection.
			// =================================================================
			// Renamed fields (from journald field → measurement field)
			// =================================================================
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Log message content (from `MESSAGE`, v188+)",
			},
			"priority": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.NoUnit,
				Desc:     "Numeric priority level 0-7 (from `PRIORITY`, v188+)",
			},
			"status": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Log status level mapped from priority: `error`, `warn`, `critical`, `notice`, `info`, `debug`, `unknown`",
			},
			"pid": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.NoUnit,
				Desc:     "Process ID (from `_PID` or `SYSLOG_PID`, v188+)",
			},
			"journald_timestamp": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.DurationNS,
				Desc:     "Journal entry timestamp in nanoseconds (from `_SOURCE_REALTIME_TIMESTAMP` or `__REALTIME_TIMESTAMP`, v188+)",
			},

			// =================================================================
			// Core message fields (v188+)
			// =================================================================
			"MESSAGE_ID": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "128-bit message identifier (`UUID` format, v188+)",
			},
			"CODE_FILE": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Source code filename for debugging (v188+)",
			},
			"CODE_LINE": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.NoUnit,
				Desc:     "Source code line number for debugging (v188+)",
			},
			"CODE_FUNC": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Function name for debugging (v188+)",
			},
			"ERRNO": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.NoUnit,
				Desc:     "Unix error number associated with message (v188+)",
			},

			// =================================================================
			// Syslog compatibility fields
			// =================================================================
			"SYSLOG_FACILITY": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.NoUnit,
				Desc:     "Syslog facility 0-23 (v188+)",
			},
			"SYSLOG_PID": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.NoUnit,
				Desc:     "Client PID from syslog, may differ from `_PID` (v188+)",
			},
			"SYSLOG_TIMESTAMP": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Original syslog timestamp as received (v188+)",
			},
			"SYSLOG_RAW": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Original syslog line if `MESSAGE` modified or timestamp lost (v240+)",
			},

			// =================================================================
			// Process identification fields (trusted, v188+ unless noted)
			// =================================================================
			"_UID": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.NoUnit,
				Desc:     "User ID, trusted cannot be spoofed (v188+)",
			},
			"_GID": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.NoUnit,
				Desc:     "Group ID, trusted (v188+)",
			},
			"_COMM": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Command name truncated to 15 chars (v188+)",
			},
			"_EXE": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Executable path, full path (v188+)",
			},
			"_CMDLINE": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Full command line, most complete process info (v188+)",
			},
			"_CAP_EFFECTIVE": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.NoUnit,
				Desc:     "Effective capabilities bitmask (v206+)",
			},

			// =================================================================
			// Audit fields (trusted, v188+)
			// =================================================================
			"_AUDIT_SESSION": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.NoUnit,
				Desc:     "Audit session ID from kernel (v188+)",
			},
			"_AUDIT_LOGINUID": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.NoUnit,
				Desc:     "Login UID from kernel audit (v188+)",
			},

			// =================================================================
			// Systemd unit fields (trusted)
			// =================================================================
			"_SYSTEMD_CGROUP": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Control group path (v188+)",
			},
			"_SYSTEMD_SLICE": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Slice unit name e.g. `system.slice` (v188+)",
			},
			"_SYSTEMD_UNIT": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Unit name e.g. `sshd.service` (v188+)",
			},
			"_SYSTEMD_USER_UNIT": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "User unit name for user sessions (v188+)",
			},
			"_SYSTEMD_USER_SLICE": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "User slice name e.g. `user.slice` (v188+)",
			},
			"_SYSTEMD_SESSION": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Login session ID (v188+)",
			},
			"_SYSTEMD_OWNER_UID": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.NoUnit,
				Desc:     "Session owner UID (v188+)",
			},
			"_SYSTEMD_INVOCATION_ID": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Unit invocation ID unique per unit start (v233+)",
			},
			"UNIT": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Unit name user-provided alternative to `_SYSTEMD_UNIT` (v251+)",
			},
			"USER_UNIT": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "User unit user-provided alternative to `_SYSTEMD_USER_UNIT` (v251+)",
			},

			// =================================================================
			// Host/boot identification fields (trusted)
			// =================================================================
			"_MACHINE_ID": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Machine ID from `/etc/machine-id` (v188+)",
			},
			"_BOOT_ID": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Boot ID 128-bit hex `UUID` (v188+)",
			},
			"_RUNTIME_SCOPE": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Runtime scope: `initrd`, `system`, or `user` (v252+)",
			},
			"_TRANSPORT": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "How entry was received: `audit`, `driver`, `syslog`, `journal`, `stdout`, `kernel` (v205+)",
			},

			// =================================================================
			// Timestamp fields
			// =================================================================
			"_SOURCE_REALTIME_TIMESTAMP": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.DurationUS,
				Desc:     "Source timestamp in microseconds `CLOCK_REALTIME` (v188+)",
			},
			"_SOURCE_BOOTTIME_TIMESTAMP": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.DurationUS,
				Desc:     "Boottime timestamp in microseconds `CLOCK_BOOTTIME` (v257+)",
			},
			"__REALTIME_TIMESTAMP": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.DurationUS,
				Desc:     "Reception timestamp in microseconds, address field export only (v188+)",
			},
			"__MONOTONIC_TIMESTAMP": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.DurationUS,
				Desc:     "Monotonic timestamp in microseconds, address field export only (v188+)",
			},
			"__CURSOR": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Entry cursor, address field export only (v188+)",
			},
			"__SEQNUM": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.NoUnit,
				Desc:     "Sequence number, address field export only (v254+)",
			},
			"__SEQNUM_ID": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Sequence ID, address field export only (v254+)",
			},

			// =================================================================
			// Stream fields (for stdout transport, v235+)
			// =================================================================
			"_STREAM_ID": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Stream connection ID 128-bit `UUID` for stdout streams (v235+)",
			},
			"_LINE_BREAK": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Line termination info: `nul`, `line-max`, `eof`, `pid-change` (v235+)",
			},

			// =================================================================
			// Kernel/device fields
			// =================================================================
			"_KERNEL_DEVICE": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Kernel device name format: `bM:N`, `cM:N`, `nN`, `+subsys:name` (v189+)",
			},
			"_KERNEL_SUBSYSTEM": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Kernel subsystem e.g. `block`, `net` (v189+)",
			},
			"_UDEV_SYSNAME": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Device name in /sys/ (v189+)",
			},
			"_UDEV_DEVNODE": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Device node in /dev/ full path (v189+)",
			},
			"_UDEV_DEVLINK": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Symlinks to device, can appear multiple times (v189+)",
			},
			"_SELINUX_CONTEXT": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "SELinux security context label (v188+)",
			},

			// =================================================================
			// Coredump fields (systemd-coredump)
			// =================================================================
			"COREDUMP_UNIT": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "System unit that crashed (v198+)",
			},
			"COREDUMP_USER_UNIT": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "User unit that crashed (v198+)",
			},
			"COREDUMP_PID": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.NoUnit,
				Desc:     "Crashed process PID (v188+)",
			},
			"COREDUMP_UID": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.NoUnit,
				Desc:     "Crashed process UID (v188+)",
			},
			"COREDUMP_GID": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.NoUnit,
				Desc:     "Crashed process GID (v188+)",
			},
			"COREDUMP_SIGNAL": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.NoUnit,
				Desc:     "Signal number that caused crash (v188+)",
			},
			"COREDUMP_TIMESTAMP": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.DurationUS,
				Desc:     "Crash timestamp in microseconds (v188+)",
			},
			"COREDUMP_EXE": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Executable path of crashed binary (v188+)",
			},
			"COREDUMP_CMDLINE": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Full command line at crash time (v188+)",
			},
			"COREDUMP_CWD": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Current working directory at crash time (v188+)",
			},
			"COREDUMP_ROOT": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Root directory, usually / (v188+)",
			},
			"COREDUMP_HOSTNAME": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Hostname at crash time (v188+)",
			},
			"COREDUMP_STACKTRACE": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Full stack trace backtrace (v188+)",
			},

			// =================================================================
			// Object fields (for logging on behalf of other processes)
			// =================================================================
			"OBJECT_PID": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.NoUnit,
				Desc:     "Target process PID, requires UID 0 to set (v205+)",
			},
			"OBJECT_UID": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.NoUnit,
				Desc:     "Target process UID (v205+)",
			},
			"OBJECT_GID": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.NoUnit,
				Desc:     "Target process GID (v205+)",
			},
			"OBJECT_COMM": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Target process comm (v205+)",
			},
			"OBJECT_EXE": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Target process executable path (v205+)",
			},
			"OBJECT_CMDLINE": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Target process full command line (v205+)",
			},
			"OBJECT_AUDIT_SESSION": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.NoUnit,
				Desc:     "Target audit session ID (v205+)",
			},
			"OBJECT_AUDIT_LOGINUID": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.NoUnit,
				Desc:     "Target login UID (v205+)",
			},
			"OBJECT_SYSTEMD_CGROUP": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Target cgroup path (v205+)",
			},
			"OBJECT_SYSTEMD_SESSION": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Target session ID (v205+)",
			},
			"OBJECT_SYSTEMD_OWNER_UID": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.NoUnit,
				Desc:     "Target session owner UID (v205+)",
			},
			"OBJECT_SYSTEMD_UNIT": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Target unit name (v205+)",
			},
			"OBJECT_SYSTEMD_USER_UNIT": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Target user unit name (v205+)",
			},
			"OBJECT_SYSTEMD_INVOCATION_ID": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Target invocation ID (v235+)",
			},

			// =================================================================
			// Container fields (v205+)
			// =================================================================
			"_CONTAINER_ID": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Container ID for nspawn/containers (v205+)",
			},
			"_CONTAINER_NAME": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Container name for nspawn/containers (v205+)",
			},
			"_CONTAINER_IMAGE": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Container image for nspawn/containers (v205+)",
			},

			// =================================================================
			// Documentation & metadata fields
			// =================================================================
			"DOCUMENTATION": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Documentation URL http/https/file/man/info (v246+)",
			},
			"TID": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.NoUnit,
				Desc:     "Thread ID numeric (v247+)",
			},
			"INVOCATION_ID": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Invocation ID for systemd code messages (v245+)",
			},
			"USER_INVOCATION_ID": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "User invocation ID for user manager messages (v245+)",
			},

			// =================================================================
			// Namespace fields (v245+)
			// =================================================================
			"_NAMESPACE": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.NoUnit,
				Desc:     "Journal namespace ID (v245+)",
			},

			// =================================================================
			// Additional dynamic fields from journal entry (not excluded via exclude_fields config)
			// =================================================================
		},
	}
}
