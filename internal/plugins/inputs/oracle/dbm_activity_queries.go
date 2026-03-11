// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oracle

const activityQueryDirect12 = `SELECT /*+ push_pred(sq) push_pred(sq_prev) */ /* DK_ACTIVITY_SAMPLING */
SYSDATE as now,
s.sid,
s.serial#,
s.username,
s.status,
s.osuser,
s.process,
s.machine,
s.port,
s.program,
s.type,
s.sql_id,
sq.force_matching_signature as force_matching_signature,
sq.plan_hash_value sql_plan_hash_value,
s.sql_exec_start,
s.sql_address,
s.prev_sql_id,
sq_prev.plan_hash_value prev_sql_plan_hash_value,
s.prev_exec_start as prev_sql_exec_start,
sq_prev.force_matching_signature as prev_force_matching_signature,
s.prev_sql_addr prev_sql_address,
s.module,
s.action,
s.client_info,
s.logon_time,
s.client_identifier,
CASE WHEN blocking_session_status = 'VALID' THEN
blocking_instance
ELSE
null
END blocking_instance,
CASE WHEN blocking_session_status = 'VALID' THEN
  blocking_session
ELSE
  null
END blocking_session,
CASE WHEN final_blocking_session_status = 'VALID' THEN
  final_blocking_instance
ELSE
  null
END final_blocking_instance,
CASE WHEN final_blocking_session_status = 'VALID' THEN
  final_blocking_session
ELSE
  null
END final_blocking_session,
CASE WHEN state = 'WAITING' THEN
  event
ELSE
  'CPU'
END event,
CASE WHEN state = 'WAITING' THEN
  wait_class
ELSE
  'CPU'
END wait_class,
s.wait_time_micro,
c.name as pdb_name,
dbms_lob.substr(sq.sql_fulltext, :sql_substr_length_1, 1) sql_fulltext,
dbms_lob.substr(sq_prev.sql_fulltext, :sql_substr_length_2, 1) prev_sql_fulltext,
comm.command_name
FROM
v$session s,
v$sql sq,
v$sql sq_prev,
v$containers c,
v$sqlcommand comm
WHERE
sq.sql_id(+)   = s.sql_id
AND sq.child_number(+) = s.sql_child_number
AND sq_prev.sql_id(+)   = s.prev_sql_id
AND sq_prev.child_number(+) = s.prev_child_number
AND ( sq.sql_text NOT LIKE '%DK_ACTIVITY_SAMPLING%' OR sq.sql_text is NULL )
AND s.con_id = c.con_id(+)
AND s.command = comm.command_type(+)`

const activityQueryDirect11 = `SELECT /*+ push_pred(sq) push_pred(sq_prev) */ /* DK_ACTIVITY_SAMPLING */
SYSDATE as now,
s.sid,
s.serial#,
s.username,
s.status,
s.osuser,
s.process,
s.machine,
s.port,
s.program,
s.type,
s.sql_id,
sq.force_matching_signature as force_matching_signature,
sq.plan_hash_value sql_plan_hash_value,
s.sql_exec_start,
s.sql_address,
s.prev_sql_id,
sq_prev.plan_hash_value prev_sql_plan_hash_value,
s.prev_exec_start as prev_sql_exec_start,
sq_prev.force_matching_signature as prev_force_matching_signature,
s.prev_sql_addr prev_sql_address,
s.module,
s.action,
s.client_info,
s.logon_time,
s.client_identifier,
CASE WHEN blocking_session_status = 'VALID' THEN
blocking_instance
ELSE
null
END blocking_instance,
CASE WHEN blocking_session_status = 'VALID' THEN
  blocking_session
ELSE
  null
END blocking_session,
CASE WHEN final_blocking_session_status = 'VALID' THEN
  final_blocking_instance
ELSE
  null
END final_blocking_instance,
CASE WHEN final_blocking_session_status = 'VALID' THEN
  final_blocking_session
ELSE
  null
END final_blocking_session,
CASE WHEN state = 'WAITING' THEN
  event
ELSE
  'CPU'
END event,
CASE WHEN state = 'WAITING' THEN
  wait_class
ELSE
  'CPU'
END wait_class,
s.wait_time_micro,
NULL as pdb_name,
dbms_lob.substr(sq.sql_fulltext, :sql_substr_length_1, 1) sql_fulltext,
dbms_lob.substr(sq_prev.sql_fulltext, :sql_substr_length_2, 1) prev_sql_fulltext,
comm.command_name
FROM
v$session s,
v$sql sq,
v$sql sq_prev,
v$sqlcommand comm
WHERE
sq.sql_id(+)   = s.sql_id
AND sq.child_number(+) = s.sql_child_number
AND sq_prev.sql_id(+)   = s.prev_sql_id
AND sq_prev.child_number(+) = s.prev_child_number
AND ( sq.sql_text NOT LIKE '%DK_ACTIVITY_SAMPLING%' OR sq.sql_text is NULL )
AND s.command = comm.command_type(+)`

// Connection queries for Oracle 12+.
const connectionQuery12 = `SELECT /* DK_CONNECTION */
    s.username AS user_name,
    s.status,
    c.name AS pdb_name,
    COUNT(*) AS connection_count
FROM
    v$session s,
    v$containers c
WHERE
    s.con_id = c.con_id(+)
    AND s.type = 'USER'
GROUP BY
    s.username,
    s.status,
    c.name`

// Connection queries for Oracle 11.
const connectionQuery11 = `SELECT /* DK_CONNECTION */
    s.username AS user_name,
    s.status,
    COUNT(*) AS connection_count
FROM
    v$session s
WHERE
    s.type = 'USER'
GROUP BY
    s.username,
    s.status`
