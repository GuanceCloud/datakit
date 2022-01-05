# pipeline worker 使用说明

## pl-worker task 相关接口、结构体、函数

```go
/*
Source:
    单个pipeline worker 任务, Source 字段不可为空，
    否则将导致构造的行协议的 measurement name 为空； 

ScriptName:
    pl 脚本的名，可以是任意字符串，对于 pl-worker 预加载的 pl 脚本为文件名:
      1. 对于存在于 datakit 安装目录下的 pipeline 目录
        或由 git 管理的指定目录下的 pl 脚本，
        pl-worker/ gitrepo 将会调用函数 LoadAllDotPScriptForWkr 进行加载，
        此时仅需填写如 "nginx.p"， "mysql.p" 等;
      2. 对于未被预加载，但是通过调用 ScriptRegister 注册的 pl 脚本，
         需要填写注册时使用的 pl 脚本的名
      3. 此参数为空时，自动根据 Source + ".p" 进行寻找

TS:
    pl 任务创建时间
*/
type Task struct {
	TaskName   string // 任务名，任意填写
	ScriptName string // 为空则根据 source 匹配对应的脚本
	Source     string // or measurement name
	Data       []TaskData
	Opt        *TaskOpt // 可选
	TS         time.Time // （pl worker）task 创建时间
}

// 发送 task 到 pl-worker，非阻塞，即 taskCh 满返回 error(ErrTaskBusy)
func FeedPipelineTask(task *Task) error

// 阻塞方式发送 task 到 pl-worker
func FeedPipelineTaskBlock(task *Task) error



type TaskData interface {
    // worker 将会调用此方法获取待处理的文本
	GetContent() string
    // 回调，用于对 pl 处理结果进行操作 
	Handler(*Result) error
}

// pl 执行结果
type Result struct {
	output *parser.Output
}

// 获取处理结果中的 tag，不存在则返回 error
func (r *Result) GetTag(k string) (string, error)

// 添加或覆盖 tag
func (r *Result) SetTag(k, v string)

// 删除 tag
func (r *Result) DeleteTag(k string)

func (r *Result) GetField(k string) (interface{}, error)

func (r *Result) SetField(k string, v interface{})

func (r *Result) DeleteField(k string)

// 标记为待丢弃的数据，不上报该条日志
func (r *Result) Drop()


type TaskOpt struct {
	Category string // 默认 datakit.Logging

	Version string  // io.Feed 参数 io.Option 中的 Version 字段

	// 忽略这些status，如果数据的status在此列表中，数据将不再上传
	// ex: "info"
	//     "debug"
	IgnoreStatus []string

	// 是否关闭添加默认status字段列，包括status字段的固定转换行为，
    // 例如'd'->'debug'
	DisableAddStatusField bool
}


```

## pl-worker script 相关变量、函数

```go
/*
    pl worker 中的 scriptCentorStore，用于存储所有 pl 脚本，
    其存储结构示例为：

    map[string]map[string]*ScriptInfo {
        // 默认 ns
        "default": {
            // 脚本的名,
            "nginx.p": &ScriptInfo {
                name: "nginx.p",
                // 脚本内容
                script: "grok(_, "%{DATA: data}")\n...",
                // pl script 更新时间
                updateTS: time.Now(),
            },
            "Registered Name": &ScriptInfo{
                name: "Registered Name",            
                script: "grok(_, "%{DATA: data}")\n...",
                updateTS: time.Now(),
            }
        },
    }   
*/
var scriptCentorStore = &dotPScriptStore{
	scripts: map[string]map[string]*ScriptInfo{
		DefaultScriptNs: {},
	},
}


/*

    加载脚本到 scriptCentorStore 的 "default" 对应的 map 下；
    
    写入此 map 的 key 为文件名，如 nginx.p !!!

    对于同 basename 的文件，内容相同则忽略；

    后加载（读取）的会覆盖之前的，加载顺序为：

    1. 安装目录的 pipeline 当前目录下的扩展名为 `.p` 的文件
    |
    V
    2. filePathList， 文件路径列表
    |
    V
    3. userDefPath, 文件夹路径列表

* 
    此函数应当仅用于 gitrepo 的重载操作和
        plWorker 初始化时的 pl 脚本预加载

*/
func LoadAllDotPScriptForWkr(userDefPath []string, 
        gitRepoPPFile []string)

/*
    ScriptRegister 用于往 scriptCentorStore[DefaultScriptNs] 添加 pl 脚本，
    无法覆盖已经注册（含预加载）的脚本，诸如 nginx.p;
    若与已注册脚本的内容的不一致则返回 error (ErrScriptExists);
    脚本解析失败也将返回 error.

param:
    name: 可以为任意字符串，
    script: 脚本文件内容
*/
func ScriptRegister(name, script string) error 
```
