package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/godror/godror"
	"golang.org/x/net/context/ctxhttp"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

var (
	fInterval        = flag.String("interval", "5m", "gather interval")
	fMetric          = flag.String("metric-name", "oracle_monitor", "gathered metric name")
	fInstance        = flag.String("instance-id", "", "oracle instance ID")
	fInstanceDesc    = flag.String("instance-desc", "", "oracle description")
	fHost            = flag.String("host", "", "oracle host")
	fPort            = flag.String("port", "1521", "oracle port")
	fUsername        = flag.String("username", "", "oracle username")
	fPassword        = flag.String("password", "", "oracle password")
	fServiceName     = flag.String("service-name", "", "oracle service name")
	fClusterType     = flag.String("cluster-type", "", "oracle cluster type(single/dg/rac)")
	fOracleVersion   = flag.String("oracle-version", "", "oracle version(support 10g/11g/12c)")
	fTags            = flag.String("tags", "", `additional tags in 'a=b,c=d,...' format`)
	fDatakitHTTPPort = flag.Int("datakit-http-port", 9529, "DataKit HTTP server port")

	fLog      = flag.String("log", filepath.Join(datakit.InstallDir, "externals", "oraclemonitor.log"), "log path")
	fLogLevel = flag.String("log-level", "info", "log file")

	l              *logger.Logger
	datakitPostURL = ""
)

type monitor struct {
	libPath       string
	metric        string
	interval      string
	instanceId    string
	user          string
	password      string
	desc          string
	host          string
	port          string
	serviceName   string
	clusterType   string
	tags          map[string]string
	oracleVersion string

	db               *sql.DB
	intervalDuration time.Duration
}

func buildMonitor() *monitor {
	m := &monitor{
		metric:        *fMetric,
		interval:      *fInterval,
		instanceId:    *fInstance,
		user:          *fUsername,
		password:      *fPassword,
		desc:          *fInstanceDesc,
		host:          *fHost,
		port:          *fPort,
		serviceName:   *fServiceName,
		clusterType:   *fClusterType,
		oracleVersion: *fOracleVersion,
	}

	if m.interval != "" {
		du, err := time.ParseDuration(m.interval)
		if err != nil {
			l.Errorf("bad interval %s: %s, use default: 10m", m.interval, err.Error())
			m.intervalDuration = 10 * time.Minute
		} else {
			m.intervalDuration = du
		}
	}

	for {
		db, err := sql.Open("godror", fmt.Sprintf("%s/%s@%s:%s/%s", m.user, m.password, m.host, m.port, m.serviceName))
		if err == nil {
			m.db = db
			break
		}

		l.Errorf("oracle connect faild %v, retry each 3 seconds...", err)
		time.Sleep(time.Second * 3)
		continue
	}

	return m
}

func main() {
	flag.Parse()

	datakitPostURL = fmt.Sprintf("http://0.0.0.0:%d/v1/write/metric?name=oraclemonitor", *fDatakitHTTPPort)

	logger.SetGlobalRootLogger(*fLog, *fLogLevel, logger.OPT_DEFAULT)
	if *fInstanceDesc != "" { // add description to logger
		l = logger.SLogger("oraclemonitor-" + *fInstanceDesc)
	} else {
		l = logger.SLogger("oraclemonitor")
	}

	m := buildMonitor()
	m.run()
}

func (m *monitor) run() {

	l.Info("start oraclemonitor...")

	tick := time.NewTicker(m.intervalDuration)
	defer tick.Stop()
	defer m.db.Close()

	wg := sync.WaitGroup{}

	for {
		select {
		case <-tick.C:
			for _, ec := range execCfgs {
				wg.Add(1)
				go func() {
					defer wg.Done()
					m.handle(ec)
				}()
			}

			wg.Wait() // blocking
		}
	}
}

func (m *monitor) handle(ec *ExecCfg) {
	res, err := m.query(ec)
	if err != nil {
		l.Errorf("oracle query `%s' faild: %v, ignored", ec.metricType, err)
		return
	}

	if res == nil {
		return
	}

	_ = handleResponse(m, ec.metricType, ec.tagsMap, res)
}

func handleResponse(m *monitor, k string, tagsKeys []string, response []map[string]interface{}) error {
	lines := [][]byte{}

	for _, item := range response {

		tags := map[string]string{}

		tags["oracle_server"] = m.serviceName
		tags["oracle_port"] = m.port
		tags["instance_id"] = m.instanceId
		tags["instance_desc"] = m.desc
		tags["product"] = "oracle"
		tags["host"] = m.host
		tags["type"] = k

		for _, tagKey := range tagsKeys {
			tags[tagKey] = String(item[tagKey])
			delete(item, tagKey)
		}

		// add user-added tags
		// XXX: this may overwrite tags within @tags
		for k, v := range m.tags {
			tags[k] = v
		}

		ptline, err := io.MakeMetric(m.metric, tags, item, time.Now())
		if err != nil {
			l.Errorf("new point failed: %s", err.Error())
			return err
		}

		lines = append(lines, ptline)

	}

	if len(lines) == 0 {
		l.Debugf("no metric collected on %s", k)
		return nil
	}

	// io 输出
	if err := WriteData(bytes.Join(lines, []byte("\n")), datakitPostURL); err != nil {
		return err
	}

	return nil
}

