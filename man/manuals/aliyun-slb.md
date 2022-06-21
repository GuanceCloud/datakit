# 阿里云 SLB
---

## 视图预览
阿里云 SLB 指标展示，包括后端 ECS 实例状态，端口连接数，QPS，网络流量，状态码等<br />![image.png](imgs/input-aliyun-slb-01.png)<br />
![image.png](imgs/input-aliyun-slb-02.png)

## 版本支持
操作系统支持：Linux  

## 前置条件

- 服务器 <[安装 Datakit](https://www.yuque.com/dataflux/datakit/datakit-install)>
- 服务器 <[安装 Func 携带版](https://www.yuque.com/dataflux/func/quick-start)>
- 阿里云 RAM 访问控制账号授权

### RAM 访问控制

1. 登录 RAM 控制台  [https://ram.console.aliyun.com/users](https://ram.console.aliyun.com/users)
1. 新建用户：人员管理 - 用户 - 创建用户

![image.png](imgs/input-aliyun-slb-03.png)

3. 保存或下载 **AccessKey** **ID** 和 **AccessKey Secret** 的 CSV 文件 (配置文件会用到)
3. 用户授权 (只读访问所有阿里云资源的权限)

![image.png](imgs/input-aliyun-slb-04.png)

## 安装配置
说明：

- 示例 Linux 版本为：CentOS Linux release 7.8.2003 (Core)
- 通过一台服务器采集所有阿里云 SLB 数据

### 部署实施

#### 脚本市场

1. 登录 Func，地址 http://ip:8088

![image.png](imgs/input-aliyun-slb-05.png)

2. 开启脚本市场，管理 - 实验性功能 - 开启脚本市场模块

![image.png](imgs/input-aliyun-slb-06.png)

3. **依次添加 **三个脚本集
   1. 观测云集成 (核心包)
   1. 观测云集成 (阿里云-云监控)
   1. 观测云集成 (阿里云-SLB)

_注：在安装「核心包」后，系统会提示安装第三方依赖包，按照正常步骤点击安装即可_<br />![image.png](imgs/input-aliyun-slb-07.png)

4. 脚本安装完成后，可以在脚本库中看到所有脚本集

![image.png](imgs/input-aliyun-slb-08.png)

#### 添加脚本

1. 开发 - 脚本库 - 添加脚本集

![image.png](imgs/input-aliyun-slb-09.png)

2. 点击该脚本集 - 添加脚本

![image.png](imgs/input-aliyun-slb-10.png)

3. 创建 ID 为 main 的脚本

![image.png](imgs/input-aliyun-slb-11.png)

4. 添加代码 (需要修改账号配置 **AccessKey ID/AccessKey Secret/Account Name**)
```
from guance_integration__runner import Runner        # 引入启动器
import guance_aliyun_slb__main as aliyun_slb         # 引入阿里云SLB采集器
import guance_aliyun_monitor__main as aliyun_monitor # 引入阿里云云监控采集器

# 账号配置
account = {
    'ak_id'     : 'AccessKey ID',
    'ak_secret' : 'AccessKey Secret',
    'extra_tags': {
        'account_name': 'Account Name',
    }
}

# 由于采集数据较多，此处需要为函数指定更大的超时时间（单位秒）
@DFF.API('执行云资产同步', timeout=300)
def run():
    # 采集器配置
    common_aliyun_configs = {
        'regions': [ 'cn-hangzhou' ], #阿里云SLB对应的地域
    }
    monitor_collector_configs = {
        'targets': [
            { 'namespace': 'acs_slb_dashboard', 'metrics': ['HeathyServerCount','UnhealthyServerCount','NewConnection','ActiveConnection','MaxConnection','Qps','StatusCodeOther','StatusCode2xx','StatusCode3xx','StatusCode4xx','StatusCode5xx','TrafficRXNew','TrafficTXNew','PacketTX','PacketRX','InstanceUpstreamCode5xx','InstanceUpstreamCode4xx','InstanceNewConnection','InstanceActiveConnection','InstanceMaxConnection','InstanceQps','InstanceStatusCodeOther','InstanceStatusCode2xx','InstanceStatusCode3xx','InstanceStatusCode4xx','InstanceStatusCode5xx','InstanceTrafficRX','InstanceTrafficTX','InstancePacketTX','InstancePacketRX'] }, # 采集云监控中SLB所有指标
        ],
    }

    # 创建采集器
    collectors = [
        aliyun_slb.DataCollector(account, common_aliyun_configs),
        aliyun_monitor.DataCollector(account, monitor_collector_configs),
    ]

    # 启动执行
    Runner(collectors).run()
```

5. **保存 **配置并 **发布**

![image.png](imgs/input-aliyun-slb-12.png)

#### 定时任务

1. 添加自动触发任务，管理 - 自动触发配置 - 新建任务

![image.png](imgs/input-aliyun-slb-13.png)

2. 自动触发配置，执行函数中添加此脚本，执行频率为 **每分钟 * * * * \***

![image.png](imgs/input-aliyun-slb-14.png)

3. 指标预览

![image.png](imgs/input-aliyun-slb-15.png)

## 场景视图
<场景 - 新建仪表板 - 内置模板库 - 阿里云 SLB>

## 监控规则
<监控 - 模板新建 - 阿里云 SLB 检测库>

## 指标详解
<[阿里云 SLB 指标列表](https://help.aliyun.com/document_detail/162852.htm?spm=a2c4g.11186623.0.0.68a46c8aQJ0fha#concept-2482356)>

## 常见问题排查

- 查看日志：Func 日志路径 /usr/local/dataflux-func/data/logs/dataflux-func.log
- 代码调试：编辑模式选择主函数，直接运行 (可以看到脚本输出)

![image.png](imgs/input-aliyun-slb-16.png)

- 连接配置：Func 无法连接 Datakit，请检查数据源配置 (Datakit 需要监听 0.0.0.0)

![image.png](imgs/input-aliyun-slb-17.png)

## 进一步阅读
<[DataFlux Func 观测云集成简介](https://www.yuque.com/dataflux/func/script-market-guance-integration-intro)><br /><[DataFlux Func 阿里云-云监控配置手册](https://www.yuque.com/dataflux/func/script-market-guance-aliyun-monitor)>
