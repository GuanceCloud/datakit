
---------------------------------------------------
-- oracle system
---------------------------------------------------
-- V$CON_SYSMETRIC displays the system metric values captured for the most
-- current time interval for the PDB long duration (60-second) system metrics.
SELECT 
	metric_name,
	value, 
	metric_unit, 
	name pdb_name 
  FROM v$con_sysmetric s, v$containers c 
  WHERE s.con_id = c.con_id(+);;;

-- V$SYSMETRIC displays the system metric values captured for the most
-- current time interval for both the long duration (60-second) and
-- short duration (15-second) system metrics.
SELECT 
	metric_name,
	value, 
	metric_unit
  FROM v$sysmetric s, v$containers c 
  WHERE s.con_id = c.con_id(+)
