import redis
import threading
import multiprocessing as mp
import time
import random
import string
import uuid
import json

from rediscluster import RedisCluster
from rediscluster import RedisClusterException


# --- 配置 ---
STARTUP_NODES = [
    {"host": "192.168.139.162", "port": 7001},
    {"host": "192.168.139.162", "port": 7002}, 
    {"host": "192.168.139.162", "port": 7003},
]

# Redis 密码
REDIS_PASSWORD = "abc123456"

# --- 负载配置（新增分片写入/热点读取） ---
# 三个分片标签，使用哈希标签确保落到不同分片
SHARD_TAGS = ["{s1}", "{s2}", "{s3}"]
# 每个分片的字符串 key 数量在此范围内随机，约 2 万左右
PER_SHARD_STRING_KEYS_RANGE = (18_000, 22_000)
KEY_TTL_SECONDS = 2 * 60 * 60  # 2 小时

# 大结构配置（每个分片各建5个）
BIG_LISTS_PER_SHARD = 5
BIG_ZSETS_PER_SHARD = 5
BIG_ELEMS_PER_STRUCT = 10_000

# 热点 key 配置（仅读）
HOT_KEYS_PER_SHARD = 100
HOT_READ_PROCESSES = 6
HOT_READ_QPS_PER_PROC = 500  # 每进程目标读速率（近似）

# 保留原有线程型 workload（默认关闭）
REGULAR_WORKERS = 0
SLOW_CMD_WORKERS = 0
BIG_TRAFFIC_WORKERS = 0
EVICTION_WORKERS = 0
KEY_MISS_WORKERS = 0
BIG_ZSET_WORKERS = 0

TOTAL_DURATION_SECONDS = 72000  # 脚本总运行时间（秒）

# --- 统计与控制 ---
stats = {
    'writes': 0, 'reads_hit': 0, 'reads_miss': 0, 'slow_commands': 0,
    'big_writes_bytes': 0, 'eviction_writes': 0, 'errors': 0,
    'big_zset_ops': 0, 'redirects': 0
}
lock = threading.Lock()
stop_event = threading.Event()
proc_stop = mp.Event()

# --- 连接 Redis Cluster ---
try:
    print("正在连接到 Redis Cluster...")
    rc = RedisCluster(
        startup_nodes=STARTUP_NODES, 
        password=REDIS_PASSWORD,
        decode_responses=True,
        skip_full_coverage_check=True,
        health_check_interval=30,
        readonly_mode=True  # 允许从 slave 节点读取
    )
    rc.ping()
    print("连接成功！")
    print(f"集群信息: {rc.cluster_info()}")
except RedisClusterException as e:
    print(f"连接失败: {e}")
    exit(1)


# --- 构建辅助 ---
def make_rc():
    return RedisCluster(
        startup_nodes=STARTUP_NODES,
        password=REDIS_PASSWORD,
        decode_responses=True,
        skip_full_coverage_check=True,
        health_check_interval=30,
        readonly_mode=True  # 允许从 slave 节点读取
    )


def batched(iterable, n):
    batch = []
    for x in iterable:
        batch.append(x)
        if len(batch) >= n:
            yield batch
            batch = []
    if batch:
        yield batch


def populate_strings_for_shard(rc_obj, tag, total):
    # 创建普通字符串 key，设置 2 小时过期
    created = 0
    for chunk in batched(range(total), 1000):
        p = rc_obj.pipeline()
        for i in chunk:
            k = f"{tag}kv:{i}"
            v = str(uuid.uuid4())
            p.set(k, v, ex=KEY_TTL_SECONDS)
        p.execute()
        created += len(chunk)
    return created


def populate_hot_keys_for_shard(rc_obj, tag, hot_n):
    hot_keys = []
    for chunk in batched(range(hot_n), 200):
        p = rc_obj.pipeline()
        for i in chunk:
            k = f"{tag}hot:{i}"
            v = str(uuid.uuid4())
            p.set(k, v, ex=KEY_TTL_SECONDS)
            hot_keys.append(k)
        p.execute()
    return hot_keys


