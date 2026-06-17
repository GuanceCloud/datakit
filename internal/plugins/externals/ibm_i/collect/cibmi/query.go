// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cibmi

import (
	"fmt"
	"strings"
)

const (
	queryDiskUsage       = "disk_usage"
	queryCPUUsage        = "cpu_usage"
	queryJobQJobStatus   = "jobq_job_status"
	queryActiveJobStatus = "active_job_status"
	queryJobMemoryUsage  = "job_memory_usage"
	queryMemoryInfo      = "memory_info"
	querySubsystem       = "subsystem"
	queryJobQueue        = "job_queue"
	queryMessageQueue    = "message_queue_info"
)

var defaultQueries = []string{
	queryDiskUsage,
	queryCPUUsage,
	queryJobQJobStatus,
	queryActiveJobStatus,
	queryJobMemoryUsage,
	queryMemoryInfo,
	querySubsystem,
	queryJobQueue,
	queryMessageQueue,
}

const systemInfoSQL = "SELECT HOST_NAME, OS_VERSION, OS_RELEASE FROM SYSIBMADM.ENV_SYS_INFO"

const baseDiskUsage72SQL = `
SELECT DISTINCT
  ASP_NUMBER,
  UNIT_NUMBER,
  UNIT_TYPE,
  UNIT_STORAGE_CAPACITY,
  UNIT_SPACE_AVAILABLE,
  PERCENT_USED
FROM QSYS2.SYSDISKSTAT
`

const baseDiskUsage73SQL = `
SELECT DISTINCT
  ASP_NUMBER,
  UNIT_NUMBER,
  UNIT_TYPE,
  SERIAL_NUMBER,
  UNIT_STORAGE_CAPACITY,
  UNIT_SPACE_AVAILABLE,
  PERCENT_USED
FROM QSYS2.SYSDISKSTAT
`

const diskUsageSQL = `
SELECT
  A.ASP_NUMBER,
  A.UNIT_NUMBER,
  A.UNIT_TYPE,
  A.SERIAL_NUMBER,
  A.RESOURCE_NAME,
  A.ELAPSED_PERCENT_BUSY,
  A.ELAPSED_IO_REQUESTS
FROM TABLE(QSYS2.SYSDISKSTAT('NO')) A
INNER JOIN TABLE(QSYS2.SYSDISKSTAT('YES')) B
  ON A.ASP_NUMBER = B.ASP_NUMBER
 AND A.UNIT_NUMBER = B.UNIT_NUMBER
 AND A.RESOURCE_NAME = B.RESOURCE_NAME
`

const cpuUsageSQL = `
SELECT
  A.AVERAGE_CPU_UTILIZATION,
  A.CONFIGURED_CPUS,
  A.CURRENT_CPU_CAPACITY,
  A.PARTITION_ID,
  A.ELAPSED_CPU_SHARED,
  (A.ELAPSED_CPU_USED * A.CURRENT_CPU_CAPACITY) / A.CONFIGURED_CPUS AS NORMALIZED_CPU_USAGE
FROM TABLE(QSYS2.SYSTEM_STATUS('NO')) A
INNER JOIN TABLE(QSYS2.SYSTEM_STATUS('YES')) B
  ON A.PARTITION_ID = B.PARTITION_ID
`

const jobQJobStatusSQL = `
SELECT
  SUBSTR(JOB_NAME,1,POSSTR(JOB_NAME,'/')-1) AS JOB_ID,
  SUBSTR(JOB_NAME,POSSTR(JOB_NAME,'/')+1,POSSTR(SUBSTR(JOB_NAME,POSSTR(JOB_NAME,'/')+1),'/')-1) AS JOB_USER,
  SUBSTR(SUBSTR(JOB_NAME,POSSTR(JOB_NAME,'/')+1),POSSTR(SUBSTR(JOB_NAME,POSSTR(JOB_NAME,'/')+1),'/')+1) AS JOB_NAME,
  JOB_SUBSYSTEM,
  'JOBQ' AS JOB_STATUS,
  JOB_QUEUE_LIBRARY,
  JOB_QUEUE_NAME,
  JOB_QUEUE_STATUS,
  1 AS STATUS,
  (DAYS(CURRENT TIMESTAMP) - DAYS(JOB_QUEUE_TIME)) * 86400
    + MIDNIGHT_SECONDS(CURRENT TIMESTAMP) - MIDNIGHT_SECONDS(JOB_QUEUE_TIME) AS JOBQ_DURATION
FROM TABLE(QSYS2.JOB_INFO('*JOBQ', '*ALL', '*ALL', '*ALL', '*ALL'))
`

