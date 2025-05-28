-- 创建一个小表用于测试 (如果不存在)
-- CREATE TABLE session_test_data (id NUMBER, val VARCHAR2(100));
-- BEGIN FOR i IN 1..1000 LOOP INSERT INTO session_test_data VALUES (i, DBMS_RANDOM.STRING('X', 50)); END LOOP; COMMIT; END;

-- 查询并排序，这会消耗一些PGA
SELECT id, val
FROM (
    SELECT id, val, ROW_NUMBER() OVER (ORDER BY DBMS_RANDOM.VALUE) as rn
    FROM session_test_data -- 使用一个包含少量数据的表
    WHERE ROWNUM <= 500 -- 限制行数，避免过多IO
)
ORDER BY val DESC, id ASC
;;;

SELECT COUNT(*) FROM ALL_OBJECTS WHERE ROWNUM <= 1000 -- 对数据字典进行适度查询