def populate_big_list_for_shard(rc_obj, tag, idx, elems):
    key = f"{tag}biglist:{idx}"
    # 按 1000 元素分批 LPUSH，确保在同一分片（依赖 tag）
    for start in range(0, elems, 1000):
        end = min(start + 1000, elems)
        items = [f"v{j}:{generate_random_string(8)}" for j in range(start, end)]
        rc_obj.lpush(key, *items)
    rc_obj.expire(key, KEY_TTL_SECONDS)
    return key


def populate_big_zset_for_shard(rc_obj, tag, idx, elems):
    key = f"{tag}bigzset:{idx}"
    # 按 1000 元素分批 ZADD
    for start in range(0, elems, 1000):
        end = min(start + 1000, elems)
        mapping = {f"m{j}:{generate_random_string(6)}": random.randint(0, 1_000_000) for j in range(start, end)}
        rc_obj.zadd(key, mapping)
    rc_obj.expire(key, KEY_TTL_SECONDS)
    return key


def populate_shard(rc_obj, tag, str_total):
    print(f"[POPULATE] 分片 {tag} 开始写入... 目标 strings={str_total}")
    n1 = populate_strings_for_shard(rc_obj, tag, str_total)
    hot_keys = populate_hot_keys_for_shard(rc_obj, tag, HOT_KEYS_PER_SHARD)

    big_lists = []
    for i in range(BIG_LISTS_PER_SHARD):
        big_lists.append(populate_big_list_for_shard(rc_obj, tag, i, BIG_ELEMS_PER_STRUCT))

    big_zsets = []
    for i in range(BIG_ZSETS_PER_SHARD):
        big_zsets.append(populate_big_zset_for_shard(rc_obj, tag, i, BIG_ELEMS_PER_STRUCT))

    print(f"[POPULATE] 分片 {tag} 完成: strings={n1}, big_lists={len(big_lists)}, big_zsets={len(big_zsets)}")
    return hot_keys


def hot_reader_proc(hot_keys, stop_evt, qps):
    # 直接用普通 Redis 客户端连接到单个 slave 节点（不使用 Cluster 模式）
    # 每个进程随机选择一个 slave 节点
    slave_nodes = [
        {"host": "192.168.139.162", "port": 7004},
        {"host": "192.168.139.162", "port": 7005},
        {"host": "192.168.139.162", "port": 7006},
        {"host": "192.168.139.162", "port": 7007},
        {"host": "192.168.139.162", "port": 7008},
        {"host": "192.168.139.162", "port": 7009},
    ]
    
    rnd = random.Random()
    slave_node = rnd.choice(slave_nodes)
    
    # 使用普通 Redis 客户端直接连接 slave
    rc_local = redis.Redis(
        host=slave_node["host"],
        port=slave_node["port"],
        password=REDIS_PASSWORD,
        decode_responses=True
    )
    
    # 发送 READONLY 命令，允许从 slave 读取（Cluster 模式必须）
    try:
        rc_local.execute_command("READONLY")
    except Exception as e:
        print(f"READONLY command failed: {e}")
    
    period = 1.0 / max(1, qps)
    try:
        while not stop_evt.is_set():
            key = rnd.choice(hot_keys)
            try:
                rc_local.get(key)
            except Exception:
                pass
            time.sleep(period)
    finally:
        try:
            rc_local.close()
        except Exception:
            pass

def generate_random_string(length=10):
    return ''.join(random.choice(string.ascii_letters + string.digits) for _ in range(length))

