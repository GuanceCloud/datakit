# 先决条件
# 创建一个小表用于测试 (如果不存在)
# CREATE TABLE session_test_data (id NUMBER, val VARCHAR2(100));
# BEGIN FOR i IN 1..1000 LOOP INSERT INTO session_test_data VALUES (i, DBMS_RANDOM.STRING('X', 50)); END LOOP; COMMIT; END;

./session.t \
	-conn "oracle://sys:123456@8.153.108.66:1522/XE,oracle://sys:123456@8.153.108.66:1521/XE" \
	-session-count 30 \
	-loop-delay 150ms
