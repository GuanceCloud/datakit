# diskcache

diskcache 是一种类似 wal 的磁盘缓存，它有如下特性：

- 支持并行读写
- 支持分片大小控制
- 支持单条数据大小控制
- 支持磁盘大小控制（FIFO）

限制：

- 不支持随机读取，只支持按照 FIFO 的顺序来消费数据

## 实现算法

```
Always put data to this file.
 |
 |   
 v   
data 
 |    Rotate: if `data` full, move to tail[data.0000000(n+1)]
 `-----------------------------------------------------------.
                                                             |
| data.00000000 | data.00000001 | data.00000002 | ....  <----`
      ^
      `----------------- Always read from this file(the file with smallest number)
```

- 当前正在写入的文件 `data` 不会实时消费，如果最近没有写入（3s），读取操作会将 `data` rotate 一下并消费
- 数据从 `data.00000001` 处开始消费（`Get`），如果队列上没有可消费的数据，`Get` 操作将返回 `ErrEOF`
- `data` 写满之后，将会在队列尾部追加一个新的文件，并重新创建 `data` 写入

## 使用

以下是基本的使用方式：

```golang
import "github.com/GuanceCloud/diskcache"

// Create new cache under /some/path
c, err := diskcache.Open(WithPath("/some/path"))

// Create new cache under /some/path, set batch size to 4MB
c, err := diskcache.Open(WithPath("/some/path"), WithBatchSize(4*1024*1024))

// Create new cache under /some/path, set cache capacity to 1GB
c, err := diskcache.Open(WithPath("/some/path"), WithCapacity(1024*1024*1024))

if err != nil {
	log.Printf(err)
	return
}

// Put data to
data := []byte("my app data...")
if err := c.Put(data); err != nil {
	log.Printf(err)
	return
}

if err := c.Get(func(x []byte) error {
	// Do something with the cached data...
	return nil
	}); err != nil {
	log.Printf(err)
	return
}

// get cache metrics
m := c.Metrics()
log.Println(m.LineProto()) // get line-protocol format of metrics
```

这种方式可以直接以并行的方式来使用，调用方无需针对这里的 diskcache 对象 `c` 做互斥处理。

## 通过 ENV 控制缓存 option

支持通过如下环境变量来覆盖默认的缓存配置：

| 环境变量                           | 描述                                                                                        |
| ---                                | ---                                                                                         |
| ENV_DISKCACHE_BATCH_SIZE           | 设置单个磁盘文件大小，单位字节，默认 64MB                                                   |
| ENV_DISKCACHE_MAX_DATA_SIZE        | 限制单次写入的字节大小，避免意料之外的巨量数据写入，单位字节，默认不限制                    |
| ENV_DISKCACHE_CAPACITY             | 限制缓存能使用的磁盘上限，一旦用量超过该限制，老数据将被移除掉。默认不限制                  |
| ENV_DISKCACHE_NO_SYNC              | 禁用磁盘写入的 sync 同步，默认不开启。一旦开启，可能导致磁盘数据丢失问题                    |
| ENV_DISKCACHE_NO_LOCK              | 禁用文件目录夹锁。默认是加锁状态，一旦不加锁，在同一个目录多开（`Open`）可能导致文件混乱    |
| ENV_DISKCACHE_NO_POS               | 禁用磁盘写入位置记录，默认带有位置记录。一旦不记录，程序重启会导致部分数据重复消费（`Get`） |
| ENV_DISKCACHE_NO_FALLBACK_ON_ERROR | 禁用错误回退机制                                                                            |


## Prometheus 指标

所有指标均有如下 label：

| label                | 取值               | 说明                                                          |
| ---                  | ---                | ---                                                           |
| no_fallback_on_error | true/false         | 是否关闭错误回退（即禁止 Get() 回调失败时，再次读到老的数据） |
| no_lock              | true/false         | 是否关闭加锁功能（即允许一个 cache 目录同时被多次 `Open()`）  |
| no_pos               | true/false         | 是否关闭 pos 功能                                             |
| no_sync              | true/false         | 是否关闭同步写入功能                                          |
| path                 | cache 所在磁盘目录 | cache 所在磁盘目录                                            |

指标列表如下：

| 指标                          | 类型    | 说明                                                           |
| ---                           | ---     | ---                                                            |
| diskcache_batch_size          | gauge   | HELP diskcache_batch_size data file size(in bytes)             |
| diskcache_capacity            | gauge   | current capacity(in bytes)                                     |
| diskcache_datafiles           | gauge   | current un-readed data files                                   |
| diskcache_dropped_bytes_total | counter | dropped bytes during Put() when capacity reached               |
| diskcache_dropped_total       | counter | dropped files during Put() when capacity reached               |
| diskcache_get_bytes_total     | counter | cache Get() bytes count                                        |
| diskcache_get_latency_sum     | summary | Get() time cost(micro-second)                                  |
| diskcache_get_latency_count   | summary | Get() time cost(micro-second)                                  |
| diskcache_get_total           | counter | cache Get() count                                              |
| diskcache_max_data            | gauge   | max data to Put(in bytes), default 0                           |
| diskcache_open_time           | gauge   | current cache Open time in unix timestamp(second)              |
| diskcache_put_bytes_total     | counter | cache Put() bytes count                                        |
| diskcache_put_latency_sum     | summary | Put() time cost(micro-second)                                  |
| diskcache_put_latency_count   | summary | Put() time cost(micro-second)                                  |
| diskcache_put_total           | counter | cache Put() count                                              |
| diskcache_rotate_total        | counter | cache rotate count, mean file rotate from data to data.0000xxx |
| diskcache_wakeup_total        | counter | total wakeup count                                             |
| diskcache_size                | gauge   | current cache size(in bytes)                                   |

## 性能估算

测试环境：

- Model Name            : MacBook Pro
- Model Identifier      : MacBookPro18,1
- Chip                  : Apple M1 Pro
- Total Number of Cores : 10 (8 performance and 2 efficiency)
- Memory                : 16 GB

> 详见测试用例 `TestConcurrentPutGetPerf`。

单次写入的数据量在 100KB ~ 1MB 之间，分别测试单线程写入、多线程写入、多线程读写情况下的性能：

| 测试情况   | worker | 性能（字节/毫秒） |
| ---        | ---    | ---               |
| 单线程写入 | 1      | 119708 bytes/ms   |
| 多线程写入 | 10     | 118920 bytes/ms   |
| 多线程读写 | 10+10  | 118920 bytes/ms   |

综合下来，不管多线程读写还是单线程读写，其 IO 性能在当前的硬件上能达到 100MB/s 的速度。

## TODO

- [ ] 支持一次 `Get()/Put()` 多个数据，提高加锁的数据吞吐量
- [ ] 支持 `Get()` 出错时重试机制（`WithErrorRetry(n)`）
- [ ] 可执行程序（*cmd/diskcache*）支持查看已有（可能正在被其它进程占用）diskcache 的存储情况