# --- 工具：检查并打印大 key ---
def check_and_print_big_key(key, len_threshold=5000, mem_threshold=10 * 1024 * 1024):
    try:
        ktype = rc.type(key)
        vlen = None
        if ktype == 'string':
            vlen = rc.strlen(key)
        elif ktype == 'list':
            vlen = rc.llen(key)
        elif ktype == 'hash':
            vlen = rc.hlen(key)
        elif ktype == 'set':
            vlen = rc.scard(key)
        elif ktype == 'zset':
            vlen = rc.zcard(key)

        vmem = None
        try:
            # Redis 4+ 支持 MEMORY USAGE；某些部署可能无权限
            vmem = rc.execute_command('MEMORY', 'USAGE', key, 'SAMPLES', 100)
        except Exception:
            pass

        hit_len = (vlen is not None and vlen >= len_threshold)
        hit_mem = (vmem is not None and vmem >= mem_threshold)
        if hit_len or hit_mem:
            print(f"[BIGKEY] key={key} type={ktype} len={vlen} mem={vmem}")
            return True
    except Exception as e:
        print(f"检查大key出错 key={key}: {e}")
    return False

# --- 特效工作函数 ---

def slow_command_worker():
    """执行长耗时命令，如 KEYS 或对大 ZSET 操作"""
    print("慢查询工作线程已启动...")
    # 先准备一个大的 ZSET
    zset_key = "{slow_test}:my_sorted_set"
    print(f"[KEY] 慢查询固定大ZSET键名: {zset_key}")
    try:
        if not rc.exists(zset_key):
            print("正在准备慢查询所需的大 ZSET...")
            pipeline = rc.pipeline()
            for i in range(10000):  # 准备 1 万个成员
                pipeline.zadd(zset_key, {f"member:{i}": random.randint(0, 100000)})
            pipeline.execute()
            print("大 ZSET 准备完毕。")
            # 打印是否为大 key
            check_and_print_big_key(zset_key)
    except Exception as e:
        print(f"准备大 ZSET 时出错: {e}")

    while not stop_event.is_set():
        try:
            # 随机选择一种慢查询执行
            if random.random() < 0.5:
                # 执行 ZREVRANGE，这在数据量大时会比较慢
                rc.zrevrange(zset_key, 0, 5000, withscores=True)
            else:
                # 使用 SCAN 替代 KEYS，更安全
                keys = []
                for key in rc.scan_iter(match="user.*", count=100):
                    keys.append(key)
                    if len(keys) >= 100:  # 限制扫描数量
                        break

            with lock:
                stats['slow_commands'] += 1
            time.sleep(random.uniform(5, 10))  # 每隔 5-10 秒执行一次慢查询
        except Exception as e:
            with lock: 
                stats['errors'] += 1
            print(f"慢查询线程错误: {e}")

def big_traffic_worker():
    """制造大量数据同步的命令"""
    print("大流量工作线程已启动...")
    while not stop_event.is_set():
        try:
            # 随机选择一种大流量操作
            if random.random() < 0.5:
                # 方式一: MSET 批量写入
                keys_values = {f"batch:{{big_traffic}}:{generate_random_string(10)}": str(uuid.uuid4()) for _ in range(200)}
                rc.mset(keys_values)
                bytes_sent = sum(len(k) + len(v) for k, v in keys_values.items())
            else:
                # 方式二: 写入一个大的 Value
                key = f"big_value:{{big_traffic}}:{generate_random_string(10)}"
                big_value = json.dumps([generate_random_string(100)] * 30)  # 约 3KB
                rc.set(key, big_value, ex=30)
                bytes_sent = len(key) + len(big_value)

            with lock:
                stats['big_writes_bytes'] += bytes_sent
            time.sleep(random.uniform(0.5, 2))  # 每隔 0.5-2 秒制造一次大流量
        except Exception as e:
            with lock: 
                stats['errors'] += 1
            print(f"大流量线程错误: {e}")

