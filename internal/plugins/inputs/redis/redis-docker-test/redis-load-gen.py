import redis
import threading
import time
import random
import string
import uuid
import json

from rediscluster import RedisCluster
from rediscluster import RedisClusterException


# --- 配置 ---
STARTUP_NODES = [
    {"host": "192.168.139.162", "port": "6379"},
]

# --- 负载配置 ---
# 主力工作线程，模拟常规读写
REGULAR_WORKERS = 64
# 特效工作线程
SLOW_CMD_WORKERS = 6  # 制造慢查询
BIG_TRAFFIC_WORKERS = 6  # 制造大流量
EVICTION_WORKERS = 6  # 制造内存淘汰
KEY_MISS_WORKERS = 6  # 制造 Key Miss

TOTAL_DURATION_SECONDS = 72000 # 脚本总运行时间（秒）

# --- 统计与控制 ---
stats = {
    'writes': 0, 'reads_hit': 0, 'reads_miss': 0, 'slow_commands': 0,
    'big_writes_bytes': 0, 'eviction_writes': 0, 'errors': 0,
}
lock = threading.Lock()
stop_event = threading.Event()

# --- 连接 Redis Cluster ---
try:
    print("正在连接到 Redis Cluster...")
    rc = RedisCluster(startup_nodes=STARTUP_NODES, decode_responses=True)
    rc.ping()
    print("连接成功！")
except RedisClusterException as e:
    print(f"连接失败: {e}")
    exit(1)

def generate_random_string(length=10):
    return ''.join(random.choice(string.ascii_letters + string.digits) for _ in range(length))

# --- 特效工作函数 ---

def slow_command_worker():
    """执行长耗时命令，如 KEYS 或对大 ZSET 操作"""
    print("慢查询工作线程已启动...")
    # 先准备一个大的 ZSET
    zset_key = "{slow_test}:my_sorted_set"
    if not rc.exists(zset_key):
        print("正在准备慢查询所需的大 ZSET...")
        pipeline = rc.pipeline()
        for i in range(20000): # 准备 2 万个成员
            pipeline.zadd(zset_key, {f"member:{i}": random.randint(0, 100000)})
        pipeline.execute()
        print("大 ZSET 准备完毕。")

    while not stop_event.is_set():
        try:
            # 随机选择一种慢查询执行
            if random.random() < 0.5:
                # 执行 ZREVRANGE，这在数据量大时会比较慢
                rc.zrevrange(zset_key, 0, 10000, withscores=True)
            else:
                # KEYS 是阻塞的，非常危险，但很适合模拟慢查询
                # 使用 scan_iter 是更优的方式，但这里为了模拟效果用 KEYS
                rc.keys("user.*")

            with lock:
                stats['slow_commands'] += 1
            time.sleep(random.uniform(5, 10)) # 每隔 5-10 秒执行一次慢查询
        except Exception as e:
            with lock: stats['errors'] += 1
            print(f"慢查询线程错误: {e}")

def big_traffic_worker():
    """制造大量数据同步的命令"""
    print("大流量工作线程已启动...")
    while not stop_event.is_set():
        try:
            # 随机选择一种大流量操作
            if random.random() < 0.5:
                # 方式一: MSET 批量写入
                keys_values = {f"batch:{{big_traffic}}:{generate_random_string(10)}": str(uuid.uuid4()) for _ in range(500)}
                rc.mset(keys_values)
                bytes_sent = sum(len(k) + len(v) for k, v in keys_values.items())
            else:
                # 方式二: 写入一个大的 Value
                key = f"big_value:{{big_traffic}}:{generate_random_string(10)}"
                big_value = json.dumps([generate_random_string(100)] * 50) # 约 5KB
                rc.set(key, big_value, ex=30)
                bytes_sent = len(key) + len(big_value)

            with lock:
                stats['big_writes_bytes'] += bytes_sent
            time.sleep(random.uniform(0.5, 2)) # 每隔 0.5-2 秒制造一次大流量
        except Exception as e:
            with lock: stats['errors'] += 1
            print(f"大流量线程错误: {e}")

