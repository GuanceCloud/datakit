package zabbix

func mySQL(tablename string) string {
	switch tablename {
	case "history":
		return mysqlHistory
	case "history_uint":
		return mysqlHistoryUInt
	case "trends":
		return mysqlTrends
	case "trends_uint":
		return mysqlTrendsUInt
	default:
		panic("unrecognized tablename")
	}
}

const mysqlTrends string = `SELECT 
-- measurement
replace(replace(CASE
    WHEN (position('$2' in ite.name) > 0) AND (position('$4' in ite.name) > 0) 
      THEN replace(replace(ite.name, '$2', SUBSTRING_INDEX(SUBSTRING_INDEX(SUBSTRING(ite.key_, LOCATE('[',ite.key_) + 1, LOCATE(']',ite.key_) - LOCATE('[',ite.key_)-1),',',2), ',',-1)), '$4', SUBSTRING_INDEX(SUBSTRING(ite.key_, LOCATE('[',ite.key_) + 1, LOCATE(']',ite.key_) - LOCATE('[',ite.key_)-1),',',-1))
	WHEN (position('$1' in ite.name) > 0) AND (position('$2' in ite.name) > 0) 
      THEN replace(replace(ite.name, '$1', SUBSTRING_INDEX(SUBSTRING(ite.key_, LOCATE('[',ite.key_) + 1, LOCATE(']',ite.key_) - LOCATE('[',ite.key_)-1),',',1)), '$2', SUBSTRING_INDEX(SUBSTRING_INDEX(SUBSTRING(ite.key_, LOCATE('[',ite.key_) + 1, LOCATE(']',ite.key_) - LOCATE('[',ite.key_)-1),',',2), ',',-1))
    WHEN (position('$1' in ite.name) > 0) 
       THEN replace(ite.name, '$1', SUBSTRING_INDEX(SUBSTRING_INDEX(SUBSTRING(ite.key_, LOCATE('[',ite.key_) + 1, LOCATE(']',ite.key_) - LOCATE('[',ite.key_)-1),',',1), ',',-1))
    WHEN (position('$2' in ite.name) > 0) 
       THEN replace(ite.name, '$2', SUBSTRING_INDEX(SUBSTRING_INDEX(SUBSTRING(ite.key_, LOCATE('[',ite.key_) + 1, LOCATE(']',ite.key_) - LOCATE('[',ite.key_)-1),',',2), ',',-1))
	WHEN (position('$3' in ite.name) > 0)
       THEN replace(ite.name, '$3', SUBSTRING_INDEX(SUBSTRING_INDEX(SUBSTRING(ite.key_, LOCATE('[',ite.key_) + 1, LOCATE(']',ite.key_) - LOCATE('[',ite.key_)-1),',',-2), ',',1))
    WHEN (position('$1' in ite.name) > 0) AND (position('$3' in ite.name) > 0)
       THEN replace(replace(ite.name, '$1', SUBSTRING_INDEX(SUBSTRING_INDEX(SUBSTRING(ite.key_, LOCATE('[',ite.key_) + 1, LOCATE(']',ite.key_) - LOCATE('[',ite.key_)-1),',',1), ',',-1)), '$3', SUBSTRING_INDEX(SUBSTRING_INDEX(SUBSTRING(ite.key_, LOCATE('[',ite.key_) + 1, LOCATE(']',ite.key_) - LOCATE('[',ite.key_)-1),',',-2), ',',1))
    ELSE ite.name
  END, ',', ''), ' ', '\\ ') 
-- tags
|| ',host_name=' || replace(hos.name, ' ', '\\ ')
|| ',group_name=' || replace(grp.name, ' ', '\\ ')
|| ',applications=' || ifnull(replace(replace((SELECT GROUP_CONCAT(app.name, ' ')
    FROM items_applications iap
    INNER JOIN applications app on app.applicationid = iap.applicationid
    WHERE iap.itemid = ite.itemid), ' ', '\\ '), ',', ''), 'N.A.')
|| ' value_min=' || CAST(tre.value_min as char)
|| ',value_avg=' || CAST(tre.value_avg as char)
|| ',value_max=' || CAST(tre.value_max as char)
-- timestamp (in ms)
|| ' ' || CAST((tre.clock * 1000000000.) as char) as INLINE
,  CAST((tre.clock * 1000000000.) as char) as clock
FROM trends tre 
INNER JOIN items ite on ite.itemid = tre.itemid
INNER JOIN hosts hos on hos.hostid = ite.hostid
INNER JOIN hosts_groups hg on hg.hostid = hos.hostid
INNER JOIN hstgrp grp on grp.groupid = hg.groupid
WHERE grp.internal=0
   AND tre.clock > ##STARTDATE##
   AND tre.clock <= ##ENDDATE##;
`

