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
      `----------------- Always read from this file
```

- 当前正在写入的文件 `data` 不会实时消费，如果最近没有写入（3s），读取操作会将 `data` rotate 一下并消费
- 数据从 `data.00000001` 处开始消费（`Get`），如果队列上没有可消费的数据，`Get` 操作将返回 `ErrEOF`
- `data` 写满之后，将会在队列尾部追加一个新的文件，并重新创建 `data` 写入

## 使用

以下是基本的使用方式：

```golang
import "github.com/GuanceCloud/diskcache"

c, err := diskcache.Open(".TestDropBatch", &Option{
	BatchSize: 4 * 1024 * 1024, // each cache file set to ~4MB
	Capacity:  32 * 1024 * 1024, // max disk usge limited to ~32MB
	}}
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
```

这种方式可以直接以并行的方式来使用，调用方无需针对这里的 diskcache 对象 `c` 做互斥处理。

## 通过 ENV 控制缓存 option

支持通过如下环境变量来覆盖默认的缓存配置：

| 环境变量                    | 描述                                                                       |
| ---                         | ---                                                                        |
| ENV_DISKCACHE_BATCH_SIZE    | 设置单个磁盘文件大小，单位字节，默认 64MB                                  |
| ENV_DISKCACHE_MAX_DATA_SIZE | 限制单次写入的字节大小，避免意料之外的巨量数据写入，单位字节，默认不限制   |
| ENV_DISKCACHE_CAPACITY      | 限制缓存能使用的磁盘上限，一旦用量超过该限制，老数据将被移除掉。默认不限制 |
| ENV_DISKCACHE_NO_ASYNC      | 禁用磁盘写入的 async 同步，默认不开启。一旦开启，可能导致磁盘数据丢失问题  |
