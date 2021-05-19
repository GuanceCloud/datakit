package zabbix

func pgSQL(tablename string) string {
	switch tablename {
	case "history":
		return pgsqlHistory
	case "history_uint":
		return pgsqlHistoryUInt
	case "trends":
		return pgsqlTrends
	case "trends_uint":
		return pgsqlTrendsUInt
	default:
		panic("unrecognized tablename")
	}
}

const pgsqlTrends string = `SELECT 
-- measurement
replace(replace(CASE
    WHEN (position('$2' in ite.name) > 0) AND (position('$4' in ite.name) > 0) 
      THEN replace(replace(ite.name, '$2', split_part(substring(ite.key_ FROM '\[(.+)\]'), ',', 2)), '$4', split_part(substring(ite.key_ FROM '\[(.+)\]'), ',', 4))
    WHEN (position('$1' in ite.name) > 0) AND (position('$2' in ite.name) > 0) 
      THEN replace(replace(ite.name, '$1', split_part(substring(ite.key_ FROM '\[(.+)\]'), ',', 1)), '$2', split_part(substring(ite.key_ FROM '\[(.+)\]'), ',', 2))
    WHEN (position('$1' in ite.name) > 0) 
       THEN replace(ite.name, '$1', split_part(substring(ite.key_ FROM '\[(.+)\]'), ',', 1))
    WHEN (position('$2' in ite.name) > 0) 
       THEN replace(ite.name, '$2', split_part(substring(ite.key_ FROM '\[(.+)\]'), ',', 2))
    WHEN (position('$3' in ite.name) > 0)
       THEN replace(ite.name, '$3', split_part(substring(ite.key_ FROM '\[(.+)\]'), ',', 3))
    WHEN (position('$1' in ite.name) > 0) AND (position('$3' in ite.name) > 0)
       THEN replace(replace(ite.name, '$1', split_part(substring(ite.key_ FROM '\[(.+)\]'), ',', 1)), '$3', split_part(substring(ite.key_ FROM '\[(.+)\]'), ',', 3))
    ELSE ite.name
  END, ',', ''), ' ', '\ ') 
-- tags
|| ',host_name=' || replace(hos.name, ' ', '\ ')
|| ',group_name=' || replace(grp.name, ' ', '\ ')
|| ',applications=' || coalesce(replace((SELECT string_agg(app.name, ' | ')
    FROM public.items_applications iap
    INNER JOIN public.applications app on app.applicationid = iap.applicationid
    WHERE iap.itemid = ite.itemid), ' ', '\ '), 'N.A.')
|| ' value_min=' || CAST(tre.value_min as varchar(32))
|| ',value_avg=' || CAST(tre.value_avg as varchar(32))
|| ',value_max=' || CAST(tre.value_max as varchar(32))
-- timestamp (in ms)
|| ' ' || CAST((tre.clock * 1000000000.) as char(19)) as INLINE
,  CAST((tre.clock * 1000000000.) as char(19)) as clock
FROM public.trends tre
INNER JOIN public.items ite on ite.itemid = tre.itemid
INNER JOIN public.hosts hos on hos.hostid = ite.hostid
INNER JOIN public.hosts_groups hg on hg.hostid = hos.hostid
INNER JOIN public.hstgrp grp on grp.groupid = hg.groupid
WHERE grp.internal=0
   AND tre.clock > ##STARTDATE##
   AND tre.clock <= ##ENDDATE##;
`
const pgsqlTrendsUInt string = `SELECT 
-- measurement
replace(replace(CASE
    WHEN (position('$2' in ite.name) > 0) AND (position('$4' in ite.name) > 0) 
      THEN replace(replace(ite.name, '$2', split_part(substring(ite.key_ FROM '\[(.+)\]'), ',', 2)), '$4', split_part(substring(ite.key_ FROM '\[(.+)\]'), ',', 4))
    WHEN (position('$1' in ite.name) > 0) AND (position('$2' in ite.name) > 0) 
      THEN replace(replace(ite.name, '$1', split_part(substring(ite.key_ FROM '\[(.+)\]'), ',', 1)), '$2', split_part(substring(ite.key_ FROM '\[(.+)\]'), ',', 2))
    WHEN (position('$1' in ite.name) > 0) 
       THEN replace(ite.name, '$1', split_part(substring(ite.key_ FROM '\[(.+)\]'), ',', 1))
    WHEN (position('$2' in ite.name) > 0) 
       THEN replace(ite.name, '$2', split_part(substring(ite.key_ FROM '\[(.+)\]'), ',', 2))
    WHEN (position('$3' in ite.name) > 0)
       THEN replace(ite.name, '$3', split_part(substring(ite.key_ FROM '\[(.+)\]'), ',', 3))
    WHEN (position('$1' in ite.name) > 0) AND (position('$3' in ite.name) > 0)
       THEN replace(replace(ite.name, '$1', split_part(substring(ite.key_ FROM '\[(.+)\]'), ',', 1)), '$3', split_part(substring(ite.key_ FROM '\[(.+)\]'), ',', 3))
    ELSE ite.name
  END, ',', ''), ' ', '\ ') 
-- tags
|| ',host_name=' || replace(hos.name, ' ', '\ ')
|| ',group_name=' || replace(grp.name, ' ', '\ ')
|| ',applications=' || coalesce(replace((SELECT string_agg(app.name, ' | ')
    FROM public.items_applications iap
    INNER JOIN public.applications app on app.applicationid = iap.applicationid
    WHERE iap.itemid = ite.itemid), ' ', '\ '), 'N.A.')
|| ' value_min=' || CAST(tre.value_min as varchar(32))
|| ',value_avg=' || CAST(tre.value_avg as varchar(32))
|| ',value_max=' || CAST(tre.value_max as varchar(32))
-- timestamp (in ms)
|| ' ' || CAST((tre.clock * 1000000000.) as char(19)) as INLINE
,  CAST((tre.clock * 1000000000.) as char(19)) as clock
FROM public.trends_uint tre
INNER JOIN public.items ite on ite.itemid = tre.itemid
INNER JOIN public.hosts hos on hos.hostid = ite.hostid
INNER JOIN public.hosts_groups hg on hg.hostid = hos.hostid
INNER JOIN public.hstgrp grp on grp.groupid = hg.groupid
WHERE grp.internal=0
   AND tre.clock > ##STARTDATE##
   AND tre.clock <= ##ENDDATE##;
`

