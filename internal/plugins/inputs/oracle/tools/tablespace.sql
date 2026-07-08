---------------------------------------------------
-- table space
---------------------------------------------------
SELECT
  c.name pdb_name,
  t.tablespace_name tablespace_name,
	t.block_size,
  CASE UPPER(t.status)
    WHEN 'ONLINE' THEN 1
    WHEN 'OFFLINE' THEN 2
    WHEN 'READ ONLY' THEN 3
    ELSE 0
  END tablespace_status,
  NVL(m.used_space * t.block_size, 0) used,
  NVL(m.tablespace_size * t.block_size, 0) size_,
  NVL(m.used_percent, 0) in_use,
  NVL(m.used_space, 0) offline_
FROM
  cdb_tablespace_usage_metrics m, cdb_tablespaces t, v$containers c
WHERE
  m.tablespace_name(+) = t.tablespace_name and c.con_id(+) = t.con_id;;;

SELECT
  m.tablespace_name,
  CASE UPPER(t.status)
    WHEN 'ONLINE' THEN 1
    WHEN 'OFFLINE' THEN 2
    WHEN 'READ ONLY' THEN 3
    ELSE 0
  END tablespace_status,
  NVL(m.used_space * t.block_size, 0) as used,
  m.tablespace_size * t.block_size as size_,
  NVL(m.used_percent, 0) as in_use,
  NVL(m.used_space, 0) as offline_
FROM
  dba_tablespace_usage_metrics m
  join dba_tablespaces t on m.tablespace_name = t.tablespace_name;;;