func (m *monitor) query(obj *ExecCfg) ([]map[string]interface{}, error) {
	// 版本执行控制
	if !strings.Contains(obj.cluster, m.clusterType) {
		l.Debugf("ignore %s on cluster type %s", obj.metricType, m.clusterType)
		return nil, nil
	}

	if !strings.Contains(obj.version, m.oracleVersion) {
		l.Debugf("ignore %s on oracle version %s", obj.metricType, m.oracleVersion)
		return nil, nil
	}

	rows, err := m.db.Query(obj.sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, _ := rows.Columns()
	columnLength := len(columns)
	cache := make([]interface{}, columnLength)
	for idx, _ := range cache {
		var a interface{}
		cache[idx] = &a
	}
	var list []map[string]interface{}
	for rows.Next() {
		_ = rows.Scan(cache...)

		item := make(map[string]interface{})
		for i, data := range cache {
			key := strings.ToLower(columns[i])
			val := *data.(*interface{})

			if val != nil {
				vType := reflect.TypeOf(val)

				switch vType.String() {
				case "int64":
					item[key] = val.(int64)
				case "string":
					var data interface{}
					str := strings.TrimSpace(val.(string))
					data, err := strconv.ParseFloat(str, 64)
					if err != nil {
						data = val
					}
					item[key] = data
				case "time.Time":
					item[key] = val.(time.Time)
				case "[]uint8":
					item[key] = string(val.([]uint8))
				default:
					return nil, fmt.Errorf("unsupport data type '%s' now\n", vType)
				}
			}
		}

		list = append(list, item)
	}
	return list, nil
}

// String converts <i> to string.
func String(i interface{}) string {
	if i == nil {
		return ""
	}
	switch value := i.(type) {
	case int:
		return strconv.FormatInt(int64(value), 10)
	case int8:
		return strconv.Itoa(int(value))
	case int16:
		return strconv.Itoa(int(value))
	case int32:
		return strconv.Itoa(int(value))
	case int64:
		return strconv.FormatInt(int64(value), 10)
	case uint:
		return strconv.FormatUint(uint64(value), 10)
	case uint8:
		return strconv.FormatUint(uint64(value), 10)
	case uint16:
		return strconv.FormatUint(uint64(value), 10)
	case uint32:
		return strconv.FormatUint(uint64(value), 10)
	case uint64:
		return strconv.FormatUint(uint64(value), 10)
	case float32:
		return strconv.FormatFloat(float64(value), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(value)
	case string:
		return value
	case []byte:
		return string(value)
	case []rune:
		return string(value)
	default:
		// Finally we use json.Marshal to convert.
		jsonContent, _ := json.Marshal(value)
		return string(jsonContent)
	}
}

func WriteData(data []byte, urlPath string) error {
	// dataway path
	ctx, _ := context.WithCancel(context.Background())
	httpReq, err := http.NewRequest("POST", urlPath, bytes.NewBuffer(data))

	if err != nil {
		l.Errorf("[error] %s", err.Error())
		return err
	}

	httpReq = httpReq.WithContext(ctx)
	tmctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	resp, err := ctxhttp.Do(tmctx, http.DefaultClient, httpReq)
	if err != nil {
		l.Errorf("[error] %s", err.Error())
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Error(err)
		return err
	}

	switch resp.StatusCode / 100 {
	case 2:
		l.Debugf("post to %s ok", urlPath)
		return nil
	default:
		l.Errorf("post to %s failed(HTTP: %d): %s", urlPath, resp.StatusCode, string(body))
		return fmt.Errorf("post datakit failed")
	}
	return nil
}

const (
	oracle_hostinfo_sql = `
SELECT stat_name, value
FROM v$osstat
WHERE stat_name IN ('PHYSICAL_MEMORY_BYTES', 'NUM_CPUS')
`

	oracle_dbinfo_sql = `
SELECT dbid AS ora_db_id, name AS db_name, db_unique_name
	, to_char(created, 'yyyy-mm-dd hh24:mi:ss') AS db_create_time, log_mode AS log_mod
	, flashback_on AS flashback_mod, database_role, platform_name AS platform, open_mode, protection_mode
	, protection_level, switchover_status
FROM v$database
`

	oracle_instinfo_sql = `
SELECT instance_number, instance_name AS ora_sid, host_name, version
	, to_char(startup_time, 'yyyy-mm-dd hh24:mi:ss') AS startup_time, status
	, CASE
		WHEN parallel = 'YES' THEN 1
		ELSE 0
	END AS is_rac
FROM v$instance
`

	oracle_psu_sql = `
select  nvl(max(id),0) max_id from  Dba_Registry_History
`
	oracle_key_params_sql = `SELECT
    name,
    value
FROM
    v$parameter
WHERE
    name IN (
        'audit_trail',
        'sessions',
        'processes'
    )
`

	oracle_blocking_sessions_sql = `
WITH sessions AS (
    SELECT
        sid,
        serial#         serial,
        to_char(logon_time, 'yyyy-mm-dd hh24:mi:ss') logon_time,
        status,
        event,
        p1,
        p2,
        p3,
        username,
        terminal,
        program,
        sql_id,
        prev_sql_id,
        last_call_et,
        blocking_session,
        blocking_instance,
        row_wait_obj#   row_wait_obj
    FROM
        v$session
)
SELECT
    *
FROM
    sessions
WHERE
    sid IN (
        SELECT
            blocking_session
        FROM
            sessions
    )
    OR blocking_session IS NOT NULL
`

	oracle_undo_stat_sql = `
SELECT
    to_char(begin_time, 'yyyy-mm-dd hh24:mi:ss') begin_time,
    to_char(end_time, 'yyyy-mm-dd hh24:mi:ss') end_time,
    undoblks,
    txncount,
    activeblks,
    unexpiredblks,
    expiredblks
FROM
    v$undostat
WHERE
    ROWNUM < 2
`

	oracle_redo_info_sql = `
SELECT
    group#      group_no,
    thread#     thread_no,
    sequence#   sequence_no,
    bytes  AS size_bytes,
    members,
    archived,
    status
FROM
    v$log
`

	oracle_standby_log_sql = `
SELECT
    *
FROM
    (
        SELECT
            severity,
            message_num,
            error_code,
            to_char(timestamp, 'yyyy-mm-dd hh24:mi:ss') AS log_time,
            message
        FROM
            v$dataguard_status
        ORDER BY
            message_num DESC
    )
WHERE
    ROWNUM < 10
`

	oracle_standby_process_sql = `
SELECT
    process,
    ROW_NUMBER() OVER(
        PARTITION BY process
        ORDER BY
            process
    ) process_seq,
    status,
    client_process,
    client_dbid,
    group#      group_no,
    thread#     thread_no,
    sequence#   sequence_no,
    blocks,
    delay_mins
FROM
    v$managed_standby
`

	oracle_asm_diskgroups_sql = `
SELECT
    group_number,
    name       AS group_name,
    sector_size,
    block_size,
    allocation_unit_size,
    state,
    type,
    total_mb   AS space_total,
    free_mb    AS space_free,
    total_mb - free_mb AS space_used,
    required_mirror_free_mb,
    usable_file_mb,
    offline_disks,
    compatibility,
    database_compatibility
FROM
    v$asm_diskgroup
`

	oracle_flash_area_info_sql = `
SELECT
    substr(name, 1, 64) AS name,
    space_limit,
    space_used,
    space_reclaimable,
    number_of_files
FROM
    v$recovery_file_dest
`

	oracle_tbs_space_sql = `
SELECT
    tablespace_name,
    SUM(bytes)  AS space_total,
    SUM(
        CASE
            WHEN autoextensible = 'YES' THEN
                maxbytes - bytes
            ELSE
                0
        END
    )  AS space_extensible,
    COUNT(*) AS num_files
FROM
    dba_data_files
WHERE
    status = 'AVAILABLE'
GROUP BY
    tablespace_name
UNION ALL
SELECT
    tablespace_name,
    SUM(bytes) AS space_total,
    SUM(
        CASE
            WHEN autoextensible = 'YES' THEN
                maxbytes - bytes
            ELSE
                0
        END
    ) AS space_extensible,
    COUNT(*) AS num_files
FROM
    dba_temp_files
WHERE
    status = 'ONLINE'
GROUP BY
    tablespace_name
`

	oracle_tbs_meta_info_sql = `
SELECT
    tablespace_name,
    block_size,
    initial_extent,
    next_extent,
    min_extents,
    max_extents,
    pct_increase,
    min_extlen,
    status,
    contents,
    logging,
    force_logging,
    extent_management,
    allocation_type,
    plugged_in,
    segment_space_management,
    def_tab_compression,
    retention,
    bigfile
FROM
    dba_tablespaces
`
	oracle_temp_segment_usage_sql = `
SELECT
    tablespace_name,
    SUM(total_blocks) sum_total_blocks,
    SUM(max_blocks) sum_max_blocks,
    SUM(used_blocks) sum_used_blocks,
    SUM(free_blocks) sum_free_blocks
FROM
    v$sort_segment
GROUP BY
    tablespace_name
`

	oracle_trans_sql = `
select
            count(*) num_trans,
            nvl(round(max(used_ublk * 8192 / 1024 / 1024), 2),0) max_undo_size,
            nvl(round(avg(used_ublk * 8192 / 1024 / 1024), 2),0) avg_undo_size,
            round(nvl((sysdate - min(to_date(start_time, 'mm/dd/yy hh24:mi:ss'))),0) * 1440 * 60,0) longest_trans
        FROM v$transaction
`

	oracle_archived_log_sql = `
select count(*) value from v$archived_log where archived='YES' and deleted='NO'
`

	oracle_pgastat_sql = `
select name,value,unit from v$pgastat
`

	oracle_accounts_sql = `
select
username
,user_id
,password
,account_status
,to_char(lock_date, 'yyyy-mm-dd hh24:mi:ss') AS lock_date
,to_char(expiry_date, 'yyyy-mm-dd hh24:mi:ss') AS expiry_date
,default_tablespace
,temporary_tablespace
,to_char(created, 'yyyy-mm-dd hh24:mi:ss') AS created
,profile
,initial_rsrc_consumer_group
,external_name
,password_versions
,editions_enabled
,authentication_type
from dba_users
`

	oracle_locks_sql = `
SELECT b.session_id AS session_id,
       NVL(b.oracle_username, '(oracle)') AS oracle_username,
       a.owner AS object_owner,
       a.object_name,
       Decode(b.locked_mode, 0, 'None',
                             1, 'Null (NULL)',
                             2, 'Row-S (SS)',
                             3, 'Row-X (SX)',
                             4, 'Share (S)',
                             5, 'S/Row-X (SSX)',
                             6, 'Exclusive (X)',
                             b.locked_mode) locked_mode,
       b.os_user_name
FROM   dba_objects a,
       v$locked_object b
WHERE  a.object_id = b.object_id
ORDER BY 1, 2, 3, 4
`

	oracle_session_ratio_sql = `
SELECT 'session_cached_cursors' parameter,
         LPAD(VALUE, 5) value,
         DECODE(VALUE, 0, ' n/a', TO_CHAR(100 * USED / VALUE, '990') ) usage
   FROM (SELECT MAX(S.VALUE) USED
            FROM V$STATNAME N, V$SESSTAT S
           WHERE N.NAME = 'session cursor cache count'
             AND S.STATISTIC# = N.STATISTIC#),
         (SELECT VALUE FROM V$PARAMETER WHERE NAME = 'session_cached_cursors')
  UNION ALL
SELECT 'open_cursors' parameter,
         LPAD(VALUE, 5) value,
         TO_CHAR(100 * USED / VALUE, '990')   usage
   FROM (SELECT MAX(SUM(S.VALUE)) USED
            FROM V$STATNAME N, V$SESSTAT S
           WHERE N.NAME IN
                 ('opened cursors current', 'session cursor cache count')
             AND S.STATISTIC# = N.STATISTIC#
           GROUP BY S.SID),
         (SELECT VALUE FROM V$PARAMETER WHERE NAME = 'open_cursors')
`

	oracle_snap_info_sql = `
SELECT dbid,
to_char(sys_extract_utc(s.startup_time), 'yyyy-mm-dd hh24:mi:ss') snap_startup_time,
to_char(sys_extract_utc(s.begin_interval_time),
       'yyyy-mm-dd hh24:mi:ss') begin_interval_time,
to_char(sys_extract_utc(s.end_interval_time), 'yyyy-mm-dd hh24:mi:ss') end_interval_time,
s.snap_id, s.instance_number,
(cast(s.end_interval_time as date) - cast(s.begin_interval_time as date))*86400 as snap_in_second
from dba_hist_snapshot  s, v$instance b
where s.end_interval_time >= sysdate - interval '2' hour
and s.INSTANCE_NUMBER = b.INSTANCE_NUMBER
`

	oralce_backup_set_info_sql = `
select backup_types,count(backup_recid) backup_recid,
to_char(max(backup_start_time),'yyyy-mm-dd hh24:mi:ss') max_backup_start_time,
to_char(min(backup_start_time),'yyyy-mm-dd hh24:mi:ss') min_backup_start_time
from (select a.recid backup_recid,
        decode (b.incremental_level,
                '', decode (backup_type, 'L', 'ARCHIVELOG', 'FULL'),
                1, 'INCR-LV1',
                0, 'INCR-LV0',
                b.incremental_level)
           backup_types,
        decode (a.status,
                'A', 'AVAILABLE',
                'D', 'DELETED',
                'X', 'EXPIRED',
                'ERROR')
           backup_status,
        a.start_time backup_start_time,
        a.completion_time backup_completion_time,
        a.elapsed_seconds backup_est_seconds,
        a.bytes backup_size_bytes,
        a.compressed backup_compressed
   from gv$backup_piece a, gv$backup_set b
  where a.set_stamp = b.set_stamp and a.deleted = 'NO' AND A.STATUS='A'
) group by backup_types
`

	oracle_failed_login_sql = `
	select b.name name,b.lcount lcount_value from dba_users a join user$ b on a.username = b.name where a.INHERITED='NO'
	`

	oracle_temp_file_info_sql = `
	select f.tablespace_name "tablespace_name",
       d.file_name "tempfile_name",
    d.status "status",
       round((f.bytes_free + f.bytes_used), 2) "t_size",
       round(nvl(p.bytes_used, 0), 2) "t_use",
       round(((f.bytes_free + f.bytes_used) - nvl(p.bytes_used, 0)),
             2) "t_free",
       round((round(nvl(p.bytes_used, 0), 2) /
             round((f.bytes_free + f.bytes_used), 2)) * 100,
             2) as "t_percent"
  	from SYS.V_$TEMP_SPACE_HEADER f,
       DBA_TEMP_FILES           d,
       SYS.V_$TEMP_EXTENT_POOL  p
 	where f.tablespace_name(+) = d.tablespace_name
	`

	oracle_object_info_sql = `
	select object_type,
 	count(*) all_count,
 	sum(case when status = 'INVALID' then 1 else 0 end) invaid_count,
 	sum(case when status != 'INVALID' then 1 else 0 end) vaid_count
 	from dba_objects
	group by object_type
	`

	oracle_broken_jobs_sql = `select count(broken) broken_value from dba_jobs`

	oracle_session_wait_info_sql = `
	select wait_class, sum(total_waits) waits_value from  v$session_wait_class group by wait_class
	`

	oracle_scn_growth_statistics_sql = `
	declare
    rsl number;
    headroom_in_scn number;
    headroom_in_sec number;
    cur_scn_compat number;
    max_scn_compat number;

	begin
	    dbms_scn.getcurrentscnparams(rsl,headroom_in_scn,headroom_in_sec,cur_scn_compat,max_scn_compat);
	    dbms_output.put_line('RSL='||rsl);
	    dbms_output.put_line('headroom_in_scn='||headroom_in_scn);
	    dbms_output.put_line('headroom_in_sec='||headroom_in_sec);
	    dbms_output.put_line('CUR_SCN_COMPAT='||cur_scn_compat);
	    dbms_output.put_line('MAX_SCN_COMPAT='||max_scn_compat);
	end
	`

	oracle_scn_max_statistics_sql = `
	declare
	    rsl number;
	    headroom_in_scn number;
	    headroom_in_sec number;
	    cur_scn_compat number;
	    max_scn_compat number;

	begin
	    dbms_scn.getcurrentscnparams(rsl,headroom_in_scn,headroom_in_sec,cur_scn_compat,max_scn_compat);
	    dbms_output.put_line('RSL='||rsl);
	    dbms_output.put_line('headroom_in_scn='||headroom_in_scn);
	    dbms_output.put_line('headroom_in_sec='||headroom_in_sec);
	    dbms_output.put_line('CUR_SCN_COMPAT='||cur_scn_compat);
	    dbms_output.put_line('MAX_SCN_COMPAT='||max_scn_compat);
	end`

	oracle_pdb_failures_jobs_sql = `select failures  failures_value from dba_jobs`

	oracle_scn_instance_statistics_sql = `
	declare
    rsl number;
    headroom_in_scn number;
    headroom_in_sec number;
    cur_scn_compat number;
    max_scn_compat number;

	begin
	    dbms_scn.getcurrentscnparams(rsl,headroom_in_scn,headroom_in_sec,cur_scn_compat,max_scn_compat);
	    dbms_output.put_line('RSL='||rsl);
	    dbms_output.put_line('headroom_in_scn='||headroom_in_scn);
	    dbms_output.put_line('headroom_in_sec='||headroom_in_sec);
	    dbms_output.put_line('CUR_SCN_COMPAT='||cur_scn_compat);
	    dbms_output.put_line('MAX_SCN_COMPAT='||max_scn_compat);
	end
	`

	oracle_tablespace_free_pct_sql = `
select a.tablespace_name tablespace_name,
      a.bytes  t_size,
      (a.bytes - b.bytes) t_use,
      b.bytes t_free,
      round(((a.bytes - b.bytes) / a.bytes) * 100, 2) t_percent
 from (select tablespace_name, sum(bytes) bytes
         from dba_data_files
        group by tablespace_name) a,
      (select tablespace_name, sum(bytes) bytes, max(bytes) largest
         from dba_free_space
        group by tablespace_name) b
where a.tablespace_name = b.tablespace_name`

	// Data Guard 快速启动故障转移观察程序
	oracle_dg_fsfo_info_sql = `
    select fs_failover_status from v$database
    `

	// Data Guard 性能
	oracle_dg_performance_info_sql = `
    select  round(avg(value),2) redo_generation_rate from gv$sysmetric where  metric_name ='Redo Generated Per Sec'
    `

	// Data Guard 故障转移
	oracle_dg_failover_sql = `
    select  dest_id,end_of_redo_type,count(*) switch_counts from v$ARCHIVED_LOG where end_of_redo_type is not null group by dest_id,end_of_redo_type
    `

	// Data Guard 延迟信息
	oracle_dg_delay_info_sql = `
    select name dl_name,value dl_value,time_computed dl_flash_time from   V$DATAGUARD_STATS
    `

	// Data Guard 重做应用速率
	oracle_dg_apply_rate_sql = `
    select sofar redo_apply_rate from V$RECOVERY_PROGRESS where item='Active Apply Rate'
    `

	// Data Guard 目录错误信息
	oracle_dg_dest_error_sql = `
    select DEST_NAME,STATUS,NAME_SPACE,TARGET,ARCHIVER,ERROR,APPLIED_SCN from v$archive_dest where target='STANDBY'
    `

	// Data Guard 进程信息
	oracle_dg_proc_info_sql = `
    select process, status from v$managed_standby
    `

	// 采集容器数据库基础信息
	oracle_cdb_db_info_sql = `
    select name as pdb_name,open_mode from v$pdbs
    `

	// 采集容器资源基础信息
	oracle_cdb_resource_info_sql = `
    select r.con_id, p.pdb_name, r.cpu_utilization_limit, r.avg_cpu_utilization,r.cpu_wait_time, r.num_cpus,r.running_sessions_limit, r.avg_running_sessions, r.avg_waiting_sessions,r.avg_active_parallel_stmts, r.avg_queued_parallel_stmts,
    r.avg_active_parallel_servers, r.avg_queued_parallel_servers, r.parallel_servers_limit,r.iops,r.iombps,r.sga_bytes, r.pga_bytes, r.buffer_cache_bytes, r.shared_pool_bytes
    from v$rsrcpdbmetric r, cdb_pdbs p
    `

	oracle_clu_info_sql = `
	resource\ora.asm\crsctl status res ora.asm
    resource\ora.cvu\crsctl status res ora.cvu
    resource\ora.gsd\crsctl status res ora.gsd
    resource\ora.LISTENER.lsnr\crsctl status res ora.LISTENER.lsnr
    resource\ora.LISTENER_SCAN1.lsnr\crsctl status res ora.LISTENER_SCAN1.lsnr
    resource\ora.LISTENER_SCAN2.lsnr\crsctl status res ora.LISTENER_SCAN2.lsnr
    resource\ora.LISTENER_SCAN3.lsnr\crsctl status res ora.LISTENER_SCAN3.lsnr
    resource\ora.net1.network\crsctl status res ora.net1.network
    resource\ora.node1.vip\crsctl status res ora.node1.vip
    resource\ora.node2.vip\crsctl status res ora.node2.vip
    resource\ora.oc4j\crsctl status res ora.oc4j
    resource\ora.ons\crsctl status res ora.ons
    resource\ora.OVDATA.dg\crsctl status res ora.OVDATA.dg
    resource\ora.rac.db\crsctl status res ora.rac.db
    resource\ora.RACDATA.dg\crsctl status res ora.RACDATA.dg
    resource\ora.RACFRA.dg\crsctl status res ora.RACFRA.dg
    resource\ora.registry.acfs\crsctl status res ora.registry.acfs
    resource\ora.scan1.vip\crsctl status res ora.scan1.vip
    resource\ora.scan2.vip\crsctl status res ora.scan2.vip
    resource\ora.scan3.vip\crsctl status res ora.scan3.vip
    server\Oracle High Availability Services\crsctl check has
    server\Cluster Ready Services\crsctl check crs | grep Ready
    server\Cluster Synchronization Services\crsctl check css
    server\Event Manager\crsctl check evm
    component\freespace\cluvfy comp freespace | tail -1
    component\clocksync\cluvfy comp clocksync  | tail -1
    component\nodeapp\cluvfy comp nodeapp | tail -1
    component\scan\cluvfy comp scan | tail -1
	`

	// 采集ASM磁盘组状态
	oracle_asm_group_info_sql = `
    select GROUP_NUMBER,NAME AS GROUP_NAME,STATE,TYPE,TOTAL_MB,FREE_MB,REQUIRED_MIRROR_FREE_MB,USABLE_FILE_MB,OFFLINE_DISKS,VOTING_FILES from v$asm_diskgroup
    `

	// 采集ASM磁盘组状态
	oracle_asm_disk_info_sql = `
    select
    group_number,group_name,disk_number,name as disk_name,mount_status,
    mode_status,state,os_mb,total_mb,free_mb,path,create_date,mount_date,
    repair_timer,reads,writes,read_errs,write_errs,read_time,write_time,
    bytes_read,bytes_written,voting_file from v$asm_disk d,
    (select group_number as g_number,name as group_name from v$asm_diskgroup) g
    where d.group_number=g.g_number
    `
)

type ExecCfg struct {
	sql        string
	metricType string
	version    string
	cluster    string
	tagsMap    []string
}

var execCfgs = []*ExecCfg{
	&ExecCfg{
		sql:        oracle_hostinfo_sql,
		metricType: "oracle_hostinfo",
		cluster:    "single",
		version:    "10g, 11g, 12c",
		tagsMap:    []string{"stat_name"},
	},

	&ExecCfg{
		sql:        oracle_dbinfo_sql,
		metricType: "oracle_dbinfo",
		cluster:    "single",
		version:    "10g, 11g, 12c",
		tagsMap:    []string{"ora_db_id"},
	},

	&ExecCfg{
		sql:        oracle_key_params_sql,
		metricType: "oracle_key_params",
		cluster:    "single",
		version:    "10g, 11g, 12c",
		tagsMap:    []string{"ora_db_id"},
	},

	&ExecCfg{
		sql:        oracle_instinfo_sql,
		metricType: "oracle_instinfo",
		cluster:    "single",
		version:    "10g, 11g, 12c",
	},

	&ExecCfg{
		sql:        oracle_psu_sql,
		metricType: "oracle_psu",
		cluster:    "single",
		version:    "10g, 11g, 12c",
	},

	&ExecCfg{
		sql:        oracle_blocking_sessions_sql,
		metricType: "oracle_blocking_sessions",
		cluster:    "single",
		version:    "10g, 11g, 12c",
		tagsMap:    []string{"sid", "serial", "username"},
	},

	&ExecCfg{
		sql:        oracle_undo_stat_sql,
		metricType: "oracle_undo_stat",
		cluster:    "single",
		version:    "10g, 11g, 12c",
		tagsMap:    []string{"stat_name"},
	},

	&ExecCfg{
		sql:        oracle_redo_info_sql,
		metricType: "oracle_redo_info",
		cluster:    "single",
		version:    "10g, 11g, 12c",
		tagsMap:    []string{"group_no", "sequence_no"},
	},

	&ExecCfg{
		sql:        oracle_standby_log_sql,
		metricType: "oracle_standby_log",
		cluster:    "single",
		version:    "10g, 11g, 12c",
		tagsMap:    []string{"message_num"},
	},

	&ExecCfg{
		sql:        oracle_standby_process_sql,
		metricType: "oracle_standby_process",
		cluster:    "single",
		version:    "10g, 11g, 12c",
		tagsMap:    []string{"process_seq"},
	},

	&ExecCfg{
		sql:        oracle_asm_diskgroups_sql,
		metricType: "oracle_asm_diskgroups",
		cluster:    "single",
		version:    "10g, 11g, 12c",
		tagsMap:    []string{"group_number", "group_name"},
	},

	&ExecCfg{
		sql:        oracle_flash_area_info_sql,
		metricType: "oracle_flash_area_info",
		cluster:    "single",
		version:    "10g, 11g, 12c",
		tagsMap:    []string{"name"},
	},

	&ExecCfg{
		sql:        oracle_tbs_space_sql,
		metricType: "oracle_tbs_space",
		cluster:    "single",
		version:    "10g, 11g, 12c",
		tagsMap:    []string{"tablespace_name"},
	},

	&ExecCfg{
		sql:        oracle_failed_login_sql,
		metricType: "oracle_failed_login",
		cluster:    "single",
		version:    "10g, 11g, 12c",
	},

	&ExecCfg{
		sql:        oracle_temp_file_info_sql,
		metricType: "oracle_temp_file_info",
		cluster:    "single",
		version:    "10g, 11g, 12c",
	},

	&ExecCfg{
		sql:        oracle_object_info_sql,
		metricType: "oracle_object_info",
		cluster:    "single",
		version:    "10g, 11g, 12c",
	},

	&ExecCfg{
		sql:        oracle_broken_jobs_sql,
		metricType: "oracle_broken_jobs",
		cluster:    "single",
		version:    "10g, 11g, 12c",
	},

	&ExecCfg{
		sql:        oracle_session_wait_info_sql,
		metricType: "oracle_session_wait_info",
		cluster:    "single",
		version:    "10g, 11g, 12c",
	},

	&ExecCfg{
		sql:        oracle_scn_growth_statistics_sql,
		metricType: "oracle_scn_growth_statistics",
		cluster:    "single",
		version:    "10g, 11g, 12c",
	},

	&ExecCfg{
		sql:        oracle_scn_max_statistics_sql,
		metricType: "oracle_scn_max_statistics",
		cluster:    "single",
		version:    "10g, 11g, 12c",
	},

	&ExecCfg{
		sql:        oracle_pdb_failures_jobs_sql,
		metricType: "oracle_pdb_failures_jobs",
		cluster:    "single",
		version:    "10g, 11g, 12c",
	},

	&ExecCfg{
		sql:        oracle_tbs_meta_info_sql,
		metricType: "oracle_tbs_meta_info",
		cluster:    "single",
		version:    "10g, 11g, 12c",
		tagsMap:    []string{"tablespace_name"},
	},

	&ExecCfg{
		sql:        oracle_temp_segment_usage_sql,
		metricType: "oracle_temp_segment_usage",
		cluster:    "single",
		version:    "10g, 11g, 12c",
		tagsMap:    []string{"tablespace_name"},
	},

	&ExecCfg{
		sql:        oracle_trans_sql,
		metricType: "oracle_trans",
		cluster:    "single",
		version:    "10g, 11g, 12c",
		tagsMap:    []string{"stat_name"},
	},

	&ExecCfg{
		sql:        oracle_archived_log_sql,
		metricType: "oracle_archived_log",
		cluster:    "single",
		version:    "10g, 11g, 12c",
		tagsMap:    []string{"stat_name"},
	},

	&ExecCfg{
		sql:        oracle_pgastat_sql,
		metricType: "oracle_pgastat",
		cluster:    "single",
		version:    "10g, 11g, 12c",
		tagsMap:    []string{"name"},
	},

	&ExecCfg{
		sql:        oracle_accounts_sql,
		metricType: "oracle_accounts",
		cluster:    "single",
		version:    "10g, 11g, 12c",
		tagsMap:    []string{"username", "user_id"},
	},

	&ExecCfg{
		sql:        oracle_locks_sql,
		metricType: "oracle_locks",
		cluster:    "single",
		version:    "10g, 11g, 12c",
		tagsMap:    []string{"session_id"},
	},

	&ExecCfg{
		sql:        oracle_session_ratio_sql,
		metricType: "oracle_session_ratio",
		cluster:    "single",
		version:    "10g, 11g, 12c",
		tagsMap:    []string{"parameter"},
	},

	&ExecCfg{
		sql:        oracle_snap_info_sql,
		metricType: "oracle_snap_info",
		cluster:    "single",
		version:    "10g, 11g, 12c",
		tagsMap:    []string{"dbid", "snap_id"},
	},

	&ExecCfg{
		sql:        oralce_backup_set_info_sql,
		metricType: "oralce_backup_set_info",
		cluster:    "single",
		version:    "10g, 11g, 12c",
		tagsMap:    []string{"backup_types"},
	},

	&ExecCfg{
		sql:        oracle_tablespace_free_pct_sql,
		metricType: "oracle_tablespace_free_pct",
		cluster:    "single",
		version:    "10g, 11g, 12c",
		tagsMap:    []string{"tablesapce_name"},
	},

	&ExecCfg{
		sql:        oracle_dg_fsfo_info_sql,
		metricType: "oracle_dg_fsfo_info",
		cluster:    "dg",
		version:    "10g, 11g, 12c",
	},

	&ExecCfg{
		sql:        oracle_dg_performance_info_sql,
		metricType: "oracle_dg_performance_info",
		cluster:    "dg",
		version:    "10g, 11g, 12c",
	},

	&ExecCfg{
		sql:        oracle_dg_failover_sql,
		metricType: "oracle_dg_failover",
		cluster:    "dg",
		version:    "10g, 11g, 12c",
	},

	&ExecCfg{
		sql:        oracle_dg_delay_info_sql,
		metricType: "oracle_dg_delay_info",
		cluster:    "dg",
		version:    "10g, 11g, 12c",
	},

	&ExecCfg{
		sql:        oracle_dg_apply_rate_sql,
		metricType: "oracle_dg_apply_rate",
		cluster:    "dg",
		version:    "10g, 11g, 12c",
	},

	&ExecCfg{
		sql:        oracle_dg_dest_error_sql,
		metricType: "oracle_dg_dest_error",
		cluster:    "dg",
		version:    "10g, 11g, 12c",
		tagsMap:    []string{"dest_name"},
	},

	&ExecCfg{
		sql:        oracle_dg_proc_info_sql,
		metricType: "oracle_dg_proc_info",
		cluster:    "dg",
		version:    "10g, 11g, 12c",
		tagsMap:    []string{"process"},
	},

	&ExecCfg{
		sql:        oracle_cdb_db_info_sql,
		metricType: "oracle_cdb_db_info",
		cluster:    "single",
		version:    "12c",
	},

	&ExecCfg{
		sql:        oracle_cdb_resource_info_sql,
		metricType: "oracle_cdb_resource_info",
		cluster:    "single",
		version:    "12c",
		tagsMap:    []string{"pdb_name"},
	},

	&ExecCfg{
		sql:        oracle_clu_info_sql,
		metricType: "oracle_clu_info",
		cluster:    "rac",
		version:    "11g",
	},

	&ExecCfg{
		sql:        oracle_asm_group_info_sql,
		metricType: "oracle_asm_group_info",
		cluster:    "rac",
		version:    "11g",
	},

	&ExecCfg{
		sql:        oracle_asm_disk_info_sql,
		metricType: "oracle_asm_disk_info",
		cluster:    "rac",
		version:    "11g",
		tagsMap:    []string{"disk_name", "group_name"},
	},
}