const pgsqlHistory string = `SELECT 
-- measurement
replace(replace(CASE
    WHEN (position('$2' in ite.name) > 0) AND (position('$4' in ite.name) > 0) 
      THEN replace(replace(ite.name, '$2', split_part(substring(ite.key_ FROM '\[(.+)\]'), ',', 2)), '$4', split_part(substring(ite.key_ FROM '\[(.+)\]'), ',', 4))
    WHEN (position('$1' in ite.name) > 0) AND (position('$2' in ite.name) > 0) 
      THEN replace(replace(ite.name, '$1', split_part(substring(ite.key_ FROM '\[(.+)\]'), ',', 1)), '$2', split_part(substring(ite.key_ FROM '\[(.+)\]'), ',', 2))
    WHEN (position('$1' in ite.name) > 0) 
       THEN replace(ite.name, '$1', split_part(substring(ite.key_ FROM '\[(.+)\]'), ',', 1))
    WHEN (position('$2' in ite.name) > 0) 
       THEN replace(ite.name, '$2', split_part(substring(ite.key_ FROM '\[(.+)\]'), ',', 2))
    WHEN (position('$3' in ite.name) > 0)
       THEN replace(ite.name, '$3', split_part(substring(ite.key_ FROM '\[(.+)\]'), ',', 3))
    WHEN (position('$1' in ite.name) > 0) AND (position('$3' in ite.name) > 0)
       THEN replace(replace(ite.name, '$1', split_part(substring(ite.key_ FROM '\[(.+)\]'), ',', 1)), '$3', split_part(substring(ite.key_ FROM '\[(.+)\]'), ',', 3))
    ELSE ite.name
  END, ',', ''), ' ', '\ ') 
-- tags
|| ',host_name=' || replace(hos.name, ' ', '\ ')
|| ',group_name=' || replace(grp.name, ' ', '\ ')
|| ',applications=' || coalesce(replace((SELECT string_agg(app.name, ' | ')
    FROM public.items_applications iap
    INNER JOIN public.applications app on app.applicationid = iap.applicationid
    WHERE iap.itemid = ite.itemid), ' ', '\ '), 'N.A.')
|| ' value=' || CAST(his.value as varchar(32))
-- timestamp (in ms)
|| ' ' || CAST((his.clock * 1000000000.) + his.ns as char(19)) as INLINE
,  CAST((his.clock * 1000000000.) as char(19)) as clock
FROM public.history his
INNER JOIN public.items ite on ite.itemid = his.itemid
INNER JOIN public.hosts hos on hos.hostid = ite.hostid
INNER JOIN public.hosts_groups hg on hg.hostid = hos.hostid
INNER JOIN public.hstgrp grp on grp.groupid = hg.groupid
WHERE grp.internal=0
   AND his.clock > ##STARTDATE##
   AND his.clock <= ##ENDDATE##;
`

