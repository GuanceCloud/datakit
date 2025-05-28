-- 在每个并发会话中执行
DECLARE
  v_dummy NUMBER;
BEGIN
  FOR i IN 1..20000 LOOP -- 循环次数可以调整
    -- latch_test_pkg.do_nothing;
    v_dummy := latch_test_pkg.get_value();
  END LOOP;
END;
/
