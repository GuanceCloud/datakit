import redis
import threading
import time
import random
import string
import uuid
import json
import argparse

from rediscluster import RedisCluster
from rediscluster import RedisClusterException

# --- 配置 (将通过命令行参数覆盖) ---
DEFAULT_NODES = [
	# 设置 /etc/hosts，增加如下行：
	#   198.19.249.182  redis-cluster.local
    "redis-cluster.local:7001",
    "redis-cluster.local:7002",
    "redis-cluster.local:7003",
]

# --- 负载类型和数量配置 ---
# 你可以按需调整每个类型的线程数来改变负载模型
NUM_STRING_WORKERS = 5     # 常规读写
NUM_LIST_WORKERS = 3       # 模拟消息队列
NUM_HASH_WORKERS = 4       # 模拟对象存储
NUM_SET_WORKERS = 3        # 模拟标签/唯一项
NUM_ZSET_WORKERS = 2       # 模拟排行榜
NUM_STREAM_WORKERS = 2     # 模拟事件流

# --- 特效工作者 (可选) ---
NUM_EVICTION_WORKERS = 1   # 持续写入以触发内存淘汰
NUM_SLOW_CMD_WORKERS = 1   # 制造慢查询

TOTAL_DURATION_SECONDS = 7200 # 脚本总运行时间（秒）

LARGE_STRING_MAX_LEN = 1024

rc = None

# --- 统计与控制 ---
stats = {
    'string_writes': 0, 'string_reads': 0, 'list_pushes': 0, 'list_pops': 0,
    'hash_sets': 0, 'hash_gets': 0, 'set_adds': 0, 'set_reads': 0,
    'zset_adds': 0, 'zset_reads': 0, 'stream_adds': 0, 'stream_reads': 0,
    'stream_acks': 0, 'eviction_writes': 0, 'slow_commands': 0, 'errors': 0,
}
lock = threading.Lock()
stop_event = threading.Event()

def generate_random_string(length=10):
    return ''.join(random.choice(string.ascii_letters + string.digits) for _ in range(length))

# ===================================================================
#  专用数据结构工作函数
# ===================================================================

def string_worker():
    """常规 String 类型读写"""
    while not stop_event.is_set():
        try:
            key = f"user:session:{{str}}:{generate_random_string(10)}"
            rc.set(key, str(uuid.uuid4()), ex=120)
            with lock: stats['string_writes'] += 1
            rc.get(key)
            with lock: stats['string_reads'] += 1
            time.sleep(random.uniform(0.05, 0.2))
        except Exception:
            with lock: stats['errors'] += 1

def list_worker():
    """模拟消息队列 (LPUSH/BRPOP)"""
    queue_name = "messages:{{list_q}}" # 使用 hash tag 确保 key 在同个节点
    while not stop_event.is_set():
        try:
            # 生产者
            rc.lpush(queue_name, f"message:{uuid.uuid4()}")
            with lock: stats['list_pushes'] += 1
            
            # 消费者 (BRPOP 会产生 blocked_clients 指标)
            if random.random() < 0.7:
                if rc.brpop(queue_name, timeout=1):
                    with lock: stats['list_pops'] += 1
            time.sleep(random.uniform(0.1, 0.5))
        except Exception:
            with lock: stats['errors'] += 1

def hash_worker():
    """模拟用户配置对象 (HSET/HGETALL)"""
    while not stop_event.is_set():
        try:
            user_id = random.randint(1000, 2000)
            hash_key = f"user:profile:{{h}}:{user_id}"
            profile_data = {
                "last_login": time.time(),
                "email": f"user{user_id}@example.com",
                "theme": random.choice(["dark", "light"]),
            }
            rc.hset(hash_key, mapping=profile_data)
            with lock: stats['hash_sets'] += 1

            if random.random() < 0.5:
                rc.hgetall(hash_key)
                with lock: stats['hash_gets'] += 1
            time.sleep(random.uniform(0.2, 0.8))
        except Exception:
            with lock: stats['errors'] += 1