const mysqlTrendsUInt string = `SELECT 
-- measurement
replace(replace(CASE
    WHEN (position('$2' in ite.name) > 0) AND (position('$4' in ite.name) > 0) 
      THEN replace(replace(ite.name, '$2', SUBSTRING_INDEX(SUBSTRING_INDEX(SUBSTRING(ite.key_, LOCATE('[',ite.key_) + 1, LOCATE(']',ite.key_) - LOCATE('[',ite.key_)-1),',',2), ',',-1)), '$4', SUBSTRING_INDEX(SUBSTRING(ite.key_, LOCATE('[',ite.key_) + 1, LOCATE(']',ite.key_) - LOCATE('[',ite.key_)-1),',',-1))
	WHEN (position('$1' in ite.name) > 0) AND (position('$2' in ite.name) > 0) 
      THEN replace(replace(ite.name, '$1', SUBSTRING_INDEX(SUBSTRING(ite.key_, LOCATE('[',ite.key_) + 1, LOCATE(']',ite.key_) - LOCATE('[',ite.key_)-1),',',1)), '$2', SUBSTRING_INDEX(SUBSTRING_INDEX(SUBSTRING(ite.key_, LOCATE('[',ite.key_) + 1, LOCATE(']',ite.key_) - LOCATE('[',ite.key_)-1),',',2), ',',-1))
    WHEN (position('$1' in ite.name) > 0) 
       THEN replace(ite.name, '$1', SUBSTRING_INDEX(SUBSTRING_INDEX(SUBSTRING(ite.key_, LOCATE('[',ite.key_) + 1, LOCATE(']',ite.key_) - LOCATE('[',ite.key_)-1),',',1), ',',-1))
    WHEN (position('$2' in ite.name) > 0) 
       THEN replace(ite.name, '$2', SUBSTRING_INDEX(SUBSTRING_INDEX(SUBSTRING(ite.key_, LOCATE('[',ite.key_) + 1, LOCATE(']',ite.key_) - LOCATE('[',ite.key_)-1),',',2), ',',-1))
	WHEN (position('$3' in ite.name) > 0)
       THEN replace(ite.name, '$3', SUBSTRING_INDEX(SUBSTRING_INDEX(SUBSTRING(ite.key_, LOCATE('[',ite.key_) + 1, LOCATE(']',ite.key_) - LOCATE('[',ite.key_)-1),',',-2), ',',1))
    WHEN (position('$1' in ite.name) > 0) AND (position('$3' in ite.name) > 0)
       THEN replace(replace(ite.name, '$1', SUBSTRING_INDEX(SUBSTRING_INDEX(SUBSTRING(ite.key_, LOCATE('[',ite.key_) + 1, LOCATE(']',ite.key_) - LOCATE('[',ite.key_)-1),',',1), ',',-1)), '$3', SUBSTRING_INDEX(SUBSTRING_INDEX(SUBSTRING(ite.key_, LOCATE('[',ite.key_) + 1, LOCATE(']',ite.key_) - LOCATE('[',ite.key_)-1),',',-2), ',',1))
    ELSE ite.name
  END, ',', ''), ' ', '\\ ') 
-- tags
|| ',host_name=' || replace(hos.name, ' ', '\\ ')
|| ',group_name=' || replace(grp.name, ' ', '\\ ')
|| ',applications=' || ifnull(replace(replace((SELECT GROUP_CONCAT(app.name, ' ')
    FROM items_applications iap
    INNER JOIN applications app on app.applicationid = iap.applicationid
    WHERE iap.itemid = ite.itemid), ' ', '\\ '), ',', ''), 'N.A.')
|| ' value_min=' || CAST(tre.value_min as char)
|| ',value_avg=' || CAST(tre.value_avg as char)
|| ',value_max=' || CAST(tre.value_max as char)
-- timestamp (in ms)
|| ' ' || CAST((tre.clock * 1000000000.) as char) as INLINE
,  CAST((tre.clock * 1000000000.) as char) as clock
FROM trends tre 
INNER JOIN items ite on ite.itemid = tre.itemid
INNER JOIN hosts hos on hos.hostid = ite.hostid
INNER JOIN hosts_groups hg on hg.hostid = hos.hostid
INNER JOIN hstgrp grp on grp.groupid = hg.groupid
WHERE grp.internal=0
   AND tre.clock > ##STARTDATE##
   AND tre.clock <= ##ENDDATE##;
`

