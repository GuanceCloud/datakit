-- 可以在一个匿名 PL/SQL 块中执行
DECLARE
  v_dummy NUMBER;
BEGIN
  FOR i IN 1..50 LOOP -- 调整循环次数以控制CPU消耗和执行时间
    v_dummy := SIN(i) * COS(i) + POWER(i, 2);
  END LOOP;
END;
;;;
