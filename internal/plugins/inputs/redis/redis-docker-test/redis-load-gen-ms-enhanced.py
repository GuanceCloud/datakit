# redis_load_generator_ms_enhanced.py
# 主从模式负载生成器 - 增强版（包含大key和热key）
#
# 特性：
# 1. 预热固定数量的 key，运行时不再创建新 key（默认3000个普通key）
# 2. 读写分离：90% 读，10% 写（进一步减少写入）
# 3. 写操作只更新已有 key，使用 KEEPTTL 保持原有过期时间
# 4. 包含热key（100个，频繁访问，无TTL）和大key（10个，list/hash/zset）
# 5. 普通key 50% 设置TTL（10-30分钟），50% 永久保存
# 6. 模拟缓存穿透（20% 访问不存在的key，但不保存）
# 7. Key 总数稳定，不会无限增长，方便 scan 操作
#
import redis
import threading
import time
import random
import string
import uuid
import json
import argparse

# --- Redis 配置 ---
REDIS_MASTER_HOST = '192.168.139.162'
REDIS_MASTER_PORT = 6379
REDIS_MASTER_PASSWORD = 'abc123456'

REDIS_SLAVE_1_HOST = '192.168.139.162'
REDIS_SLAVE_1_PORT = 6380
REDIS_SLAVE_1_PASSWORD = 'abc123456'

REDIS_SLAVE_2_HOST = '192.168.139.162'
REDIS_SLAVE_2_PORT = 6381
REDIS_SLAVE_2_PASSWORD = 'abc123456'

# --- 负载配置 ---
NUM_THREADS = 16                    # 工作线程数（减少）
TOTAL_DURATION_SECONDS = 3600       # 运行时长
DB = 0
READ_WRITE_RATIO = 0.9              # 读写比例：90% 读，10% 写（进一步减少写入）

# --- 热key配置 ---
HOT_KEYS_COUNT = 100                # 热key数量
HOT_KEY_READ_PROBABILITY = 0.3      # 30% 的读操作访问热key

# --- 大key配置 ---
BIG_KEYS_COUNT = 10                 # 大key数量
BIG_LIST_ELEMENTS = 5000            # 大list的元素数量
BIG_HASH_FIELDS = 3000              # 大hash的字段数量
BIG_ZSET_MEMBERS = 4000             # 大zset的成员数量

# --- 普通key配置 ---
NORMAL_KEYS_POOL_SIZE = 3000        # 普通key池大小（减少到3000）
MISS_KEY_PROBABILITY = 0.2          # 20% 概率访问不存在的key（模拟缓存穿透）

# --- TTL配置 ---
TTL_PROBABILITY = 0.5               # 50% 的普通key设置TTL
TTL_RANGE = (600, 1800)             # TTL范围：10分钟到30分钟（缩短）

# --- 统计与控制 ---
stats = {
    'writes_master': 0,
    'reads_slave_hit': 0,
    'reads_slave_miss': 0,
    'hot_key_reads': 0,
    'big_key_ops': 0,
    'big_writes_bytes': 0,
    'errors': 0
}

lock = threading.Lock()
stop_event = threading.Event()

rc_master = None
rc_slave1 = None
rc_slave2 = None
threads = []
hot_keys = []       # 存储热key列表
big_keys = []       # 存储大key列表
normal_keys = []    # 存储一些普通key用于读取

def generate_random_string(length=10):
    return ''.join(random.choice(string.ascii_letters + string.digits) for _ in range(length))

def populate_hot_keys():
    """初始化热key - 这些key会被频繁访问"""
    print(f"[初始化] 正在生成 {HOT_KEYS_COUNT} 个热key...")
    keys = []
    pipeline = rc_master.pipeline()
    
    for i in range(HOT_KEYS_COUNT):
        key = f"hot:user:{i}"
        value = json.dumps({
            'user_id': i,
            'name': generate_random_string(10),
            'email': f'user{i}@example.com',
            'created_at': time.time()
        })
        # 热key不设置过期时间，保证一直存在
        pipeline.set(key, value)
        keys.append(key)
        
        if (i + 1) % 100 == 0:
            pipeline.execute()
            pipeline = rc_master.pipeline()
    
    if len(pipeline) > 0:
        pipeline.execute()
    
    print(f"[初始化] 热key生成完成: {len(keys)} 个")
    return keys

