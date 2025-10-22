# redis_load_generator_ms.py
import redis
import threading
import time
import random
import string
import uuid
import json
import argparse

# --- 配置 ---
# *** 这里是关键：我们将所有写操作指向 Master ***
REDIS_MASTER_HOST = '192.168.139.162'
REDIS_MASTER_PORT = 6379
REDIS_MASTER_PASSWORD = 'abc123456' # 替换成你的密码

# *** 我们将所有读操作指向 Slave ***
REDIS_SLAVE_1_HOST = '192.168.139.162'
REDIS_SLAVE_1_PORT = 6380
REDIS_SLAVE_1_PASSWORD = 'abc123456' # 替换成你的密码

REDIS_SLAVE_2_HOST = '192.168.139.162'
REDIS_SLAVE_2_PORT = 6381
REDIS_SLAVE_2_PASSWORD = 'abc123456' # 替换成你的密码

# --- 负载配置 ---
NUM_THREADS = 64          # 增加并发，以便产生可见的压力
TOTAL_DURATION_SECONDS = 3600 # 运行时长
DB = 0

# --- 统计与控制 ---
stats = {
    'writes_master': 0, 'reads_slave_hit': 0, 'reads_slave_miss': 0, 
    'big_writes_bytes': 0, 'errors': 0
}

lock = threading.Lock()
stop_event = threading.Event()

rc_master = None
rc_slave1 = None
rc_slave2 = None
threads = []

def generate_random_string(length=10):
    return ''.join(random.choice(string.ascii_letters + string.digits) for _ in range(length))

def worker(worker_id):
    """
    工作线程，模拟单个客户端的读写分离行为
    """
    print(f"工作线程 #{worker_id} 已启动...")
    global stats
    
    while not stop_event.is_set():
        try:
            # 1. 写入操作: 总是发往 Master
            key = f"user.{generate_random_string(12)}"
            value = str(uuid.uuid4())
            rc_master.set(key, value, ex=3600)  # 改为1小时过期
            with lock:
                stats['writes_master'] += 1

            # 2. 制造大流量: 偶尔发往 Master
            if random.random() < 0.05: # 5% 的几率
                big_key = f"big_log:{generate_random_string(10)}"
                big_value = json.dumps([generate_random_string(100)] * 50)
                rc_master.set(big_key, big_value, ex=3600)  # 改为1小时过期
                with lock:
                    stats['big_writes_bytes'] += len(big_key) + len(big_value)

            # 3. 读取操作: 总是发往 Slave
            # 50% 的几率读一个刚写入的 key (测试同步延迟)
            # 50% 的几率读一个随机 key (测试 miss)
            read_key = key if random.random() < 0.5 else f"user.{generate_random_string(12)}"
            
            # 从 Slave 读取数据
            result = rc_slave1.get(read_key)
            
            if result:
                with lock:
                    stats['reads_slave_hit'] += 1
            else:
                with lock:
                    stats['reads_slave_miss'] += 1

            result = rc_slave2.get(read_key)
            
            if result:
                with lock:
                    stats['reads_slave_hit'] += 1
            else:
                with lock:
                    stats['reads_slave_miss'] += 1

            time.sleep(random.uniform(0.01, 0.05)) # 模拟思考时间

        except Exception as e:
            with lock:
                stats['errors'] += 1
            # 在高延迟时，从 slave 读取可能因为数据还没到而出错，这里我们打印但不退出
            # print(f"线程错误: {e}")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Redis master/slave 负载生成器")
    parser.add_argument('--db', type=int, default=DB, help=f"要施压的数据库")
    args = parser.parse_args()

    DB = args.db

    # --- 连接 Redis ---
    try:
        print("正在连接到 Redis Master 和 Slave...")
        # 用于写入的 Master 客户端
        rc_master = redis.Redis(
            host=REDIS_MASTER_HOST,
            port=REDIS_MASTER_PORT, 
            db=DB,
            password=REDIS_MASTER_PASSWORD,
            decode_responses=True
        )
        # 用于读取的 Slave 客户端
        rc_slave1 = redis.Redis(
            host=REDIS_SLAVE_1_HOST,
            port=REDIS_SLAVE_1_PORT, 
            db=DB,
            password=REDIS_SLAVE_1_PASSWORD,
            decode_responses=True
        )

        rc_slave2 = redis.Redis(
            host=REDIS_SLAVE_2_HOST,
            port=REDIS_SLAVE_2_PORT, 
            db=DB,
            password=REDIS_SLAVE_2_PASSWORD,
            decode_responses=True
        )

        rc_master.ping()
        rc_slave1.ping()
        rc_slave2.ping()

        print("连接成功！")
    except redis.exceptions.RedisError as e:
        print(f"连接失败，请检查你的 Redis 地址和状态: {e}")
        exit(1)
    
    threads = []
    print(f"将在 {NUM_THREADS} 个线程中产生读写分离负载，持续 {TOTAL_DURATION_SECONDS} 秒...")

    for i in range(NUM_THREADS):
        thread = threading.Thread(target=worker, args=(i,))
        thread.start()
        threads.append(thread)

    start_time = time.time()
    last_report_time = start_time

    try:
        while time.time() - start_time < TOTAL_DURATION_SECONDS:
            time.sleep(5)
            current_time = time.time()
            elapsed = current_time - last_report_time
            
            with lock:
                total_reads = stats['reads_slave_hit'] + stats['reads_slave_miss']
                hit_rate = (stats['reads_slave_hit'] / total_reads * 100) if total_reads > 0 else 0
                
                print("\n----- 负载状态报告 -----")
                print(f"时间: {time.strftime('%H:%M:%S')}")
                print(f"Master 写入 QPS: {stats['writes_master'] / elapsed:.2f}")
                print(f"Slave 读取 QPS: {total_reads / elapsed:.2f}")
                print(f"Slave 命中率: {hit_rate:.2f}%")
                print(f"大流量写入: {(stats['big_writes_bytes'] / 1024) / elapsed:.2f} KB/s")
                print(f"错误数: {stats['errors']}")
                
                stats = {k: 0 for k in stats} # 重置计数器
            
            last_report_time = current_time

    except KeyboardInterrupt:
        print("\n收到停止信号...")
    finally:
        print("正在停止所有工作线程...")
        stop_event.set()
        for t in threads:
            t.join()
        print("所有线程已停止。脚本退出。")
