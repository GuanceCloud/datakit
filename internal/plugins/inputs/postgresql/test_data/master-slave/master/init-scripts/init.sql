
CREATE DATABASE app_db;
CREATE DATABASE test_db;
SELECT datname FROM pg_database WHERE datistemplate = false;

\c app_db;
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

INSERT INTO users (username, email) VALUES
('alice', 'alice@example.com'),
('bob', 'bob@example.com'),
('charlie', 'charlie@example.com');

-- create datakit account
create user datakit with password '123456';
grant pg_monitor to datakit;
grant SELECT ON pg_stat_database to datakit;