def set_worker():
    """模拟文章标签 (SADD/SISMEMBER)"""
    while not stop_event.is_set():
        try:
            article_id = random.randint(1, 100)
            set_key = f"article:tags:{{set}}:{article_id}"
            tag = random.choice(["tech", "life", "go", "python", "docker", "redis"])
            rc.sadd(set_key, tag)
            with lock: stats['set_adds'] += 1

            rc.sismember(set_key, random.choice(["tech", "life", "go"]))
            with lock: stats['set_reads'] += 1
            time.sleep(random.uniform(0.1, 0.4))
        except Exception:
            with lock: stats['errors'] += 1

def zset_worker():
    """模拟游戏排行榜 (ZADD/ZREVRANGE)"""
    leaderboard_key = "leaderboard:global:{{zset}}"
    while not stop_event.is_set():
        try:
            player_id = f"player:{random.randint(1, 500)}"
            score = random.randint(1, 100000)
            rc.zadd(leaderboard_key, {player_id: score})
            with lock: stats['zset_adds'] += 1

            if random.random() < 0.2:
                rc.zrevrange(leaderboard_key, 0, 9) # 获取 Top 10
                with lock: stats['zset_reads'] += 1
            time.sleep(random.uniform(0.1, 0.6))
        except Exception:
            with lock: stats['errors'] += 1

def stream_worker():
    """模拟事件流 (XADD/XREADGROUP)"""
    stream_key = "iot:events:{{stream}}"
    group_name = "my_group"
    consumer_name = f"consumer-{threading.get_ident()}"

    # 尝试创建消费者组，如果已存在会报错，忽略即可
    try:
        rc.xgroup_create(stream_key, group_name, id='0', mkstream=True)
    except redis.exceptions.ResponseError as e:
        if "BUSYGROUP" not in str(e):
            raise e

    while not stop_event.is_set():
        try:
            # 生产者
            event_data = {
                "sensor_id": f"sensor-{random.randint(1, 10)}",
                "temp": f"{random.uniform(15.0, 30.0):.2f}",
            }
            rc.xadd(stream_key, event_data)
            with lock: stats['stream_adds'] += 1

            # 消费者
            if random.random() < 0.6:
                response = rc.xreadgroup(group_name, consumer_name, {stream_key: '>'}, count=5, block=2000)
                if response:
                    with lock: stats['stream_reads'] += len(response[0][1])
                    # 确认消息
                    message_ids = [msg[0] for msg in response[0][1]]
                    rc.xack(stream_key, group_name, *message_ids)
                    with lock: stats['stream_acks'] += len(message_ids)
            time.sleep(random.uniform(0.5, 1.5))
        except Exception:
            with lock: stats['errors'] += 1

# ===================================================================
#  特效工作函数 (从 v2 脚本保留)
# ===================================================================
def eviction_worker():
    # ... (代码与 v2 版本相同)
    while not stop_event.is_set():
        try:
            key = f"evict_me:{generate_random_string(15)}"
            rc.set(key, "a" * random.randint(0, LARGE_STRING_MAX_LEN)) 
            with lock: stats['eviction_writes'] += 1
            time.sleep(0.01) 
        except redis.exceptions.ResponseError as e:
            if "OOM command not allowed" in str(e): time.sleep(5)
        except Exception:
            with lock: stats['errors'] += 1

def slow_command_worker():
    # ... (代码与 v2 版本相同)
    zset_key = "{slow_test}:my_sorted_set"
    if not rc.exists(zset_key):
        pipe = rc.pipeline()
        for i in range(20000): pipe.zadd(zset_key, {f"member:{i}": i})
        pipe.execute()
    while not stop_event.is_set():
        try:
            rc.zrevrange(zset_key, 0, 10000)
            with lock: stats['slow_commands'] += 1
            time.sleep(random.uniform(5, 10))
        except Exception:
            with lock: stats['errors'] += 1

