import redis
import threading
import time
import random
import string
import uuid
import argparse

# --- 全局控制 ---
stop_event = threading.Event()
threads = []

# --- 模拟器函数 ---

def simulate_blocked_client(r_conn, client_id):
    """使用 BLPOP 创建一个永久阻塞的客户端"""
    try:
        r_conn.client_setname(f"blocked_client_{client_id}")
        print(f"[Thread-{client_id}]  -> Blocked client started.")
        blocking_key = f"blocking_queue:{uuid.uuid4()}"
        r_conn.blpop(blocking_key, timeout=0)
    except (redis.exceptions.ConnectionError, redis.exceptions.ResponseError):
        print(f"[Thread-{client_id}]  -> Blocked client connection closed.")
    except Exception as e:
        print(f"Error in blocked client {client_id}: {e}")

def simulate_transaction_client(r_conn, client_id):
    """创建一个处于 MULTI 状态但永不 EXEC 的客户端"""
    try:
        r_conn.client_setname(f"txn_client_{client_id}")
        
        # 使用 execute_command 发送原生的 MULTI 命令
        r_conn.execute_command('MULTI')
        
        print(f"[Thread-{client_id}]  -> In-Transaction client started.")

        # 在事务中排队一个命令，使其在 client list 中可见
        r_conn.set("in_txn", "this_is_queued")
        
        # 保持连接活动
        while not stop_event.is_set():
            time.sleep(1)
    except Exception as e:
        print(f"Error in transaction client {client_id}: {e}")

def simulate_pubsub_client(r_conn, client_id):
    """创建一个订阅了频道的 Pub/Sub 客户端"""
    try:
        r_conn.client_setname(f"pubsub_client_{client_id}")
        p = r_conn.pubsub()
        channel_name = f"channel:{uuid.uuid4()}"
        p.subscribe(channel_name)
        print(f"[Thread-{client_id}]  -> Pub/Sub client started on channel '{channel_name}'.")
        for _ in p.listen():
            if stop_event.is_set():
                break
    except Exception as e:
        print(f"Error in pubsub client {client_id}: {e}")

def simulate_idle_client(r_conn, client_id):
    """创建一个长时间存活但空闲的客户端"""
    try:
        r_conn.client_setname(f"idle_client_{client_id}")
        print(f"[Thread-{client_id}]  -> Idle/Long-lived client started.")
        while not stop_event.is_set():
            time.sleep(1)
    except Exception as e:
        print(f"Error in idle client {client_id}: {e}")

def simulate_monitor_client(r_conn, client_id):
    """创建一个处于 MONITOR 模式的客户端 (危险操作!)"""
    try:
        r_conn.client_setname(f"monitor_client_{client_id}")
        print(f"[Thread-{client_id}]  -> MONITOR client started. CAUTION: This impacts server performance.")
        m = r_conn.monitor()
        for _ in m.listen():
            if stop_event.is_set():
                break
    except Exception as e:
        print(f"Error in monitor client {client_id}: {e}")

def simulate_large_obuf_client(r_conn, client_id):
    """创建一个有巨大输出缓冲区的客户端"""
    try:
        r_conn.client_setname(f"large_obuf_client_{client_id}")
        
        list_key = f"large_list:{uuid.uuid4()}"
        print(f"[Thread-{client_id}]  -> Preparing large data for obuf test...")
        pipeline = r_conn.pipeline(transaction=False)
        for _ in range(5000):
            pipeline.lpush(list_key, ''.join(random.choices(string.ascii_letters, k=1024)))
        pipeline.execute()
        r_conn.expire(list_key, 60)

        print(f"[Thread-{client_id}]  -> Requesting large data and going to sleep...")
        r_conn.lrange(list_key, 0, -1)
        print(f"[Thread-{client_id}]  -> Large obuf client is now 'stuck'. Check CLIENT LIST.")
        while not stop_event.is_set():
            time.sleep(1)
    except Exception as e:
        print(f"Error in large obuf client {client_id}: {e}")


# --- 主程序 (无改动) ---
if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="Redis Client State Simulator. Creates clients in various states for monitoring and testing.",
        formatter_class=argparse.RawTextHelpFormatter
    )
    parser.add_argument('-H', '--host', default='localhost', help="Redis server host")
    parser.add_argument('-p', '--port', type=int, default=6379, help="Redis server port")
    parser.add_argument('-a', '--password', help="Redis server password")
    
    parser.add_argument('--num-blocked', type=int, default=2, help="Number of blocked clients to create.")
    parser.add_argument('--num-txn', type=int, default=2, help="Number of in-transaction clients to create.")
    parser.add_argument('--num-pubsub', type=int, default=5, help="Number of Pub/Sub clients to create.")
    parser.add_argument('--num-idle', type=int, default=10, help="Number of idle/long-lived clients to create.")
    parser.add_argument('--num-large-obuf', type=int, default=1, help="Number of clients with a large output buffer.")
    parser.add_argument('--enable-monitor', action='store_true', help="Enable a single MONITOR client. WARNING: DANGEROUS FOR PRODUCTION!")

    args = parser.parse_args()

    def create_connection():
        # For this script, creating a new connection per thread is fine and avoids any potential sharing issues.
        return redis.Redis(host=args.host, port=args.port, password=args.password, decode_responses=True)

    simulations = {
        'blocked': (simulate_blocked_client, args.num_blocked),
        'transaction': (simulate_transaction_client, args.num_txn),
        'pubsub': (simulate_pubsub_client, args.num_pubsub),
        'idle': (simulate_idle_client, args.num_idle),
        'large_obuf': (simulate_large_obuf_client, args.num_large_obuf)
    }
    
    if args.enable_monitor:
        simulations['monitor'] = (simulate_monitor_client, 1)

    print("--- Starting Redis Client Simulator ---")
    print(f"Connecting to {args.host}:{args.port}")
    
    client_id_counter = 0
    for sim_type, (target_func, count) in simulations.items():
        if count > 0:
            print(f"Starting {count} '{sim_type}' client(s)...")
            for _ in range(count):
                client_id_counter += 1
                try:
                    r_conn = create_connection()
                    # Test connection before starting thread
                    r_conn.ping() 
                    thread = threading.Thread(target=target_func, args=(r_conn, client_id_counter))
                    thread.daemon = True
                    thread.start()
                    threads.append(thread)
                except redis.exceptions.RedisError as e:
                    print(f"ERROR: Could not connect to Redis to start a '{sim_type}' client. Aborting. Reason: {e}")
                    exit(1)


    print("\n--- Simulation running. Press Ctrl+C to stop. ---")
    print("Use 'redis-cli CLIENT LIST' in another terminal to observe the clients.")

    try:
        while True:
            time.sleep(1)
    except KeyboardInterrupt:
        print("\n--- Shutting down simulator ---")
        stop_event.set()
        time.sleep(2)
        print("Simulator stopped.")