def eviction_worker():
    """持续写入数据以触发内存淘汰"""
    print("内存淘汰工作线程已启动...")
    while not stop_event.is_set():
        try:
            # 写入永不过期的 key，持续填充内存
            key = f"evict_me.{generate_random_string(15)}"
            rc.set(key, "a" * random.randint(0, 1024)) # 写入 1KB 以内的数据
            with lock:
                stats['eviction_writes'] += 1
            time.sleep(0.01) # 快速写入
        except redis.exceptions.ResponseError as e:
            if "OOM command not allowed" in str(e):
                print("触发内存上限 (OOM)，等待... (这是预期行为)")
                time.sleep(5)
            else:
                with lock: stats['errors'] += 1
        except Exception as e:
            with lock: stats['errors'] += 1
            print(f"内存淘汰线程错误: {e}")

def key_miss_worker():
    """高频请求不存在的 key"""
    print("Key Miss 工作线程已启动...")
    while not stop_event.is_set():
        try:
            # 请求一个几乎不可能存在的 key
            miss_key = f"miss:{uuid.uuid4()}"
            rc.get(miss_key)
            with lock:
                stats['reads_miss'] += 1 # 明确这是 miss
            time.sleep(random.uniform(0.01, 0.05))
        except Exception as e:
            with lock: stats['errors'] += 1
            print(f"Key Miss 线程错误: {e}")

def regular_worker():
    """常规读写工作线程"""
    while not stop_event.is_set():
        try:
            # 写入带 TTL 的常规数据
            key = f"user.{generate_random_string(8)}"
            value = str(uuid.uuid4())
            rc.set(key, value, ex=60)
            with lock:
                stats['writes'] += 1

            # 读取一个可能存在的 key
            if rc.get(key):
                with lock:
                    stats['reads_hit'] += 1
            time.sleep(random.uniform(0.02, 0.1))
        except Exception as e:
            with lock: stats['errors'] += 1

# --- 主程序 ---
if __name__ == "__main__":
    worker_pool = {
        "常规读写": (regular_worker, REGULAR_WORKERS),
        "慢查询": (slow_command_worker, SLOW_CMD_WORKERS),
        "大流量": (big_traffic_worker, BIG_TRAFFIC_WORKERS),
        "内存淘汰": (eviction_worker, EVICTION_WORKERS),
        "Key Miss": (key_miss_worker, KEY_MISS_WORKERS),
    }
    
    threads = []
    for name, (target_func, count) in worker_pool.items():
        if count > 0:
            print(f"启动 {count} 个 '{name}' 工作线程...")
            for i in range(count):
                thread = threading.Thread(target=target_func)
                thread.start()
                threads.append(thread)

    start_time = time.time()
    last_report_time = start_time

    try:
        while time.time() - start_time < TOTAL_DURATION_SECONDS:
            time.sleep(10) # 每 10 秒打印一次报告
            current_time = time.time()
            elapsed = current_time - last_report_time
            
            with lock:
                total_reads = stats['reads_hit'] + stats['reads_miss']
                hit_rate = (stats['reads_hit'] / total_reads * 100) if total_reads > 0 else 0
                
                print("\n" + "="*20 + " 负载状态报告 " + "="*20)
                print(f"时间: {time.strftime('%H:%M:%S')} | 报告周期: {elapsed:.2f}s")
                print(f"  常规写入: {stats['writes'] / elapsed:.2f} QPS")
                print(f"  读取命中/未命中: {stats['reads_hit'] / elapsed:.2f} / {stats['reads_miss'] / elapsed:.2f} QPS | 命中率: {hit_rate:.2f}%")
                print(f"  慢查询执行次数: {stats['slow_commands']} 次")
                print(f"  大流量写入: {(stats['big_writes_bytes'] / 1024) / elapsed:.2f} KB/s")
                print(f"  内存填充写入: {stats['eviction_writes'] / elapsed:.2f} QPS")
                print(f"  错误数: {stats['errors']}")
                
                # 重置计数器
                stats = {k: 0 for k in stats}
            
            last_report_time = current_time

    except KeyboardInterrupt:
        print("\n收到停止信号...")
    finally:
        print("正在停止所有工作线程...")
        stop_event.set()
        for t in threads:
            t.join(timeout=2)
        print("所有线程已停止。脚本退出。")
