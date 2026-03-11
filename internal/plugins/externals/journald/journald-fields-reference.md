# Journald Fields Complete Reference

## Overview

This document provides a comprehensive reference of all journald fields with their version history, collected from official systemd documentation and changelogs.

## Core Message Fields (Available since v188/earliest)

| Field | Added In | Description | Type | Notes |
|-------|----------|-------------|------|-------|
| `MESSAGE` | v188 | Human-readable message string | string | Primary field, can contain newlines |
| `MESSAGE_ID` | v188 | 128-bit message identifier | string | UUID-format recommended |
| `PRIORITY` | v188 | Syslog priority (0-7) | int | 0=emerg, 7=debug |
| `CODE_FILE` | v188 | Source filename | string | Debugging info |
| `CODE_LINE` | v188 | Source line number | int | Debugging info |
| `CODE_FUNC` | v188 | Function name | string | Debugging info |
| `ERRNO` | v188 | Unix error number | int | Added in v188 |

## Syslog Compatibility Fields

| Field | Added In | Description | Type | Notes |
|-------|----------|-------------|------|-------|
| `SYSLOG_FACILITY` | v188 | Syslog facility (0-23) | int | Decimal string |
| `SYSLOG_IDENTIFIER` | v188 | Program identifier/tag | string | From glibc `program_invocation_short_name` |
| `SYSLOG_PID` | v188 | Client PID from syslog | int | May differ from `_PID` |
| `SYSLOG_TIMESTAMP` | v188 | Original syslog timestamp | string | As received |
| `SYSLOG_RAW` | v240 | Original syslog line | string | Only if MESSAGE modified or timestamp lost |

## Process Identification Fields (Trusted)

| Field | Added In | Description | Type | Notes |
|-------|----------|-------------|------|-------|
| `_PID` | v188 | Process ID | int | Trusted, cannot be spoofed |
| `_UID` | v188 | User ID | int | Trusted |
| `_GID` | v188 | Group ID | int | Trusted |
| `_COMM` | v188 | Command name (comm) | string | Truncated to 15 chars |
| `_EXE` | v188 | Executable path | string | Full path |
| `_CMDLINE` | v188 | Full command line | string | Most complete process info |
| `_CAP_EFFECTIVE` | v206 | Effective capabilities | int | Bitmask |

## Audit Fields (Trusted)

| Field | Added In | Description | Type | Notes |
|-------|----------|-------------|------|-------|
| `_AUDIT_SESSION` | v188 | Audit session ID | int | From kernel audit |
| `_AUDIT_LOGINUID` | v188 | Login UID | int | From kernel audit |

## Systemd Unit Fields (Trusted)

| Field | Added In | Description | Type | Notes |
|-------|----------|-------------|------|-------|
| `_SYSTEMD_CGROUP` | v188 | Control group path | string | Full cgroup path |
| `_SYSTEMD_SLICE` | v188 | Slice unit name | string | e.g., `system.slice` |
| `_SYSTEMD_UNIT` | v188 | Unit name | string | e.g., `sshd.service` |
| `_SYSTEMD_USER_UNIT` | v188 | User unit name | string | For user sessions |
| `_SYSTEMD_USER_SLICE` | v188 | User slice name | string | e.g., `user.slice` |
| `_SYSTEMD_SESSION` | v188 | Session ID | string | Login session |
| `_SYSTEMD_OWNER_UID` | v188 | Session owner UID | int | User who owns session |
| `_SYSTEMD_INVOCATION_ID` | v233 | Unit invocation ID | string | Unique per unit start |
| `UNIT` | v251 | Unit name (user-provided) | string | Alternative to `_SYSTEMD_UNIT` |
| `USER_UNIT` | v251 | User unit (user-provided) | string | Alternative to `_SYSTEMD_USER_UNIT` |

## Host/Boot Identification Fields (Trusted)

