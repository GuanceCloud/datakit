# Mon Feb 10 11:03:17 UTC 2020

- 多 RP 写入/查询/订阅
- Flow 类型数据支持
- 修正 alert 数据中 duration 作为 tag 的问题
- python inner API 邀请链接失效优化

# Wed Jan 15 12:38:33 UTC 2020

- 对于符合行协议的数据，如果数据有问题（如类型冲突/DB 不存在/），kodo 会丢弃
- 重命名 `keyevent` 成 `$keyevent`
- 增加多 RP 写入支持
- 部分重构，清理掉部分垃圾代码
