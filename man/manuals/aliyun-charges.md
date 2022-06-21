# 阿里云账单及费用
---

## 视图预览
![image.png](imgs/input-aliyun-charges-01.png)

## 安装部署

- 说明：<br />示例 Linux 版本为：CentOS Linux release 7.8.2003 (Core)<br />通过一台服务器采集所有阿里云账单费用数据


## 前置条件

- 服务器 <[安装 Datakit](https://www.yuque.com/dataflux/datakit/datakit-install)>
- 服务器 <[安装 Func 携带版](https://www.yuque.com/dataflux/func/quick-start)>
- 阿里云 RAM 访问控制账号授权


### RAM 访问控制

1. 登录 RAM 控制台 [https://ram.console.aliyun.com/users](https://ram.console.aliyun.com/users)
1. 新建用户：人员管理 - 用户 - 创建用户<br />![image.png](imgs/input-aliyun-charges-02.png)
1. 保存或下载 **AccessKeyID** 和 **AccessKey Secret** 的 CSV 文件 (配置文件会用到)
1. 用户授权（账单权限）

![](imgs/input-aliyun-charges-03.png)

## 配置实施

1. 登录 Func，地址 http://ip:8088（默认 admin/admin）

![](imgs/input-aliyun-charges-04.png)

2. 输入标题/描述信息

![image.png](imgs/input-aliyun-charges-05.png)

3. 编辑脚本
```python
import sys, datetime
import time
import json
import urllib
import hmac
from hashlib import sha1
import base64
import random
import requests


# 请求中所需要的公共参数（就是调用 API 都需要用到的参数）
D = {
    'BillingCycle':str(time.strftime("%Y-%m", time.gmtime())),
    'Action':'QuerySettleBill',
    # 'PageNum':'5',
    'Format':'JSON',
    'Version':'2017-12-14',
    'AccessKeyId':'<AccessKeyId>',
    'SignatureMethod':'HMAC-SHA1',
    'MaxResults' : '300',
    # 'NextToken':"", #?
    'Timestamp':str(time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())),
    'SignatureVersion':'1.0'
    # 'SignatureNonce':str_seed
}
#当前时间
now_time = str(time.strftime("%Y-%m-%d", time.gmtime()))

#链接本地 Datakit
datakit = DFF.SRC('datakit')

# 使用 python 中的 urllib 库来进行编码
def percentEncode(str):
        res = urllib.parse.quote(str.encode('utf8'), '')
        res = res.replace('+', '%20')
        res = res.replace('*', '%2A')
        res = res.replace('%7E', '~')
        return res


#获取账单
@DFF.API('getBill')
def getBill():
    # 账单当前记录位置
    next_token = ""
    # 循环获取账单写入 DataKit
    for i in range(10000):
        random.seed()
        # 唯一随机数，用于防止网络重放攻击。用户在不同请求间要使用不同的随机数值。
        D["SignatureNonce"] = str(random.random())
        D["NextToken"] = next_token
        # 由于签名要求唯一性，包括顺序，所以需要按照参数名称排序
        sortedD = sorted(D.items(),key=lambda x: x[0])
        canstring = ''
        for k,v in sortedD:
            canstring += '&' + percentEncode(k) + '=' + percentEncode(v)
        # 生成标准化请求字符串
        stringToSign = 'GET&%2F&' + percentEncode(canstring[1:])
        # access_key_secret
        access_key_secret = '<access_key_secret>'
        # 计算 HMAC 值
        h = hmac.new((access_key_secret + "&").encode('utf8'), stringToSign.encode('utf8'), sha1)
        # 计算签名值生成 signature 签名
        signature = base64.encodestring(h.digest()).strip()
        # 添加签名
        D['Signature'] = signature
        # 最终调用 API
        url = 'http://business.aliyuncs.com/?' + urllib.parse.urlencode(D)
        # 请求阿里云账单费用
        response = requests.get(url)
        billing_cycle = response.json()["Data"]["BillingCycle"]
        account_id = response.json()["Data"]["AccountID"]
        next_token = response.json()["Data"]["NextToken"]
        if next_token is not None:
            bill = response.json()["Data"]["Items"]["Item"]
            print(bill)
            # 写入当天账单到观测云
            for i in bill:
                print(i["UsageEndTime"])
                time = i["UsageEndTime"].split(" ")[0]
                print(time,now_time)
                if time == now_time:
                    measurement = "aliyunSettleBill"
                    tags = {
                        "BillingCycle":billing_cycle,
                        "AccountID":account_id
                    }
                    fields = {
                        "ProductName":i["ProductName"],
                        "SubOrderId":i["SubOrderId"],
                        "BillAccountID":i["BillAccountID"],
                        "DeductedByCashCoupons":i["DeductedByCashCoupons"],
                        "PaymentTime":i["PaymentTime"],
                        "PaymentAmount":i["PaymentAmount"],
                        "DeductedByPrepaidCard":i["DeductedByPrepaidCard"],
                        "InvoiceDiscount":i["InvoiceDiscount"],
                        "UsageEndTime":i["UsageEndTime"],
                        "Item":i["Item"],
                        "SubscriptionType":i["SubscriptionType"],
                        "PretaxGrossAmount":i["PretaxGrossAmount"],
                        "Currency":i["Currency"],
                        "CommodityCode":i["CommodityCode"],
                        "UsageStartTime":i["UsageStartTime"],
                        "AdjustAmount":i["AdjustAmount"],
                        "Status":i["Status"],
                        "DeductedByCoupons":i["DeductedByCoupons"],
                        "RoundDownDiscount":i["RoundDownDiscount"],
                        "ProductDetail":i["ProductDetail"],
                        "ProductCode":i["ProductCode"],
                        "ProductType":i["ProductType"],
                        "OutstandingAmount":i["OutstandingAmount"],
                        "BizType":i["BizType"], 
                        "PipCode":i["PipCode"],
                        "PretaxAmount":i["PretaxAmount"],
                        "OwnerID":i["OwnerID"],
                        "BillAccountName":i["BillAccountName"],
                        "RecordID":i["RecordID"],
                        "CashAmount":i["CashAmount"],
                    }
                    try:
                        status_code, result = datakit.write_logging(measurement=measurement, tags=tags, fields=fields)
                        print(status_code,result)
                    except:
                        print("插入失败！")
                else:
                    break
            else:
                continue
            break
        else:
            break


```

4. **保存 **配置并 **发布**

![image.png](imgs/input-aliyun-charges-06.png)

5. 添加自动触发任务，管理 - 自动触发配置 - 新建任务

由于账单为每日账单，所以采集频率设置每天一次就可以了<br />![image.png](imgs/input-aliyun-charges-07.png)<br />
![image.png](imgs/input-aliyun-charges-08.png)

6. 日志预览

![image.png](imgs/input-aliyun-charges-09.png)

### 链路分析
暂无

# 场景视图
暂无

# 查看器
![image.png](imgs/input-aliyun-charges-10.png)

# 异常检测
暂无

# 最佳实践
暂无

# 故障排查

1. Func 日志路径：/usr/local/dataflux-func/data/logs/dataflux-func.log
1. 代码调试，选择主函数，直接运行 (可以看到脚本输出)

![image.png](imgs/input-aliyun-charges-11.png)

