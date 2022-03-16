# pipeline worker 使用说明

pipeline worker 模块提供 pipeline script 的存储和更新管理，以及数据的异步处理

可通过函数 `FeedPipelineTask(task *Task)` 和 `FeedPipelineTaskBlock(task *Task)` 发送 task 到创建的 pipeline worker 上, 后者发送 task 时将阻塞到 task channal 就绪

## `struct Task`

```go

type Task struct {
    TaskName string

    Source     string // measurement name
    ScriptName string // 为空则根据 source 匹配对应的脚本

    Opt *TaskOpt

    Data TaskData

    TS time.Time

    MaxMessageLen int

    // 保留字段
    Namespace string
}

type TaskData interface {
    ContentType() string // TaskDataString or TaskDataByte

    // 根据 content type 决定调用哪一方法
    
    GetContentStr() []string

    GetContentByte() [][]byte
    ContentEncode() string

    // 当 worker 对任务所有的 content 执行 pl script 结束时调用此
    // 需要自行 feed io； 
    // 可在实现接口时用 channal 取回 task ptr 和 []*pipeline.Result，
    // 但如有此需求可以考虑使用 new 一个 Pipeline 实例来执行 pl script 
    Callback(*Task, []*pipeline.Result) error
}

```
