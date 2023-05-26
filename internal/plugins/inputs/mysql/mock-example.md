# Mock 编写示例

为便于 datakit 自测功能（self-contained/reliable），需大面积 mock 各种数据，包括但不限于：

- 采集器数据
- 各个功能模块验证数据
- 中心接口数据

现有几乎所有采集器，均缺少 mock 数据，这会导致一个严重问题：用户环境千奇百怪，我们无法预测，即使同一个版本的软件，因配置不同，其采集到的数据都可能有差异，如：

- MySQL 8.0 跟 MySQL 5.7 某个采集的数据不同，后续测试中，难道分别搭建两个 MySQL 来测试么？
- Ngxin 版本很多，用户做了不同的配置，导致 Nginx 指标行为有差异，我们的采集器能否快速复现并支持？后续对 Ngxin 采集器做调整的时候，如何保证多个版本的 Nginx 采集不会受到影响？
- 硬件数据就更不必说了，不是所有硬件都能长时间保持随时可用的状态，开发还可以，时间久了，可能找都找不到了。

以上这些问题，都可以通过将「被采集数据 mock 一份」来解决，避免重复搭建测试环境，能极大提升软件重构、升级的效率和信心。

Mock 的做法，参考了[这篇文章的做法](https://www.myhatchpad.com/insight/mocking-techniques-for-go/)，其中也有考虑现有的一些 [Mock 框架](https://github.com/golang/mock)。综合下来，前者更适合 DataKit 的场景，后者过于庞杂，不易控制。

现已 MySQL 采集器一个小模块的 mock 编写为例，简单叙述一下 mock 的一种做法。

## binlog 空间指标采集

在 MySQL 中，有一个统计 binlog 所占磁盘空间的指标，其实现方式很简单，通过 SQL 查询即可得到各个 binlog 的大小：

```sql
SHOW BINARY LOGS;

 Log_name        | File_size
-----------------|-----------
mysql-bin.000001 |  177
mysql-bin.000002 |  177
mysql-bin.000003 |  154
mysql-bin.000004 |  154
mysql-bin.000005 |  154
...
```

将 `File_size` 各个累加，即可得到 binlog 所占磁盘大小，实现非常简单，此处不再赘述。

为便于引入 Mock 机制，我稍微重构了一下 MySQL 采集器，主要涉及：

- 封装所有 MySQL 查询操作：在现有 MySQL 采集器中，只有查询操作，故封装查询即可。
- 将数据提取操作，抽象出一层 interface 出来，便于 mock
- 重写现有 binlog 指标采集函数

### 封装查询操作

通过定义如下函数，封装下所有 MySQL 查询：

```golang
func (i *Input) q(s string) rows { }
```

注意，此处没有错误返回，如遇错误，返回 `nil` 即可，或者往 IO 模块报告 last error。

### 抽象数据提取

在 `rows.go` 中定义了一个接口：

```golang
type rows interface {}
```

上面封装的 `q()` 函数，返回的就是这个 `rows` 接口。注意，代码里面的接口列表，就是标准库 `sql.Rows` 实现的一组函数，这里我们只抽取其中几个即可，具体参见代码。

### 重写 binlog 采集函数

重写原有 binlog 采集函数，此处不讨论之前写法问题。重写后的函数为（具体参见代码）：

```golang
func binlogMetrics(r rows) (map[string]interface{}, error) {}
```

注意，这里并未将 `binlogMetrics` 定义成 `Input` 成员函数，主要便于对其施加单元测试，在采集器的 `Run()/Collect()` 函数中，这样调用它：

```golang
// 此处 i 即 input 实例
if res, err := binlogMetrics(i.q("SHOW BINARY LOGS;")); err != nil {
	return err
} else {
	i.binlog = res
}
```

## 施加单元测试

做好了上述准备，下面开始写 mock 测试，这里的方法论，只要参考的是[这篇文章的做法](https://www.myhatchpad.com/insight/mocking-techniques-for-go/)，更精确一点，就是用接口替换（Interface Substitution）技术。

### 构造 MySQL 返回数据

```golang
type mockRows struct {
	data [][]interface{} // 里面放 MySQL mock 数据
	pos int              // 数据提取游标
	closed bool          // rows 是否关闭
	...
}

// ----------------------
// 实现一组接口(具体参见代码)
// ----------------------
func (r *mockRows) Scan(args ...interface{}) error {}
func (r *mockRows) Next() bool {}
func (r *mockRows) Close() error {}
func (r *mockRows) Err() error {}
```

### 编写测试用例

`binlogMetrics()` 测试用例如下：

```golang
func TestBinlogMetrics(t *testing.T) {
	cases := []struct {
		name   string
		rows   *mockRows
		err    error
		expect int64
	}{
		{
			name: "basic",
			rows: &mockRows{
				t: t,
				data: [][]interface{}{
					[]interface{}{"mysql-bin.000001", "123"},
					[]interface{}{"mysql-bin.000002", "456"},
					[]interface{}{"mysql-bin.000003", "789"},
				},
			},
			expect: int64(123 + 456 + 789),
			err:    nil,
		},

		{
			name: "error-bin-log-size",
			rows: &mockRows{
				t: t,
				data: [][]interface{}{
					[]interface{}{"mysql-bin.000001", "123"},
					[]interface{}{"mysql-bin.000002", "456"},
					[]interface{}{"mysql-bin.000003", "abc123"}, // ignored
				},
			},
			expect: int64(123 + 456),
			err:    nil,
		},

		{
			name: "no-bin-log",
			rows: &mockRows{
				t:    t,
				data: [][]interface{}{},
			},
			expect: int64(0),
			err:    nil,
		},
		{
			name: "invalid-bin-log-size",
			rows: &mockRows{
				t: t,
				data: [][]interface{}{
					[]interface{}{"mysql-bin.000001", "-1"},   // ignored
					[]interface{}{"mysql-bin.000002", "3.14"}, // ignored
					[]interface{}{"mysql-bin.000003",
						fmt.Sprintf("%d", uint64(math.MaxInt64)+1)}, // ignored
				},
			},
			expect: int64(0),
			err:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := binlogMetrics(tc.rows)
			if tc.err != nil {
				tu.Assert(t, err != nil, "should failed with error %s", tc.err)
			} else {
				tu.Assert(t, err == nil, "expect no err, but got %s", err)
			}

			tu.Equals(t, tc.expect, res["Binlog_space_usage_bytes"].(int64))
		})
	}
}
```

运行以上测试函数：

```shell
go test -test.v -timeout 1h -run TestBinlogMetrics

=== RUN   TestBinlogMetrics
=== RUN   TestBinlogMetrics/basic
=== RUN   TestBinlogMetrics/error-bin-log-size
=== RUN   TestBinlogMetrics/no-bin-log
=== RUN   TestBinlogMetrics/invalid-bin-log-size
--- PASS: TestBinlogMetrics (0.00s)
    --- PASS: TestBinlogMetrics/basic (0.00s)
    --- PASS: TestBinlogMetrics/error-bin-log-size (0.00s)
    --- PASS: TestBinlogMetrics/no-bin-log (0.00s)
    --- PASS: TestBinlogMetrics/invalid-bin-log-size (0.00s)
PASS
ok  	gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/mysql	0.277s
```

## 总结

综上所属，我们不难发现一些 mock 的基本方法和技巧：

- 对关键数据路径进行数据上的 mock 和处理：比如 MySQL 中的 mock 就以 MySQL 数据查询作为 mock 入手点，即抽象出 rows 数据处理接口，使得 `sql.Rows` 也能满足其接口约束接口。

因为程序真实要处理的数据对象是 `sql.Rows`，但测试时，我们无法自行构造 sql.Rows，为了让相关处理代码能走完测试流程，可自行构造一类接口，使得该接口能 cover 住 `sql.Rows` 对象上我们用到的那些方法即可。这里要注意一点，即 sql.Rows 加入它实现了 8 个方法，但在我们 MySQL 采集器中只用到了 4 个的话，那么我们只需要 mock 出一个含有改 4 个方法的 interface{} 即可，无需 8 个方法全部 mock 进这个 interface 中。

- 对 mock 出来的对象，我们可以从真实环境获取数据，使得其跟真实对象表现如一：在 `metric_test.go` 中，我们构造了一些测试用例，其中一些数据就是从真实的 MySQL 中查找出来的，这对于测试正常的流程是不可或缺的。当然，也需要构造一些异常数据，而这些异常数据，才是 mock 真实的意图所在。

我们无法保证用户环境的数据一定按照常理出牌，但一旦有异常数据，那么我们至少能通过一些方式拿到这些异常数据（跑一个 select 语句就能拿到数据），将其在 mock 中复现并校验我们的程序是否处理得当。否则到用户的环境调试、打日志，很不雅观。

- 覆盖率测试会、缺陷测试、性能测试：在真实的环境中，我们可能无法发现程序的内存泄露以及性能问题，毕竟真实的 MySQL 采集默认是 10s 一次，如果是较为隐匿的问题（边界问题、正常流程不可能执行的代码分支问题、缓慢泄露问题等），正常开发测试可能无法发现。但是在 mock 测试中，性能测试、[内存泄露测试](https://www.freecodecamp.org/news/how-i-investigated-memory-leaks-in-go-using-pprof-on-a-large-codebase-4bec4325e192/)就变得可行了。

关于覆盖率测试，可以做一个 alias 命令，他会输出一个 html 文件，通过浏览器即可打开这个文件，以查看哪些代码覆盖到来了，哪些没有覆盖到：

```shell
# alias for `go test cover show`
alias gtcovershow='go test -v -cover . -coverprofile=coverage.out && go tool cover -html=coverage.out'
```

上面已经将比较难以 Mock 的情况作了示例。其它一些复杂的采集器（如 MySQL/CPU/Disk 等）、模块（如选举模块、Git 模块、IO 模块等），均可采用这种 Interface Substitution 技术来实现（现有的一些 mock 框架，其实现原理也类似），另外，[参考文章中](https://www.myhatchpad.com/insight/mocking-techniques-for-go/)也提及了其它几种 mock 技术，也可采纳。