const mysqlHistory string = `SELECT 
-- measurement
replace(replace(CASE
    WHEN (position('$2' in ite.name) > 0) AND (position('$4' in ite.name) > 0) 
      THEN replace(replace(ite.name, '$2', SUBSTRING_INDEX(SUBSTRING_INDEX(SUBSTRING(ite.key_, LOCATE('[',ite.key_) + 1, LOCATE(']',ite.key_) - LOCATE('[',ite.key_)-1),',',2), ',',-1)), '$4', SUBSTRING_INDEX(SUBSTRING(ite.key_, LOCATE('[',ite.key_) + 1, LOCATE(']',ite.key_) - LOCATE('[',ite.key_)-1),',',-1))
	WHEN (position('$1' in ite.name) > 0) AND (position('$2' in ite.name) > 0) 
      THEN replace(replace(ite.name, '$1', SUBSTRING_INDEX(SUBSTRING(ite.key_, LOCATE('[',ite.key_) + 1, LOCATE(']',ite.key_) - LOCATE('[',ite.key_)-1),',',1)), '$2', SUBSTRING_INDEX(SUBSTRING_INDEX(SUBSTRING(ite.key_, LOCATE('[',ite.key_) + 1, LOCATE(']',ite.key_) - LOCATE('[',ite.key_)-1),',',2), ',',-1))
    WHEN (position('$1' in ite.name) > 0) 
       THEN replace(ite.name, '$1', SUBSTRING_INDEX(SUBSTRING_INDEX(SUBSTRING(ite.key_, LOCATE('[',ite.key_) + 1, LOCATE(']',ite.key_) - LOCATE('[',ite.key_)-1),',',1), ',',-1))
    WHEN (position('$2' in ite.name) > 0) 
       THEN replace(ite.name, '$2', SUBSTRING_INDEX(SUBSTRING_INDEX(SUBSTRING(ite.key_, LOCATE('[',ite.key_) + 1, LOCATE(']',ite.key_) - LOCATE('[',ite.key_)-1),',',2), ',',-1))
	WHEN (position('$3' in ite.name) > 0)
       THEN replace(ite.name, '$3', SUBSTRING_INDEX(SUBSTRING_INDEX(SUBSTRING(ite.key_, LOCATE('[',ite.key_) + 1, LOCATE(']',ite.key_) - LOCATE('[',ite.key_)-1),',',-2), ',',1))
    WHEN (position('$1' in ite.name) > 0) AND (position('$3' in ite.name) > 0)
       THEN replace(replace(ite.name, '$1', SUBSTRING_INDEX(SUBSTRING_INDEX(SUBSTRING(ite.key_, LOCATE('[',ite.key_) + 1, LOCATE(']',ite.key_) - LOCATE('[',ite.key_)-1),',',1), ',',-1)), '$3', SUBSTRING_INDEX(SUBSTRING_INDEX(SUBSTRING(ite.key_, LOCATE('[',ite.key_) + 1, LOCATE(']',ite.key_) - LOCATE('[',ite.key_)-1),',',-2), ',',1))
    ELSE ite.name
  END, ',', ''), ' ', '\\ ') 
-- tags
|| ',host_name=' || replace(hos.name, ' ', '\\ ')
|| ',group_name=' || replace(grp.name, ' ', '\\ ')
|| ',applications=' || ifnull(replace(replace((SELECT GROUP_CONCAT(app.name, ' ')
    FROM items_applications iap
    INNER JOIN applications app on app.applicationid = iap.applicationid
    WHERE iap.itemid = ite.itemid), ' ', '\\ '), ',', ''), 'N.A.')
|| ' value=' || CAST(his.value as char)
-- timestamp (in ms)
|| ' ' || CAST((his.clock * 1000000000.) as char) as INLINE
,  CAST((his.clock * 1000000000.) as char) as clock
FROM history his
INNER JOIN items ite on ite.itemid = his.itemid
INNER JOIN hosts hos on hos.hostid = ite.hostid
INNER JOIN hosts_groups hg on hg.hostid = hos.hostid
INNER JOIN hstgrp grp on grp.groupid = hg.groupid
WHERE grp.internal=0
   AND his.clock > ##STARTDATE##
   AND his.clock <= ##ENDDATE##;
`