| Field | Added In | Description | Type | Notes |
|-------|----------|-------------|------|-------|
| `_HOSTNAME` | v188 | Hostname | string | System hostname |
| `_MACHINE_ID` | v188 | Machine ID | string | From `/etc/machine-id` |
| `_BOOT_ID` | v188 | Boot ID | string | 128-bit hex UUID |
| `_RUNTIME_SCOPE` | v252 | Runtime scope | string | "initrd", "system", or "user" |
| `_TRANSPORT` | v205 | How entry was received | string | See below for values |

## Transport Type Values (`_TRANSPORT`)

| Value | Added In | Description |
|-------|----------|-------------|
| `audit` | v227 | From kernel audit subsystem |
| `driver` | v205 | Internally generated messages |
| `syslog` | v205 | From local syslog socket |
| `journal` | v205 | From native journal protocol |
| `stdout` | v205 | From service stdout/stderr |
| `kernel` | v205 | From kernel (dmesg) |

## Timestamp Fields

| Field | Added In | Description | Type | Notes |
|-------|----------|-------------|------|-------|
| `_SOURCE_REALTIME_TIMESTAMP` | v188 | Source timestamp (μs) | int | `CLOCK_REALTIME` |
| `_SOURCE_BOOTTIME_TIMESTAMP` | v257 | Boottime timestamp (μs) | int | `CLOCK_BOOTTIME` |
| `__REALTIME_TIMESTAMP` | v188 | Reception time (μs) | int | Address field (export only) |
| `__MONOTONIC_TIMESTAMP` | v188 | Monotonic time (μs) | int | Address field (export only) |
| `__CURSOR` | v188 | Entry cursor | string | Address field (export only) |
| `__SEQNUM` | v254 | Sequence number | int | Address field (export only) |
| `__SEQNUM_ID` | v254 | Sequence ID | string | Address field (export only) |

## Stream Fields (for stdout transport)

| Field | Added In | Description | Type | Notes |
|-------|----------|-------------|------|-------|
| `_STREAM_ID` | v235 | Stream connection ID | string | 128-bit UUID for stdout streams |
| `_LINE_BREAK` | v235 | Line termination info | string | `nul`, `line-max`, `eof`, `pid-change` |

## Kernel/Device Fields

| Field | Added In | Description | Type | Notes |
|-------|----------|-------------|------|-------|
| `_KERNEL_DEVICE` | v189 | Kernel device name | string | Format: `bM:N`, `cM:N`, `nN`, `+subsys:name` |
| `_KERNEL_SUBSYSTEM` | v189 | Kernel subsystem | string | e.g., `block`, `net` |
| `_UDEV_SYSNAME` | v189 | Device name in /sys/ | string | Kernel device name |
| `_UDEV_DEVNODE` | v189 | Device node in /dev/ | string | Full path |
| `_UDEV_DEVLINK` | v189 | Symlinks to device | string | Can appear multiple times |
| `_SELINUX_CONTEXT` | v188 | SELinux security context | string | Label from SELinux |

## Coredump Fields (systemd-coredump)

| Field | Added In | Description | Type | Notes |
|-------|----------|-------------|------|-------|
| `COREDUMP_UNIT` | v198 | System unit that crashed | string | For coredumpctl |
| `COREDUMP_USER_UNIT` | v198 | User unit that crashed | string | For user sessions |
| `COREDUMP_PID` | v188 | Crashed process PID | int | From coredump |
| `COREDUMP_UID` | v188 | Crashed process UID | int | From coredump |
| `COREDUMP_GID` | v188 | Crashed process GID | int | From coredump |
| `COREDUMP_SIGNAL` | v188 | Signal number | int | Crash signal |
| `COREDUMP_TIMESTAMP` | v188 | Crash timestamp | int | Microseconds |
| `COREDUMP_EXE` | v188 | Executable path | string | Crashed binary |
| `COREDUMP_CMDLINE` | v188 | Command line | string | Full cmdline |
| `COREDUMP_CWD` | v188 | Current working directory | string | At crash time |
| `COREDUMP_ROOT` | v188 | Root directory | string | Usually `/` |
| `COREDUMP_HOSTNAME` | v188 | Hostname | string | At crash time |
| `COREDUMP_STACKTRACE` | v188 | Stack trace | string | Full backtrace |

## Object Fields (for logging on behalf of other processes)