const pgsqlHistoryUInt string = `SELECT 
-- measurement
replace(replace(CASE
    WHEN (position('$2' in ite.name) > 0) AND (position('$4' in ite.name) > 0) 
      THEN replace(replace(ite.name, '$2', split_part(substring(ite.key_ FROM '\[(.+)\]'), ',', 2)), '$4', split_part(substring(ite.key_ FROM '\[(.+)\]'), ',', 4))
    WHEN (position('$1' in ite.name) > 0) AND (position('$2' in ite.name) > 0) 
      THEN replace(replace(ite.name, '$1', split_part(substring(ite.key_ FROM '\[(.+)\]'), ',', 1)), '$2', split_part(substring(ite.key_ FROM '\[(.+)\]'), ',', 2))
    WHEN (position('$1' in ite.name) > 0) 
       THEN replace(ite.name, '$1', split_part(substring(ite.key_ FROM '\[(.+)\]'), ',', 1))
	  WHEN (position('$2' in ite.name) > 0) 
       THEN replace(ite.name, '$2', split_part(substring(ite.key_ FROM '\[(.+)\]'), ',', 2))
    WHEN (position('$3' in ite.name) > 0)
       THEN replace(ite.name, '$3', split_part(substring(ite.key_ FROM '\[(.+)\]'), ',', 3))
    WHEN (position('$1' in ite.name) > 0) AND (position('$3' in ite.name) > 0)
       THEN replace(replace(ite.name, '$1', split_part(substring(ite.key_ FROM '\[(.+)\]'), ',', 1)), '$3', split_part(substring(ite.key_ FROM '\[(.+)\]'), ',', 3))
    ELSE ite.name
  END, ',', ''), ' ', '\ ') 
-- tags
|| ',host_name=' || replace(hos.name, ' ', '\ ')
|| ',group_name=' || replace(grp.name, ' ', '\ ')
|| ',applications=' || coalesce(replace((SELECT string_agg(app.name, ' | ')
    FROM public.items_applications iap
    INNER JOIN public.applications app on app.applicationid = iap.applicationid
    WHERE iap.itemid = ite.itemid), ' ', '\ '), 'N.A.')
|| ' value=' || CAST(his.value as varchar(32))
-- timestamp (in ms)
|| ' ' || CAST((his.clock * 1000000000.) + his.ns as char(19)) as INLINE
,  CAST((his.clock * 1000000000.) as char(19)) as clock
FROM public.history_uint his
INNER JOIN public.items ite on ite.itemid = his.itemid
INNER JOIN public.hosts hos on hos.hostid = ite.hostid
INNER JOIN public.hosts_groups hg on hg.hostid = hos.hostid
INNER JOIN public.hstgrp grp on grp.groupid = hg.groupid
WHERE grp.internal=0
   AND his.clock > ##STARTDATE##
   AND his.clock <= ##ENDDATE##;
`