# ===================================================================
#  主程序
# ===================================================================
if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Redis Cluster 负载生成器 (支持读写分离)")
    parser.add_argument('--nodes', nargs='+', default=DEFAULT_NODES, help=f"集群的启动节点列表，用空格分隔 (默认: {' '.join(DEFAULT_NODES)})")
    parser.add_argument('-a', '--password', help="集群密码")
    parser.add_argument('--duration', type=int, default=TOTAL_DURATION_SECONDS, help=f"脚本总运行时间 (秒) (默认: {TOTAL_DURATION_SECONDS})")

    args = parser.parse_args()
    TOTAL_DURATION_SECONDS = args.duration
    nodes = [{"host": host.split(':')[0], "port": int(host.split(':')[1])} for host in args.nodes]
    
    try:
        print("正在连接到 Redis Cluster 并启用读写分离...")
        rc = RedisCluster(
            startup_nodes=nodes,
            skip_full_coverage_check=True,
            decode_responses=True,
            password=args.password,
            # --- 核心配置：启用只读路由 ---
            #readonly_mode=True,
            # 路由只读命令到从库。可以选择 'latency' (延迟最低) 或 'random'
            #readonly_routing_table='latency'
        )
        rc.ping()
        print("连接成功！")
    except Exception as e:
        print(f"连接 Redis Cluster 失败: {e}, {nodes}, {args.password}")
        exit(1)

    worker_pool = {
        "String": (string_worker, NUM_STRING_WORKERS),
        "List": (list_worker, NUM_LIST_WORKERS),
        "Hash": (hash_worker, NUM_HASH_WORKERS),
        "Set": (set_worker, NUM_SET_WORKERS),
        "ZSet": (zset_worker, NUM_ZSET_WORKERS),
        "Stream": (stream_worker, NUM_STREAM_WORKERS),
        "Eviction": (eviction_worker, NUM_EVICTION_WORKERS),
        "SlowCmd": (slow_command_worker, NUM_SLOW_CMD_WORKERS),
    }
    
    threads = []
    print("--- 启动工作线程 ---")
    for name, (target_func, count) in worker_pool.items():
        if count > 0:
            print(f"  -> 启动 {count} 个 '{name}' 工作线程...")
            for i in range(count):
                thread = threading.Thread(target=target_func, daemon=True)
                thread.start()
                threads.append(thread)

    start_time = time.time()
    last_report_time = start_time

    try:
        while time.time() - start_time < TOTAL_DURATION_SECONDS:
            time.sleep(10)
            current_time = time.time()
            elapsed = current_time - last_report_time
            if elapsed == 0: continue
            
            with lock:
                print("\n" + "="*25 + " 负载状态报告 " + "="*25)
                print(f"时间: {time.strftime('%H:%M:%S')} | 周期: {elapsed:.2f}s | 总错误数: {stats['errors']}")
                print(f"  String (W/R):      {stats['string_writes']/elapsed:.1f} / {stats['string_reads']/elapsed:.1f} QPS")
                print(f"  List (Push/Pop):   {stats['list_pushes']/elapsed:.1f} / {stats['list_pops']/elapsed:.1f} QPS")
                print(f"  Hash (Set/Get):    {stats['hash_sets']/elapsed:.1f} / {stats['hash_gets']/elapsed:.1f} QPS")
                print(f"  Set (Add/Read):    {stats['set_adds']/elapsed:.1f} / {stats['set_reads']/elapsed:.1f} QPS")
                print(f"  ZSet (Add/Read):   {stats['zset_adds']/elapsed:.1f} / {stats['zset_reads']/elapsed:.1f} QPS")
                print(f"  Stream (Add/Read): {stats['stream_adds']/elapsed:.1f} / {stats['stream_reads']/elapsed:.1f} QPS (Acks: {stats['stream_acks']/elapsed:.1f})")
                if NUM_EVICTION_WORKERS > 0 or NUM_SLOW_CMD_WORKERS > 0:
                    print("-" * 70)
                    print(f"  特效 - Eviction Writes: {stats['eviction_writes']/elapsed:.1f} QPS | Slow Commands: {stats['slow_commands']} 次")
                
                # 重置计数器
                for k in stats: stats[k] = 0
            
            last_report_time = current_time

    except KeyboardInterrupt:
        print("\n收到停止信号...")
    finally:
        print("正在停止所有工作线程...")
        stop_event.set()
        # 等待线程退出
        main_thread = threading.current_thread()
        for t in threading.enumerate():
            if t is main_thread:
                continue
            t.join(timeout=2)
        print("所有线程已停止。脚本退出。")
