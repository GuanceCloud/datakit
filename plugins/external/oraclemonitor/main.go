package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/godror/godror"
	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/influxdata/toml"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/oraclemonitor"
)

var (
	flagCfg    = flag.String("cfg", "", "toml config file path")
	flagGetCfg = flag.String("get-cfg", "", "get config sample, default write to ./instances.toml")

	flagRPCServer = flag.String("rpc-server", "unix://"+config.GRPCDomainSock, "gRPC server")
	flagLog       = flag.String("log", filepath.Join(config.InstallDir, "external", "oraclemonitor.log"), "log file")
	flagLogLevel  = flag.String("log-level", "info", "log file")

	l         *zap.SugaredLogger
	instances []*oraclemonitor.Instance
	rpcCli    io.DataKitClient
)

type impl struct {
	LibPath   string                    `toml:"libPath"` // no used
	Instances []*oraclemonitor.Instance `toml:"instances"`
}

func main() {
	flag.Parse()

	if *flagGetCfg != "" {
		// TODO
		panic("no implemented")
	}

	if *flagCfg == "" {
		panic("toml config missing")
	}

	logger.SetGlobalRootLogger(*flagLog, *flagLogLevel, logger.OPT_ENC_CONSOLE|logger.OPT_SHORT_CALLER)
	l = logger.SLogger("oraclemonitor")

	data, err := ioutil.ReadFile(*flagCfg)
	if err != nil {
		l.Fatal(err)
		return
	}

	tbl, err := toml.Parse(data)
	if err != nil {
		l.Fatal(err)
		return
	}

	var i impl
	if err := toml.UnmarshalTable(tbl, &i); err != nil {
		l.Error(err)
		return
	}

	l.Infof("gRPC dial %s...", *flagRPCServer)
	conn, err := grpc.Dial(*flagRPCServer, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(time.Second*5))
	if err != nil {
		l.Fatalf("connect RCP failed: %s", err)
	}

	l.Infof("gRPC connect %s ok", *flagRPCServer)
	defer conn.Close()

	rpcCli = io.NewDataKitClient(conn)

	i.Start()
}

func (i *impl) Start() {
	l.Infof("start monit %d oracle instances...", len(i.Instances))

	wg := sync.WaitGroup{}

	for _, inst := range i.Instances {
		connStr := fmt.Sprintf("%s/%s@%s/%s", inst.User, inst.Password, inst.Host, inst.Server)
		db, err := sql.Open("godror", connStr)
		if err != nil {
			l.Errorf("oracle connect faild %v", err)
			continue
		}

		du, err := time.ParseDuration(inst.Interval)
		if err != nil {
			l.Error(err)
			continue
		}

		inst.DB = db
		inst.IntervalDuration = du
		wg.Add(1)
		go func() {
			defer wg.Done()
			run(inst)
		}()
	}

	wg.Wait()
}

func run(i *oraclemonitor.Instance) {
	for {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		for key, stmt := range metricMap {

			l.Debugf("sql: %s", key)
			res, err := query(i.DB, stmt)
			if err != nil {
				l.Errorf("oracle connect faild %v", err)
				continue
			}

			handleResponse(i, key, res)
		}

		internal.SleepContext(ctx, i.IntervalDuration)
	}
}

