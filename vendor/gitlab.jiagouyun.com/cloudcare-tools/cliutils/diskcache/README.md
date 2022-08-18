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
