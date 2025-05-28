CREATE OR REPLACE PACKAGE latch_test_pkg AS
  PROCEDURE do_nothing;
  FUNCTION get_value RETURN NUMBER;
END latch_test_pkg;
/

CREATE OR REPLACE PACKAGE BODY latch_test_pkg AS
  PROCEDURE do_nothing IS
  BEGIN
    NULL; -- 只是为了被调用
  END do_nothing;

  FUNCTION get_value RETURN NUMBER IS
    v_num NUMBER;
  BEGIN
    SELECT COUNT(*) INTO v_num FROM dual; -- 任意简单查询
    RETURN v_num;
  END get_value;
END latch_test_pkg;
/