const mysqlHistoryUInt string = `SELECT 
-- measurement
replace(replace(CASE
    WHEN (position('$2' in ite.name) > 0) AND (position('$4' in ite.name) > 0) 
      THEN replace(replace(ite.name, '$2', SUBSTRING_INDEX(SUBSTRING_INDEX(SUBSTRING(ite.key_, LOCATE('[',ite.key_) + 1, LOCATE(']',ite.key_) - LOCATE('[',ite.key_)-1),',',2), ',',-1)), '$4', SUBSTRING_INDEX(SUBSTRING(ite.key_, LOCATE('[',ite.key_) + 1, LOCATE(']',ite.key_) - LOCATE('[',ite.key_)-1),',',-1))
	WHEN (position('$1' in ite.name) > 0) AND (position('$2' in ite.name) > 0) 
      THEN replace(replace(ite.name, '$1', SUBSTRING_INDEX(SUBSTRING(ite.key_, LOCATE('[',ite.key_) + 1, LOCATE(']',ite.key_) - LOCATE('[',ite.key_)-1),',',1)), '$2', SUBSTRING_INDEX(SUBSTRING_INDEX(SUBSTRING(ite.key_, LOCATE('[',ite.key_) + 1, LOCATE(']',ite.key_) - LOCATE('[',ite.key_)-1),',',2), ',',-1))
    WHEN (position('$1' in ite.name) > 0) 
       THEN replace(ite.name, '$1', SUBSTRING_INDEX(SUBSTRING_INDEX(SUBSTRING(ite.key_, LOCATE('[',ite.key_) + 1, LOCATE(']',ite.key_) - LOCATE('[',ite.key_)-1),',',1), ',',-1))
    WHEN (position('$2' in ite.name) > 0) 
       THEN replace(ite.name, '$2', SUBSTRING_INDEX(SUBSTRING_INDEX(SUBSTRING(ite.key_, LOCATE('[',ite.key_) + 1, LOCATE(']',ite.key_) - LOCATE('[',ite.key_)-1),',',2), ',',-1))
	WHEN (position('$3' in ite.name) > 0)
       THEN replace(ite.name, '$3', SUBSTRING_INDEX(SUBSTRING_INDEX(SUBSTRING(ite.key_, LOCATE('[',ite.key_) + 1, LOCATE(']',ite.key_) - LOCATE('[',ite.key_)-1),',',-2), ',',1))
    WHEN (position('$1' in ite.name) > 0) AND (position('$3' in ite.name) > 0)
       THEN replace(replace(ite.name, '$1', SUBSTRING_INDEX(SUBSTRING_INDEX(SUBSTRING(ite.key_, LOCATE('[',ite.key_) + 1, LOCATE(']',ite.key_) - LOCATE('[',ite.key_)-1),',',1), ',',-1)), '$3', SUBSTRING_INDEX(SUBSTRING_INDEX(SUBSTRING(ite.key_, LOCATE('[',ite.key_) + 1, LOCATE(']',ite.key_) - LOCATE('[',ite.key_)-1),',',-2), ',',1))
    ELSE ite.name
  END, ',', ''), ' ', '\\ ') 
-- tags
|| ',host_name=' || replace(hos.name, ' ', '\\ ')
|| ',group_name=' || replace(grp.name, ' ', '\\ ')
|| ',applications=' || ifnull(replace(replace((SELECT GROUP_CONCAT(app.name, ' ')
    FROM items_applications iap
    INNER JOIN applications app on app.applicationid = iap.applicationid
    WHERE iap.itemid = ite.itemid), ' ', '\\ '), ',', ''), 'N.A.')
|| ' value=' || CAST(his.value as char)
-- timestamp (in ms)
|| ' ' || CAST((his.clock * 1000000000.) as char) as INLINE
,  CAST((his.clock * 1000000000.) as char) as clock
FROM history_uint his
INNER JOIN items ite on ite.itemid = his.itemid
INNER JOIN hosts hos on hos.hostid = ite.hostid
INNER JOIN hosts_groups hg on hg.hostid = hos.hostid
INNER JOIN hstgrp grp on grp.groupid = hg.groupid
WHERE grp.internal=0
   AND his.clock > ##STARTDATE##
   AND his.clock <= ##ENDDATE##;
`