def eviction_worker():
    """持续写入数据以触发内存淘汰"""
    print("内存淘汰工作线程已启动...")
    while not stop_event.is_set():
        try:
            # 写入永不过期的 key，持续填充内存
            key = f"evict_me.{generate_random_string(15)}"
            rc.set(key, "a" * random.randint(100, 1024))  # 写入 100B-1KB 的数据
            with lock:
                stats['eviction_writes'] += 1
            time.sleep(0.1)  # 快速写入
        except redis.exceptions.ResponseError as e:
            if "OOM command not allowed" in str(e):
                print("触发内存上限 (OOM)，等待... (这是预期行为)")
                time.sleep(5)
            else:
                with lock: 
                    stats['errors'] += 1
        except Exception as e:
            with lock: 
                stats['errors'] += 1
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
                stats['reads_miss'] += 1  # 明确这是 miss
            time.sleep(random.uniform(0.01, 0.05))
        except Exception as e:
            with lock: 
                stats['errors'] += 1
            print(f"Key Miss 线程错误: {e}")

def big_zset_worker():
    """制造大 ZSET 操作，用于测试大键扫描"""
    print("大 ZSET 工作线程已启动...")
    zset_key = f"big_zset:{{big_key}}:{generate_random_string(8)}"
    print(f"[KEY] 动态大ZSET键名: {zset_key}")
    
    while not stop_event.is_set():
        try:
            # 创建大 ZSET
            pipeline = rc.pipeline()
            for i in range(1000):  # 每次添加 1000 个成员
                pipeline.zadd(zset_key, {f"member_{i}_{generate_random_string(5)}": random.randint(0, 100000)})
            pipeline.execute()
            
            # 执行一些 ZSET 操作
            rc.zcard(zset_key)
            rc.zrange(zset_key, 0, 100, withscores=True)
            # 打印是否为大 key
            check_and_print_big_key(zset_key)
            
            with lock:
                stats['big_zset_ops'] += 1
            
            # 设置过期时间，避免内存无限增长
            rc.expire(zset_key, 300)  # 5分钟后过期
            
            time.sleep(random.uniform(2, 5))
        except Exception as e:
            with lock: 
                stats['errors'] += 1
            print(f"大 ZSET 线程错误: {e}")

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
            with lock: 
                stats['errors'] += 1

# --- 主程序 ---
if __name__ == "__main__":
    # 1) 预写每个分片
    agg_hot_keys = []
    for tag in SHARD_TAGS:
        str_total = random.randint(*PER_SHARD_STRING_KEYS_RANGE)
        agg_hot_keys.extend(populate_shard(rc, tag, str_total))

    # 2) 启动热点只读进程
    unique_hot_keys = list(agg_hot_keys)
    procs = []
    if unique_hot_keys:
        print(f"启动 {HOT_READ_PROCESSES} 个热点只读进程，共 {len(unique_hot_keys)} 个热key")
        for _ in range(HOT_READ_PROCESSES):
            p = mp.Process(target=hot_reader_proc, args=(unique_hot_keys, proc_stop, HOT_READ_QPS_PER_PROC))
            p.start()
            procs.append(p)

    # 3) 可选：保留旧的线程型负载（默认关闭）
    worker_pool = {
        "常规读写": (regular_worker, REGULAR_WORKERS),
        "慢查询": (slow_command_worker, SLOW_CMD_WORKERS),
        "大流量": (big_traffic_worker, BIG_TRAFFIC_WORKERS),
        "内存淘汰": (eviction_worker, EVICTION_WORKERS),
        "Key Miss": (key_miss_worker, KEY_MISS_WORKERS),
        "大 ZSET": (big_zset_worker, BIG_ZSET_WORKERS),
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
    try:
        while time.time() - start_time < TOTAL_DURATION_SECONDS:
            time.sleep(10)
    except KeyboardInterrupt:
        print("\n收到停止信号...")
    finally:
        print("正在停止所有进程/线程...")
        proc_stop.set()
        for p in procs:
            p.join(timeout=3)
        stop_event.set()
        for t in threads:
            t.join(timeout=2)
        print("已停止。")
