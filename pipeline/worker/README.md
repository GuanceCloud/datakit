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

```