def populate_big_keys():
    """初始化大key - 包含 list、hash、zset 类型"""
    print(f"[初始化] 正在生成 {BIG_KEYS_COUNT} 个大key...")
    keys = []
    
    # 生成大 LIST
    for i in range(BIG_KEYS_COUNT // 3):
        key = f"biglist:log:{i}"
        print(f"  生成大LIST: {key} ({BIG_LIST_ELEMENTS} 个元素)")
        pipeline = rc_master.pipeline()
        
        for j in range(BIG_LIST_ELEMENTS):
            log_entry = json.dumps({
                'timestamp': time.time(),
                'level': random.choice(['INFO', 'WARN', 'ERROR']),
                'message': generate_random_string(100)
            })
            pipeline.lpush(key, log_entry)
            
            if (j + 1) % 1000 == 0:
                pipeline.execute()
                pipeline = rc_master.pipeline()
        
        if len(pipeline) > 0:
            pipeline.execute()
        
        # 设置较长的过期时间
        rc_master.expire(key, 7200)
        keys.append(key)
    
    # 生成大 HASH
    for i in range(BIG_KEYS_COUNT // 3):
        key = f"bighash:session:{i}"
        print(f"  生成大HASH: {key} ({BIG_HASH_FIELDS} 个字段)")
        pipeline = rc_master.pipeline()
        
        for j in range(BIG_HASH_FIELDS):
            field = f"field_{j}"
            value = generate_random_string(50)
            pipeline.hset(key, field, value)
            
            if (j + 1) % 1000 == 0:
                pipeline.execute()
                pipeline = rc_master.pipeline()
        
        if len(pipeline) > 0:
            pipeline.execute()
        
        rc_master.expire(key, 7200)
        keys.append(key)
    
    # 生成大 ZSET
    for i in range(BIG_KEYS_COUNT // 3 + 1):
        key = f"bigzset:leaderboard:{i}"
        print(f"  生成大ZSET: {key} ({BIG_ZSET_MEMBERS} 个成员)")
        pipeline = rc_master.pipeline()
        
        for j in range(BIG_ZSET_MEMBERS):
            member = f"player_{j}_{generate_random_string(8)}"
            score = random.uniform(0, 10000)
            pipeline.zadd(key, {member: score})
            
            if (j + 1) % 1000 == 0:
                pipeline.execute()
                pipeline = rc_master.pipeline()
        
        if len(pipeline) > 0:
            pipeline.execute()
        
        rc_master.expire(key, 7200)
        keys.append(key)
    
    print(f"[初始化] 大key生成完成: {len(keys)} 个")
    return keys

def populate_normal_keys():
    """初始化一批普通key用于读取和更新"""
    print(f"[初始化] 正在生成 {NORMAL_KEYS_POOL_SIZE} 个普通key（固定池大小）...")
    keys = []
    pipeline = rc_master.pipeline()
    
    for i in range(NORMAL_KEYS_POOL_SIZE):
        key = f"normal:data:{i}"
        value = str(uuid.uuid4())
        # 60% 设置TTL
        if random.random() < TTL_PROBABILITY:
            pipeline.set(key, value, ex=random.randint(*TTL_RANGE))
        else:
            pipeline.set(key, value)
        keys.append(key)
        
        if (i + 1) % 100 == 0:
            pipeline.execute()
            pipeline = rc_master.pipeline()
    
    if len(pipeline) > 0:
        pipeline.execute()
    
    print(f"[初始化] 普通key生成完成: {len(keys)} 个（这些key会被反复更新和读取）")
    return keys

def worker(worker_id):
    """
    工作线程：模拟各种读写操作
    策略：90% 读操作，10% 写操作；只更新已有key，不创建新key
    """
    global stats
    
    # 每个worker选择一个slave进行读取
    rc_slave = rc_slave1 if worker_id % 2 == 0 else rc_slave2
    
    while not stop_event.is_set():
        try:
            # === 根据读写比例决定操作类型 ===
            if random.random() < READ_WRITE_RATIO:
                # ===== 读取操作（90%）=====
                read_choice = random.random()
                
                if read_choice < HOT_KEY_READ_PROBABILITY and hot_keys:
                    # 30% 读取热key
                    read_key = random.choice(hot_keys)
                    with lock:
                        stats['hot_key_reads'] += 1
                elif read_choice < (HOT_KEY_READ_PROBABILITY + 0.5) and normal_keys:
                    # 50% 读取普通key池中的key
                    read_key = random.choice(normal_keys)
                else:
                    # 20% 读取不存在的key（模拟缓存穿透，不保存）
                    read_key = f"notexist:{generate_random_string(12)}"
                
                result = rc_slave.get(read_key)
                
                if result:
                    with lock:
                        stats['reads_slave_hit'] += 1
                else:
                    with lock:
                        stats['reads_slave_miss'] += 1
                
                # === 偶尔操作大key ===
                if random.random() < 0.02 and big_keys:  # 2% 概率
                    big_key = random.choice(big_keys)
                    key_type = big_key.split(':')[0]
                    
                    try:
                        if key_type == 'biglist':
                            rc_slave.lrange(big_key, 0, 10)  # 读取前10个元素
                        elif key_type == 'bighash':
                            rc_slave.hlen(big_key)  # 只获取长度，不读取全部
                        elif key_type == 'bigzset':
                            rc_slave.zrange(big_key, 0, 100, withscores=True)  # 读取top 100
                        
                        with lock:
                            stats['big_key_ops'] += 1
                    except:
                        pass  # 大key可能已过期
            
            else:
                # ===== 写入操作（10%）=====
                # 只更新已有key，不创建新key
                if not normal_keys:
                    time.sleep(0.01)
                    continue
                
                # 随机选择一个已有key进行更新
                key = random.choice(normal_keys)
                
                # 5% 概率写入大数据
                if random.random() < 0.05:
                    value = json.dumps([generate_random_string(100)] * 50)
                    with lock:
                        stats['big_writes_bytes'] += len(key) + len(value)
                else:
                    # 普通写入
                    value = json.dumps({
                        'id': str(uuid.uuid4()),
                        'timestamp': time.time(),
                        'worker': worker_id,
                        'counter': random.randint(1, 1000)
                    })
                
                # 更新key（保持原有的TTL设置）
                # 使用SET命令更新值，KEEPTTL保持原有过期时间
                try:
                    rc_master.set(key, value, keepttl=True)
                    with lock:
                        stats['writes_master'] += 1
                except:
                    # 如果KEEPTTL不支持，就用普通SET
                    rc_master.set(key, value)
                    with lock:
                        stats['writes_master'] += 1
            
            time.sleep(random.uniform(0.005, 0.02))  # 模拟思考时间
            
        except Exception as e:
            with lock:
                stats['errors'] += 1

def report_stats():
    """定期报告统计信息"""
    last_report_time = time.time()
    start_time = last_report_time
    
    while not stop_event.is_set():
        time.sleep(5)
        current_time = time.time()
        elapsed = current_time - last_report_time
        total_elapsed = current_time - start_time
        
        with lock:
            total_reads = stats['reads_slave_hit'] + stats['reads_slave_miss']
            hit_rate = (stats['reads_slave_hit'] / total_reads * 100) if total_reads > 0 else 0
            
            # 获取当前key数量
            try:
                dbsize = rc_master.dbsize()
            except:
                dbsize = "N/A"
            
            print("\n" + "="*60)
            print(f"负载状态报告 - 运行时间: {int(total_elapsed)}秒")
            print("="*60)
            print(f"【数据库状态】（Key数量应该保持稳定）")
            print(f"  当前 Key 总数: {dbsize}")
            print(f"  - 热 Key: {len(hot_keys)} 个（频繁访问，无TTL）")
            print(f"  - 大 Key: {len(big_keys)} 个（list/hash/zset）")
            print(f"  - 普通 Key 池: {len(normal_keys)} 个（固定池，反复更新）")
            print(f"\n【写入统计】（读写比 {int(READ_WRITE_RATIO*100)}:{int((1-READ_WRITE_RATIO)*100)}）")
            print(f"  Master 更新 QPS: {stats['writes_master'] / elapsed:.2f}")
            print(f"  大数据写入: {(stats['big_writes_bytes'] / 1024) / elapsed:.2f} KB/s")
            print(f"  ⚠️ 注意：只更新已有key，不创建新key")
            print(f"\n【读取统计】")
            print(f"  Slave 读取 QPS: {total_reads / elapsed:.2f}")
            print(f"  - 命中: {stats['reads_slave_hit']} ({hit_rate:.2f}%)")
            print(f"  - 未命中: {stats['reads_slave_miss']} （访问不存在的key，不保存）")
            print(f"  - 热key读取: {stats['hot_key_reads']}")
            print(f"\n【大key操作】")
            print(f"  大key操作次数: {stats['big_key_ops']}")
            print(f"\n【错误统计】")
            print(f"  错误数: {stats['errors']}")
            print("="*60)
            
            # 重置计数器
            stats['writes_master'] = 0
            stats['reads_slave_hit'] = 0
            stats['reads_slave_miss'] = 0
            stats['hot_key_reads'] = 0
            stats['big_key_ops'] = 0
            stats['big_writes_bytes'] = 0
        
        last_report_time = current_time

if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="Redis 主从模式负载生成器 - 增强版（包含大key和热key）",
        formatter_class=argparse.RawTextHelpFormatter
    )
    parser.add_argument('--db', type=int, default=DB, help="数据库编号")
    parser.add_argument('--threads', type=int, default=NUM_THREADS, help="工作线程数")
    parser.add_argument('--duration', type=int, default=TOTAL_DURATION_SECONDS, help="运行时长（秒）")
    parser.add_argument('--hot-keys', type=int, default=HOT_KEYS_COUNT, help="热key数量")
    parser.add_argument('--big-keys', type=int, default=BIG_KEYS_COUNT, help="大key数量")
    args = parser.parse_args()

    DB = args.db
    NUM_THREADS = args.threads
    TOTAL_DURATION_SECONDS = args.duration
    HOT_KEYS_COUNT = args.hot_keys
    BIG_KEYS_COUNT = args.big_keys

    # --- 连接 Redis ---
    try:
        print("="*60)
        print("正在连接到 Redis Master 和 Slaves...")
        print("="*60)
        
        rc_master = redis.Redis(
            host=REDIS_MASTER_HOST,
            port=REDIS_MASTER_PORT,
            db=DB,
            password=REDIS_MASTER_PASSWORD,
            decode_responses=True
        )
        
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
        
        print("✓ 连接成功！")
    except redis.exceptions.RedisError as e:
        print(f"✗ 连接失败: {e}")
        exit(1)
    
    # --- 初始化数据 ---
    print("\n" + "="*60)
    print("初始化测试数据...")
    print("="*60)
    print(f"⚠️ 提示：如果之前有大量旧数据，建议先清理：")
    print(f"   redis-cli -h {REDIS_MASTER_HOST} -p {REDIS_MASTER_PORT} -a {REDIS_MASTER_PASSWORD} FLUSHDB")
    print()
    
    # 检查当前key数量
    try:
        current_keys = rc_master.dbsize()
        print(f"当前数据库已有 {current_keys} 个key")
        if current_keys > 10000:
            print(f"⚠️ 警告：key数量较多，建议清理后再运行")
            response = input("是否继续？(y/n): ")
            if response.lower() != 'y':
                print("已取消运行")
                exit(0)
    except:
        pass
    
    print("\n开始预热数据...")
    hot_keys = populate_hot_keys()
    big_keys = populate_big_keys()
    normal_keys = populate_normal_keys()
    
    # 等待数据同步到slave
    print("\n等待3秒，让数据同步到 Slave...")
    time.sleep(3)
    
    # 显示最终的key数量
    try:
        final_keys = rc_master.dbsize()
        print(f"✓ 预热完成，当前共有 {final_keys} 个key")
        print(f"  预计稳定在 {final_keys} ± 几百个key（部分会过期）")
    except:
        pass
    
    # --- 启动工作线程 ---
    print("\n" + "="*60)
    print(f"启动 {NUM_THREADS} 个工作线程，持续 {TOTAL_DURATION_SECONDS} 秒...")
    print("="*60)
    
    # 启动统计报告线程
    report_thread = threading.Thread(target=report_stats)
    report_thread.daemon = True
    report_thread.start()
    
    # 启动工作线程
    for i in range(NUM_THREADS):
        thread = threading.Thread(target=worker, args=(i,))
        thread.daemon = True
        thread.start()
        threads.append(thread)
    
    start_time = time.time()
    
    try:
        while time.time() - start_time < TOTAL_DURATION_SECONDS:
            time.sleep(1)
    except KeyboardInterrupt:
        print("\n收到停止信号...")
    finally:
        print("\n正在停止所有工作线程...")
        stop_event.set()
        time.sleep(2)
        print("✓ 所有线程已停止。脚本退出。")