| Field | Added In | Description | Type | Notes |
|-------|----------|-------------|------|-------|
| `OBJECT_PID` | v205 | Target process PID | int | Requires UID 0 to set |
| `OBJECT_UID` | v205 | Target process UID | int | Added by journald |
| `OBJECT_GID` | v205 | Target process GID | int | Added by journald |
| `OBJECT_COMM` | v205 | Target process comm | string | Added by journald |
| `OBJECT_EXE` | v205 | Target process exe | string | Added by journald |
| `OBJECT_CMDLINE` | v205 | Target process cmdline | string | Added by journald |
| `OBJECT_AUDIT_SESSION` | v205 | Target audit session | int | Added by journald |
| `OBJECT_AUDIT_LOGINUID` | v205 | Target login UID | int | Added by journald |
| `OBJECT_SYSTEMD_CGROUP` | v205 | Target cgroup | string | Added by journald |
| `OBJECT_SYSTEMD_SESSION` | v205 | Target session | string | Added by journald |
| `OBJECT_SYSTEMD_OWNER_UID` | v205 | Target session owner | int | Added by journald |
| `OBJECT_SYSTEMD_UNIT` | v205 | Target unit | string | Added by journald |
| `OBJECT_SYSTEMD_USER_UNIT` | v205 | Target user unit | string | Added by journald |
| `OBJECT_SYSTEMD_INVOCATION_ID` | v235 | Target invocation ID | string | Added by journald |

## Container Fields

| Field | Added In | Description | Type | Notes |
|-------|----------|-------------|------|-------|
| `_CONTAINER_ID` | v205 | Container ID | string | For nspawn/containers |
| `_CONTAINER_NAME` | v205 | Container name | string | For nspawn/containers |
| `_CONTAINER_IMAGE` | v205 | Container image | string | For nspawn/containers |

## Documentation & Metadata Fields

| Field | Added In | Description | Type | Notes |
|-------|----------|-------------|------|-------|
| `DOCUMENTATION` | v246 | Documentation URL | string | http/https/file/man/info URLs |
| `TID` | v247 | Thread ID | int | Numeric thread ID |
| `INVOCATION_ID` | v245 | Invocation ID | string | For systemd code messages |
| `USER_INVOCATION_ID` | v245 | User invocation ID | string | For user manager messages |

## Namespace Fields

| Field | Added In | Description | Type | Notes |
|-------|----------|-------------|------|-------|
| `_NAMESPACE` | v245 | Journal namespace ID | string | For journal namespaces |

## Recommendation for PID/UID Field Handling

### Current Implementation is CORRECT ✅

The journald collector stores:
- `_PID` → `pid` (numeric)
- `_UID` → stored as-is (numeric)
- `_CMDLINE` → already captured from journal

### DO NOT Convert to Names Because:

1. **Historical accuracy**: Journal entries persist after process exit. PID lookup will fail for old entries.
2. **Race conditions**: Process can exit between journal write and collection.
3. **UID/GID stability**: Users can be deleted/renamed, but numeric IDs remain accurate.
4. **Industry standard**: systemd journal, rsyslog, Prometheus, ELK stack all store numeric values.
5. **`_CMDLINE` already exists**: Full command line is captured at log time (most useful field).
6. **Query-time enrichment**: If needed, can enrich at query time via DataKit pipelines.

### Optional Enhancement (If Requested)

Add a configuration flag for optional enrichment:
```go
// Option struct
EnrichProcessInfo bool `mapstructure:"enrich-process-info"`

// In entryToPoints
if ipt.config.EnrichProcessInfo {
    // Best-effort lookup via getpwuid(), /proc/PID/cmdline
    // Only for live processes
}
```

**Default: `false`** (keep numeric, fast, accurate)

## References

- [systemd.journal-fields(7)](https://www.freedesktop.org/software/systemd/man/systemd.journal-fields.html)
- [systemd.journal-file(5)](https://www.freedesktop.org/software/systemd/man/systemd.journal-file.html)
- [sd-journal(3)](https://www.freedesktop.org/software/systemd/man/sd-journal.html)
- systemd/NEWS changelog (v188-v260)
