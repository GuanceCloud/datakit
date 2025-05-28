-- 登录到数据库，例如使用 SCOTT 用户，或者其他有权限的用户
-- 如果 SCOTT 用户不存在或被锁定，请用其他用户或解锁 SCOTT

-- 创建一个测试表
CREATE TABLE locked_test_table (
    id NUMBER PRIMARY KEY,
    data VARCHAR2(100)
)
;;;

-- 插入一些数据
INSERT INTO locked_test_table (id, data) VALUES (1, 'Initial data');;;
INSERT INTO locked_test_table (id, data) VALUES (2, 'Another data');;;
COMMIT
;;;

-- 在会话1中锁定对象：打开第一个 SQL*Plus 窗口或 SQL Developer 连接（我们称之为 会话1）
-- 会话1
-- 更新表中的一行数据，这将对该行以及表本身施加锁
UPDATE locked_test_table
SET data = 'Data locked by session 1'
WHERE id = 1;

-- 重要：不要执行 COMMIT 或 ROLLBACK！此时锁已生效。

-- 然后执行 run.sql lock.sql
