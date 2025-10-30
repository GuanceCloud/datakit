CREATE DATABASE app_db;
CREATE DATABASE app_db_1;
CREATE DATABASE audit_db;
CREATE DATABASE analytics_db;
CREATE DATABASE lock_db;

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

\c app_db_1;
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

\c audit_db;
CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    action VARCHAR(100) NOT NULL,
    user_id INT,
    action_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO audit_logs (action, user_id) VALUES
('login', 1),
('create_post', 1),
('login', 2),
('delete_comment', 2);

\c analytics_db;
CREATE TABLE metrics (
    id SERIAL PRIMARY KEY,
    metric_name VARCHAR(50) NOT NULL,
    metric_value NUMERIC(10,2) NOT NULL,
    recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO metrics (metric_name, metric_value) VALUES
('active_users', 156),
('page_views', 3240),
('conversion_rate', 4.8);

\c lock_db;

CREATE SCHEMA IF NOT EXISTS test_schema;

CREATE TABLE IF NOT EXISTS test_schema.inventory (
    product_id SERIAL PRIMARY KEY,
    product_name VARCHAR(100) NOT NULL,
    quantity INT NOT NULL,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO test_schema.inventory (product_name, quantity) VALUES
('Laptop', 100),
('Phone', 200),
('Tablet', 150),
('Headphones', 300),
('Charger', 500);
\c test_db

CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

CREATE SCHEMA IF NOT EXISTS test_schema;

CREATE TABLE IF NOT EXISTS test_schema.orders (
    id SERIAL PRIMARY KEY,
    customer_id INT NOT NULL,
    product_id INT NOT NULL,
    quantity INT NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS test_schema.inventory (
    product_id SERIAL PRIMARY KEY,
    product_name VARCHAR(100) NOT NULL,
    quantity INT NOT NULL,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO test_schema.inventory (product_name, quantity) VALUES
('Laptop', 100),
('Phone', 200),
('Tablet', 150),
('Headphones', 300),
('Charger', 500);

INSERT INTO test_schema.orders (customer_id, product_id, quantity, status) VALUES
(101, 1, 2, 'pending'),
(102, 2, 1, 'completed'),
(103, 3, 3, 'pending'),
(104, 4, 5, 'shipped'),
(105, 5, 10, 'pending');

CREATE INDEX IF NOT EXISTS idx_orders_status ON test_schema.orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON test_schema.orders(customer_id);

CREATE TABLE large_data (
    id SERIAL PRIMARY KEY,
    value TEXT,
    num NUMERIC,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

-- create datakit account
create user datakit with password '123456';
grant pg_monitor to datakit;
grant SELECT ON pg_stat_database to datakit;

-- DBM role

\c test_db;
CREATE SCHEMA datakit;
GRANT USAGE ON SCHEMA datakit TO datakit;
GRANT USAGE ON SCHEMA public TO datakit;
GRANT USAGE ON SCHEMA test_schema TO datakit;
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

CREATE OR REPLACE FUNCTION datakit.explain_statement(
   l_query TEXT,
   OUT explain JSON
)
RETURNS SETOF JSON AS
$$
DECLARE
curs REFCURSOR;
plan JSON;

BEGIN
   OPEN curs FOR EXECUTE pg_catalog.concat('EXPLAIN (FORMAT JSON) ', l_query);
   FETCH curs INTO plan;
   CLOSE curs;
   RETURN QUERY SELECT plan;
END;
$$
LANGUAGE 'plpgsql'
RETURNS NULL ON NULL INPUT
SECURITY DEFINER;