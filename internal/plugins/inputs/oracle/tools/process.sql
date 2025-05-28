---------------------------------------------------
-- process
---------------------------------------------------
-- For oracle 12
SELECT 
	name pdb_name, 
	pname,
	pid, 
	program,
	nvl(pga_used_mem,0) pga_used_mem, 
	nvl(pga_alloc_mem,0) pga_alloc_mem, 
	nvl(pga_freeable_mem,0) pga_freeable_mem, 
	-- nvl(CPU_USED,0) cpu_used,
	nvl(pga_max_mem,0) pga_max_mem
  FROM v$process p, v$containers c
  WHERE
  	c.con_id(+) = p.con_id
;;;

-- For oracle 11
SELECT
    PROGRAM,
    PGA_USED_MEM,
    PGA_ALLOC_MEM,
    PGA_FREEABLE_MEM,
    PGA_MAX_MEM
FROM GV$PROCESS;;;

---------------------------------------------------
-- 更细致的后台进程分类 (基于 PNAME)
---------------------------------------------------
SELECT
    p.PNAME, -- 后台进程的名称，如 PMON, LGWR 等
    COUNT(*) AS process_count
FROM V$PROCESS p
WHERE p.BACKGROUND = 1 AND p.PNAME IS NOT NULL
GROUP BY p.PNAME
;;;

---区分后台进程和用户服务器进程
SELECT
    CASE WHEN p.BACKGROUND = 1 THEN 'BACKGROUND' ELSE 'SERVER' END AS process_type,
    COUNT(*) AS process_count
FROM V$PROCESS p
GROUP BY CASE WHEN p.BACKGROUND = 1 THEN 'BACKGROUND' ELSE 'SERVER' END
;;;
---------------------------------------------------
-- 进程总数
---------------------------------------------------
-- SELECT COUNT(*) FROM V$PROCESS

---------------------------------------------------
-- 假设通过关联 V$SESSION 的 STATUS='ACTIVE' 来判断
---------------------------------------------------
SELECT COUNT(*) AS active_server_processes
FROM V$PROCESS p
JOIN V$SESSION s ON p.ADDR = s.PADDR
WHERE (p.BACKGROUND IS NULL OR p.BACKGROUND = 0) AND s.STATUS = 'ACTIVE'
;;;

---------------------------------------------------
-- Top N PGA 使用者
---------------------------------------------------
SELECT * FROM (
    SELECT p.PID, p.SPID, p.USERNAME, p.PROGRAM, p.PGA_USED_MEM
    FROM V$PROCESS p
    WHERE p.BACKGROUND IS NULL OR p.BACKGROUND = 0 -- 服务器进程
    ORDER BY p.PGA_USED_MEM DESC
) WHERE ROWNUM <= 10
;;;

---------------------------------------------------
-- PGA 总和
---------------------------------------------------
SELECT
    SUM(p.PGA_USED_MEM) AS total_pga_used_mem,         -- 当前已使用的PGA内存
    SUM(p.PGA_ALLOC_MEM) AS total_pga_alloc_mem,       -- 当前已分配的PGA内存
    SUM(p.PGA_FREEABLE_MEM) AS total_pga_freeable_mem, -- 可释放回操作系统的PGA内存
    SUM(p.PGA_MAX_MEM) AS total_pga_max_mem            -- 进程生命周期内PGA使用的峰值
FROM V$PROCESS p
