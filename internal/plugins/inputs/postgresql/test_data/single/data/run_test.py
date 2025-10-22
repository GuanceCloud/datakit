import psycopg2
import time
import threading
import random
import os
import logging
from datetime import datetime
from psycopg2 import pool
from contextlib import contextmanager

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - Thread %(thread)d - %(levelname)s - %(message)s',
    handlers=[
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

# Database configuration
DB_CONFIG = {
    "dbname": os.getenv("DB_NAME", "test_db"),
    "user": os.getenv("DB_USER", "postgres"),
    "password": os.getenv("DB_PASSWORD", "postgres"),
    "host": os.getenv("DB_HOST", "postgres"),
    "port": os.getenv("DB_PORT", "5432")
}

# Test configuration
NUM_THREADS = int(os.getenv("NUM_THREADS", 10))
RUN_DURATION = int(os.getenv("RUN_DURATION", 300))
MAX_RETRIES = 3  # Number of retries for database operations
CONNECTION_POOL_SIZE = max(NUM_THREADS, 5)  # Connection pool size

# Operation types and their weights (probability)
OPERATIONS = [
    ("read", 20),
    ("update", 20),
    ("long_transaction", 10),
    ("table_lock", 5),
    ("schema_operation", 5),
    ("insert_users", 15),
    ("select_users", 15),
    ("large_data", 15),
    ("deadlock", 10),
    ("prepared_statement", 10),
]

# Connection pool initialization
connection_pool = None

def init_connection_pool():
    """Initialize database connection pool"""
    global connection_pool
    try:
        connection_pool = psycopg2.pool.ThreadedConnectionPool(
            minconn=1,
            maxconn=CONNECTION_POOL_SIZE,** DB_CONFIG
        )
        if connection_pool:
            logger.info("Database connection pool initialized successfully")
    except Exception as e:
        logger.error(f"Failed to initialize database connection pool: {e}")
        raise

@contextmanager
def get_db_connection():
    """Context manager for getting database connections"""
    conn = None
    try:
        conn = connection_pool.getconn()
        yield conn
    except Exception as e:
        logger.error(f"Failed to get database connection: {e}")
        raise
    finally:
        if conn:
            connection_pool.putconn(conn)

def random_sleep(min_seconds=0.1, max_seconds=2.0):
    """Sleep for a random period of time"""
    time.sleep(random.uniform(min_seconds, max_seconds))

def with_retry(operation, *args, **kwargs):
    """Database operation executor with retry mechanism"""
    for attempt in range(MAX_RETRIES):
        try:
            return operation(*args, **kwargs)
        except Exception as e:
            if attempt == MAX_RETRIES - 1:
                raise
            logger.warning(f"Operation failed, will retry (attempt {attempt + 1}/{MAX_RETRIES}): {e}")
            time.sleep(0.5 * (attempt + 1))  # Exponential backoff

def perform_read_operation(conn, thread_id):
    """Perform a read operation"""
    with conn.cursor() as cur:
        product_id = random.randint(1, 5)
        cur.execute("""
            SELECT * FROM test_schema.inventory 
            WHERE product_id = %s
        """, (product_id,))
        cur.fetchone()
        random_sleep()
    logger.info(f"Completed read operation, Product ID: {product_id}")

def perform_update_operation(conn, thread_id):
    """Perform an update operation"""
    original_autocommit = conn.autocommit
    try:
        conn.autocommit = False
        with conn.cursor() as cur:
            product_id = random.randint(1, 5)
            new_quantity = random.randint(1, 10)
            
            cur.execute("""
                SELECT * FROM test_schema.inventory 
                WHERE product_id = %s FOR UPDATE
            """, (product_id,))
            cur.fetchone()
            
            random_sleep()
            
            cur.execute("""
                UPDATE test_schema.inventory 
                SET quantity = quantity - %s,
                    last_updated = CURRENT_TIMESTAMP
                WHERE product_id = %s
            """, (new_quantity, product_id))
            
            conn.commit()
            logger.info(f"Completed update operation, Product ID: {product_id}, Quantity: {new_quantity}")
    except Exception as e:
        conn.rollback()
        raise e
    finally:
        conn.autocommit = original_autocommit

def perform_long_transaction(conn, thread_id):
    """Perform a long transaction"""
    original_autocommit = conn.autocommit
    try:
        conn.autocommit = False
        with conn.cursor() as cur:
            order_id = random.randint(1, 5)
            
            cur.execute("""
                SELECT * FROM test_schema.orders 
                WHERE id = %s FOR UPDATE
            """, (order_id,))
            cur.fetchone()
            
            time.sleep(random.uniform(5, 15))  # Long transaction
            
            new_status = random.choice(['processing', 'shipped', 'completed'])
            cur.execute("""
                UPDATE test_schema.orders 
                SET status = %s
                WHERE id = %s
            """, (new_status, order_id))
            
            conn.commit()
            logger.info(f"Completed long transaction, Order ID: {order_id}, New status: {new_status}")
    except Exception as e:
        conn.rollback()
        raise e
    finally:
        conn.autocommit = original_autocommit

def perform_table_lock(conn, thread_id):
    """Perform a table lock operation"""
    original_autocommit = conn.autocommit
    try:
        conn.autocommit = False
        with conn.cursor() as cur:
            table = random.choice(['orders', 'inventory'])
            lock_mode = random.choice([
                'SHARE MODE', 
                'ROW EXCLUSIVE MODE',
                'EXCLUSIVE MODE'
            ])
            
            cur.execute(f"""
                LOCK TABLE test_schema.{table} IN {lock_mode}
            """)
            
            time.sleep(random.uniform(2, 8))
            conn.commit()
            logger.info(f"Completed table lock, Table name: {table}, Lock mode: {lock_mode}")
    except Exception as e:
        conn.rollback()
        raise e
    finally:
        conn.autocommit = original_autocommit

def perform_schema_operation(conn, thread_id):
    """Perform a schema operation"""
    original_autocommit = conn.autocommit
    try:
        conn.autocommit = True  # Schema operations need auto-commit
        with conn.cursor() as cur:
            if random.random() < 0.3: 
                index_name = f"temp_idx_{thread_id}_{int(time.time())}"
                cur.execute(f"""
                    CREATE INDEX {index_name} 
                    ON test_schema.orders(customer_id)
                """)
                time.sleep(1)
                cur.execute(f"DROP INDEX test_schema.{index_name}")
                logger.info(f"Completed index operation, Index name: {index_name}")
    except Exception as e:
        logger.error(f"Schema operation failed: {e}")
        raise e
    finally:
        conn.autocommit = original_autocommit

def perform_insert_users(conn, thread_id):
    """Insert user data"""
    conn = get_new_db_connection("app_db")
    with conn.cursor() as cur:
        username = f"user_{thread_id}_{int(time.time())}"
        email = f"{username}@example.com"
        cur.execute("""
            INSERT INTO users (username, email)
            VALUES (%s, %s)
        """, (username, email))
        conn.commit()
        logger.info(f"Inserted user: {username}")

def perform_large_data(conn, thread_id):
    """Insert large data"""
    with conn.cursor() as cur:
        start_time = time.time()
        for i in range(100):
            cur.execute("""
                INSERT INTO large_data (value, num)
                VALUES (%s, %s)
            """, (f"test data {i}" * 10, i * 0.123456789))
            
            if i % 1000 == 0:
                conn.commit()
                logger.info(f"inserted {i} rows")
        
        conn.commit()
        end_time = time.time()
        logger.info(f"inserted {i} rows, cost {end_time - start_time:.2f} seconds")

        start_time = time.time()
        cur.execute("""
            SELECT * FROM large_data 
            ORDER BY value DESC
        """)
        cur.fetchmany(10)
        
        start_time = time.time()
        cur.execute("""
            SELECT TRUNC(num, 0) as num_group, COUNT(*) 
            FROM large_data 
            GROUP BY TRUNC(num, 0)
            ORDER BY num_group DESC
        """)
        cur.fetchmany(10)
            

def perform_select_users(conn, thread_id):
    """Query user data"""
    try:
        with conn.cursor() as cur:
            usernames = ['alice', 'bob', 'charlie', f"user_{thread_id}"]
            username = random.choice(usernames)
            
            cur.execute("""
                SELECT * FROM users 
                WHERE username = %s
            """, (username,)) 
            
            result = cur.fetchall()
            logger.info(f"Queried user {username}, returned {len(result)} records (index used)")
    finally:
        conn.close()
def perform_deadlock_operation(conn, thread_id):
    conn = get_new_db_connection("lock_db")
    """Create deadlock scenario by acquiring locks in opposite orders"""
    original_autocommit = conn.autocommit
    try:
        conn.autocommit = False
        with conn.cursor() as cur:
            # Alternate lock acquisition order based on thread ID parity
            if thread_id % 2 == 0:
                # Even threads: lock product 1 first, then product 2
                cur.execute("""
                    SELECT * FROM test_schema.inventory 
                    WHERE product_id = 1 FOR UPDATE
                """)
                cur.fetchone()
                logger.info(f"Thread {thread_id} locked product_id=1")
                random_sleep(1, 3)  # Create window for lock contention
                
                cur.execute("""
                    SELECT * FROM test_schema.inventory 
                    WHERE product_id = 2 FOR UPDATE
                """)
                cur.fetchone()
                logger.info(f"Thread {thread_id} locked product_id=2")
            else:
                # Odd threads: lock product 2 first, then product 1
                cur.execute("""
                    SELECT * FROM test_schema.inventory 
                    WHERE product_id = 2 FOR UPDATE
                """)
                cur.fetchone()
                logger.info(f"Thread {thread_id} locked product_id=2")
                random_sleep(1, 3)  # Create window for lock contention
                
                cur.execute("""
                    SELECT * FROM test_schema.inventory 
                    WHERE product_id = 1 FOR UPDATE
                """)
                cur.fetchone()
                logger.info(f"Thread {thread_id} locked product_id=1")

            # Perform update after acquiring both locks
            cur.execute("""
                UPDATE test_schema.inventory 
                SET quantity = quantity - 1, 
                    last_updated = CURRENT_TIMESTAMP
                WHERE product_id IN (1, 2)
            """)
            conn.commit()
            logger.info(f"Thread {thread_id} completed deadlock operation")
            
    except psycopg2.OperationalError as e:
        if "deadlock detected" in str(e):
            logger.warning(f"Thread {thread_id} encountered expected deadlock: {e}")
        conn.rollback()
    except Exception as e:
        conn.rollback()
        logger.error(f"Thread {thread_id} deadlock operation failed: {e}")
    finally:
        conn.autocommit = original_autocommit

def perform_prepared_statement(conn, thread_id):
    with conn.cursor() as cur:
        try:
            query = """
                SELECT customer_id FROM test_schema.orders 
                WHERE status = $1 
            """
            logger.info(f"Thread {thread_id} prepared statement created")
            
            for _ in range(random.randint(3, 10)):
                status = random.choice(["shipped", "pending", "cancelled"])
                cur.execute(query, (status,)) 
                cur.fetchone()
                random_sleep(0.2, 0.5)
            
            logger.info(f"Thread {thread_id} completed prepared statement operations")
        except Exception as e:
            logger.error(f"Prepared statement operation failed: {e}")
            raise

def perform_operation(thread_id):
    """Main function for executing database operations"""
    end_time = time.time() + RUN_DURATION
    logger.info(f"Thread {thread_id} started, duration: {RUN_DURATION} seconds")
    
    # Build operation list (based on weights)
    operation_list = []
    for op, weight in OPERATIONS:
        operation_list.extend([op] * weight)
    
    try:
        while time.time() < end_time:
            operation = random.choice(operation_list)
            
            try:
                with get_db_connection() as conn:
                    if operation == "read":
                        with_retry(perform_read_operation, conn, thread_id)
                    elif operation == "update":
                        with_retry(perform_update_operation, conn, thread_id)
                    elif operation == "long_transaction":
                        with_retry(perform_long_transaction, conn, thread_id)
                    elif operation == "table_lock":
                        with_retry(perform_table_lock, conn, thread_id)
                    elif operation == "schema_operation":
                        with_retry(perform_schema_operation, conn, thread_id)
                    elif operation == "insert_users":
                        with_retry(perform_insert_users, conn, thread_id)
                    elif operation == "select_users":
                        with_retry(perform_select_users, conn, thread_id)
                    elif operation == "large_data":
                        with_retry(perform_large_data, conn, thread_id)
                    elif operation == "deadlock":
                        with_retry(perform_deadlock_operation, conn, thread_id)
                    elif operation == "prepared_statement":
                        with_retry(perform_prepared_statement, conn, thread_id) 

                # Random interval between operations
                time.sleep(random.uniform(1, 3))
                
            except Exception as e:
                logger.error(f"Thread {thread_id} failed to execute operation {operation}: {e}")
                time.sleep(1)  # Wait a bit before retrying after error
    
    except Exception as e:
        logger.error(f"Thread {thread_id} encountered a fatal error: {e}")
    
    finally:
        logger.info(f"Thread {thread_id} finished running")

def get_new_db_connection(dbname):
    # copy DB_CONFIG
    new_db_config = DB_CONFIG.copy()
    new_db_config["dbname"] = dbname
    return psycopg2.connect(**new_db_config)

def main():
    """Main function"""
    try:
        # Initialize connection pool
        init_connection_pool()
        
        while True:
            logger.info(f"Starting new test round, number of threads: {NUM_THREADS}, duration: {RUN_DURATION} seconds")
            
            threads = []
            for i in range(NUM_THREADS):
                thread = threading.Thread(target=perform_operation, args=(i,), name=f"DBTestThread-{i}")
                threads.append(thread)
                thread.start()
                time.sleep(0.5)  # Stagger thread start times
            
            # Wait for all threads to complete
            for thread in threads:
                thread.join()
            
            logger.info("Test round completed, resting for 10 seconds before next round")
            time.sleep(10)
            
    except KeyboardInterrupt:
        logger.info("Program interrupted by user")
    except Exception as e:
        logger.error(f"Program encountered an error: {e}")
    finally:
        if connection_pool:
            connection_pool.closeall()
            logger.info("Database connection pool has been closed")

if __name__ == "__main__":
    main()
    