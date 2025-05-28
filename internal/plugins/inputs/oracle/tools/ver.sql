---------------------------------------------------
-- show oracle version
---------------------------------------------------
SELECT
banner,
REGEXP_SUBSTR(banner, 'Release (\d+\.\d+\.\d+\.\d+\.\d+)', 1, 1, NULL, 1) AS full_version
FROM v$version
WHERE banner LIKE 'Oracle%'