const activeJobStatusSQL = `
SELECT
  SUBSTR(A.JOB_NAME,1,POSSTR(A.JOB_NAME,'/')-1) AS JOB_ID,
  SUBSTR(A.JOB_NAME,POSSTR(A.JOB_NAME,'/')+1,POSSTR(SUBSTR(A.JOB_NAME,POSSTR(A.JOB_NAME,'/')+1),'/')-1) AS JOB_USER,
  SUBSTR(SUBSTR(A.JOB_NAME,POSSTR(A.JOB_NAME,'/')+1),POSSTR(SUBSTR(A.JOB_NAME,POSSTR(A.JOB_NAME,'/')+1),'/')+1) AS JOB_NAME,
  A.SUBSYSTEM,
  'ACTIVE' AS JOB_STATUS,
  A.JOB_STATUS AS JOB_ACTIVE_STATUS,
  1 AS STATUS,
  CASE WHEN A.ELAPSED_TIME = 0 THEN 0 ELSE A.ELAPSED_CPU_TIME / (10 * A.ELAPSED_TIME) END AS CPU_USAGE,
  A.ELAPSED_CPU_PERCENTAGE AS CPU_USAGE_PCT,
  (DAYS(CURRENT TIMESTAMP) - DAYS(A.JOB_ACTIVE_TIME)) * 86400
    + MIDNIGHT_SECONDS(CURRENT TIMESTAMP) - MIDNIGHT_SECONDS(A.JOB_ACTIVE_TIME) AS ACTIVE_DURATION
FROM TABLE(QSYS2.ACTIVE_JOB_INFO('NO', '', '', '', 'ALL')) A
INNER JOIN TABLE(QSYS2.ACTIVE_JOB_INFO('YES', '', '', '')) B
  ON A.INTERNAL_JOB_ID = B.INTERNAL_JOB_ID
`

const jobMemoryUsageSQL = `
SELECT
  SUBSTR(JOB_NAME,1,POSSTR(JOB_NAME,'/')-1) AS JOB_ID,
  SUBSTR(JOB_NAME,POSSTR(JOB_NAME,'/')+1,POSSTR(SUBSTR(JOB_NAME,POSSTR(JOB_NAME,'/')+1),'/')-1) AS JOB_USER,
  SUBSTR(SUBSTR(JOB_NAME,POSSTR(JOB_NAME,'/')+1),POSSTR(SUBSTR(JOB_NAME,POSSTR(JOB_NAME,'/')+1),'/')+1) AS JOB_NAME,
  SUBSYSTEM,
  JOB_STATUS,
  MEMORY_POOL,
  TEMPORARY_STORAGE
FROM TABLE(QSYS2.ACTIVE_JOB_INFO('NO', '', '', ''))
`

const memoryInfoSQL = `
SELECT
  POOL_NAME,
  SUBSYSTEM_NAME,
  CURRENT_SIZE,
  RESERVED_SIZE,
  DEFINED_SIZE
FROM QSYS2.MEMORY_POOL_INFO
`

const subsystemSQL = `
SELECT
  SUBSYSTEM_DESCRIPTION,
  CASE WHEN STATUS = 'ACTIVE' THEN 1 ELSE 0 END AS ACTIVE,
  CURRENT_ACTIVE_JOBS
FROM QSYS2.SUBSYSTEM_INFO
`

const jobQueueSQL = `
SELECT
  JOB_QUEUE_NAME,
  JOB_QUEUE_STATUS,
  SUBSYSTEM_NAME,
  NUMBER_OF_JOBS,
  RELEASED_JOBS,
  SCHEDULED_JOBS,
  HELD_JOBS
FROM QSYS2.JOB_QUEUE_INFO
`

func messageQueueSQL(severityThreshold int, selectedQueues []string) (string, []any) {
	var where string
	args := make([]any, 0, len(selectedQueues))
	if len(selectedQueues) > 0 {
		where = "WHERE MESSAGE_QUEUE_NAME IN (" + placeholders(len(selectedQueues)) + ") "
		for _, q := range selectedQueues {
			args = append(args, strings.ToUpper(strings.TrimSpace(q)))
		}
	}

	return fmt.Sprintf(`
SELECT
  MESSAGE_QUEUE_NAME,
  MESSAGE_QUEUE_LIBRARY,
  COUNT(*) AS MESSAGE_QUEUE_SIZE,
  SUM(CASE WHEN SEVERITY >= %d THEN 1 ELSE 0 END) AS CRITICAL_MESSAGE_QUEUE_SIZE
FROM QSYS2.MESSAGE_QUEUE_INFO
%sGROUP BY MESSAGE_QUEUE_NAME, MESSAGE_QUEUE_LIBRARY
`, severityThreshold, where), args
}

func normalizeQueries(values []string) ([]string, error) {
	if len(values) == 0 {
		return append([]string{}, defaultQueries...), nil
	}

	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, raw := range values {
		name := strings.ToLower(strings.TrimSpace(raw))
		if name == "" {
			continue
		}
		if !knownQuery(name) {
			return nil, fmt.Errorf("unknown query %q", raw)
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	if len(out) == 0 {
		return append([]string{}, defaultQueries...), nil
	}
	return out, nil
}

func knownQuery(name string) bool {
	for _, query := range defaultQueries {
		if name == query {
			return true
		}
	}
	return false
}

func placeholders(count int) string {
	return strings.TrimSuffix(strings.Repeat("?,", count), ",")
}
