# 函数禁用扫描工具

本工具主要用于扫描代码中一些比较危险的函数，这些函数可能会造成一些恶劣的影响，比如在高频代码中加入了 `fmt.Print{f,ln}` 函数，这些函数可能会打爆用户的系统日志。

## 配置示例

```toml
# 扫描的代码目录
src = ["internal/", ]

# 忽略特定的目录下的所有文件
skip_dirs = ["internal/colorprint", ]

# 忽略特定的代码文件
skip_files = []

# 禁用的 package 和对应的函数名列表
[pkg_functions]
	fmt = ["Println", "Printf", ]
```
