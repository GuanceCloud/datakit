# pipeline worker 使用说明

pipeline worker 模块提供 pipeline script 的存储和更新管理，以及数据的异步处理

可通过函数 `FeedPipelineTask(task Task)` 和 `FeedPipelineTaskBlock(task Task)` 发送 task 到创建的 pipeline worker 上, 后者发送 task 时将阻塞到 task channal 就绪

## `interface Task`

```go

type Task interface {
    GetSource() string
    GetScriptName() string // 待调用的 pipeline 脚本
    GetMaxMessageLen() int

    ContentType() string // TaskDataString or TaskDataByte
    ContentEncode() string

    GetContent() interface{} // []string or [][]byte

    // feed 给 pipeline 时 pl worker 会调用此方法
    Callback([]*pipeline.Result) error
}

```