<<<<<<< HEAD
func (i *instance) handleResponse(m string, response []map[string]interface{}) error {
=======
func handleResponse(i *oraclemonitor.Instance, m string, response []map[string]interface{}) error {

>>>>>>> 6f0257fec0a3186da151b9b25cb5dd724cf3fe38
	lines := []string{}

	for _, item := range response {
		tags := map[string]string{}

		tags["oracle_server"] = i.Server
		tags["oracle_port"] = i.Port
		tags["instance_id"] = i.InstanceId
		tags["instance_desc"] = i.Desc
		tags["product"] = "oracle"
		tags["host"] = i.Host
		tags["type"] = m

		if tagKeys, ok := tagsMap[m]; ok {
			for _, tagKey := range tagKeys {
				tags[tagKey] = String(item[tagKey])
				delete(item, tagKey)
			}
		}

		pt, err := influxdb.NewPoint(i.Metric, tags, item, time.Now())
		if err != nil {
			l.Errorf("new point failed: %s", err.Error())
			return err
		}
		lines = append(lines, pt.String())
	}

	if len(lines) == 0 {
		l.Debugf("no metric collected on %s", m)
		return nil
	}

	// TODO: RPC post to datakit

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := rpcCli.Send(ctx, &io.Request{
		Lines:     []byte(strings.Join(lines, "\n")),
		Precision: "ns",
		Name:      "oraclemonitor",
	})
	if err != nil {
		l.Errorf("feed error: %s", err.Error())
		return err
	}

	l.Debugf("feed %d points, error: %s", r.GetPoints(), r.GetErr())

	return nil
}

func query(db *sql.DB, sql string) ([]map[string]interface{}, error) {
	rows, err := db.Query(sql)
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

const (
  configSample = `
  {
  "log":       "/usr/local/cloudcare/DataFlux/datakit/oraclemonitor.log",
  "log_level": "info",
	"oracles" : [
		{
			"interval" : "10s",
			"metricName" : "",
			"instanceId" : "",
			"instanceDesc" : "",
			"host" : "",
			"port" : "",
			"username" : "",
			"password" : "",
			"server" : "",
			"type": "singleInstance"
		}
	]
}`

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
	oracle_key_params = `SELECT
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
    select fs_failover_status from v$database;
    ` 
    
	// Data Guard 性能
    oracle_dg_performance_info_sql = `
    select  round(avg(value),2) redo_generation_rate from gv$sysmetric where  metric_name ='Redo Generated Per Sec';
    `
    
	// Data Guard 故障转移
    oracle_dg_failover_sql = `
    select  dest_id,end_of_redo_type,count(*) switch_counts from v$ARCHIVED_LOG where end_of_redo_type is not null group by dest_id,end_of_redo_type;
    `
	
	// Data Guard 延迟信息
    oracle_dg_delay_info_sql = `
    select name dl_name,value dl_value,time_computed dl_flash_time from   V$DATAGUARD_STATS;
    `
	
	// Data Guard 重做应用速率
    oracle_dg_apply_rate_sql = `
    select sofar redo_apply_rate from V$RECOVERY_PROGRESS where item='Active Apply Rate';
    `
	
	// Data Guard 目录错误信息
    oracle_dg_dest_error_sql = `
    select DEST_NAME,STATUS,NAME_SPACE,TARGET,ARCHIVER,ERROR,APPLIED_SCN from v$archive_dest where target='STANDBY';
    `
	
	// Data Guard 进程信息
    oracle_dg_proc_info_sql = `
    select process, status from v$managed_standby;
    `
	
	// 采集容器数据库基础信息
	oracle_cdb_db_info_sql = `
    select name as pdb_name,open_mode from v$pdbs;
    `
    
	// 采集容器资源基础信息
	oracle_cdb_resource_info_sql = `
    select r.con_id, p.pdb_name, r.cpu_utilization_limit, r.avg_cpu_utilization,r.cpu_wait_time, r.num_cpus,r.running_sessions_limit, r.avg_running_sessions, r.avg_waiting_sessions,r.avg_active_parallel_stmts, r.avg_queued_parallel_stmts,
    r.avg_active_parallel_servers, r.avg_queued_parallel_servers, r.parallel_servers_limit,r.iops,r.iombps,r.sga_bytes, r.pga_bytes, r.buffer_cache_bytes, r.shared_pool_bytes
    from v$rsrcpdbmetric r, cdb_pdbs p
    `

	// 采集ASM磁盘组状态
	oracle_asm_group_info_sql = `
    select GROUP_NUMBER,NAME AS GROUP_NAME,STATE,TYPE,TOTAL_MB,FREE_MB,REQUIRED_MIRROR_FREE_MB,USABLE_FILE_MB,OFFLINE_DISKS,VOTING_FILES from v$asm_diskgroup;
    `

	// 采集ASM磁盘组状态
	oracle_asm_disk_info_sql = `
    select 
    group_number,group_name,disk_number,name as disk_name,mount_status,
    mode_status,state,os_mb,total_mb,free_mb,path,create_date,mount_date,
    repair_timer,reads,writes,read_errs,write_errs,read_time,write_time,
    bytes_read,bytes_written,voting_file from v$asm_disk d,
    (select group_number as g_number,name as group_name from v$asm_diskgroup) g
    where d.group_number=g.g_number;
    `
)
    
  

var (
	metricMap = map[string]string{
		"oracle_hostinfo":            oracle_hostinfo_sql,
		"oracle_dbinfo":              oracle_dbinfo_sql,
		"oracle_instinfo":            oracle_instinfo_sql,
		"oracle_psu":                 oracle_psu_sql,
		"oracle_blocking_sessions":   oracle_blocking_sessions_sql,
		"oracle_undo_stat":           oracle_undo_stat_sql,
		"oracle_redo_info":           oracle_redo_info_sql,
		"oracle_standby_log":         oracle_standby_log_sql,
		"oracle_standby_process":     oracle_standby_process_sql,
		"oracle_asm_diskgroups":      oracle_asm_diskgroups_sql,
		"oracle_flash_area_info":     oracle_flash_area_info_sql,
		"oracle_tbs_space":           oracle_tbs_space_sql,
		"oracle_tbs_meta_info":       oracle_tbs_meta_info_sql,
		"oracle_temp_segment_usage":  oracle_temp_segment_usage_sql,
		"oracle_trans":               oracle_trans_sql,
		"oracle_archived_log":        oracle_archived_log_sql,
		"oracle_pgastat":             oracle_pgastat_sql,
		"oracle_accounts":            oracle_accounts_sql,
		"oracle_locks":               oracle_locks_sql,
		"oracle_session_ratio":       oracle_session_ratio_sql,
		"oracle_snap_info":           oracle_snap_info_sql,
		"oralce_backup_set_info":     oralce_backup_set_info_sql,
		"oracle_tablespace_free_pct": oracle_tablespace_free_pct_sql,
		"oracle_dg_fsfo_info":        oracle_dg_fsfo_info_sql,
		"oracle_dg_performance_info": oracle_dg_performance_info_sql,
		"oracle_dg_failover":         oracle_dg_failover_sql,
		"oracle_dg_delay_info":       oracle_dg_delay_info_sql,
		"oracle_dg_apply_rate":       oracle_dg_apply_rate_sql,
		"oracle_dg_dest_error":       oracle_dg_dest_error_sql,
		"oracle_dg_proc_info":        oracle_dg_proc_info_sql,
		"oracle_cdb_db_info":         oracoracle_cdb_db_info_sql,
		"oracle_cdb_resource_info":   oracle_cdb_resource_info_sql,
		"oracle_asm_group_info":      oracle_asm_group_info_sql,
		"oracle_asm_disk_info":       oracle_asm_disk_info_sql;
	}

	tagsMap = map[string][]string{
		"oracle_hostinfo":            []string{"stat_name"},
		"oracle_dbinfo":              []string{"ora_db_id"},
		"oracle_key_params":          []string{"name"},
		"oracle_blocking_sessions":   []string{"sid", "serial", "username"},
		"oracle_redo_info":           []string{"group_no", "sequence_no"},
		"oracle_standby_log":         []string{"message_num"},
		"oracle_standby_process":     []string{"process_seq"},
		"oracle_asm_diskgroups":      []string{"group_number", "group_name"},
		"oracle_flash_area_info":     []string{"name"},
		"oracle_tbs_space":           []string{"tablespace_name"},
		"oracle_tbs_meta_info":       []string{"tablespace_name"},
		"oracle_temp_segment_usage":  []string{"tablespace_name"},
		"oracle_pgastat":             []string{"name"},
		"oracle_accounts":            []string{"username", "user_id"},
		"oracle_locks":               []string{"session_id"},
		"oracle_session_ratio":       []string{"parameter"},
		"oracle_snap_info":           []string{"dbid", "snap_id"},
		"oralce_backup_set_info":     []string{"backup_types"},
		"oracle_tablespace_free_pct": []string{"tablesapce_name"},
	}
)
