---
title     : 'windows_remote'
summary   : 'Collect metrics and object data from SNMP or WMI devices'
tags:
  - 'SNMP'
  - 'WMI'
__int_icon      : 'icon/snmp'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---


{{.AvailableArchs}}

---

Windows Remote 采集器通过 WMI 或 SNMP 远程采集 Windows 各项基础指标和对象，并将数据上传到观测云，帮助监控分析 Windows 各种异常情况。

## 配置 {#config}

### 前置条件 {#requirements}

- Windows 2008 R2 及以上
- 使用 SNMP 采集方式，需先在 Windows 开启 SNMP 服务（采集器只支持 SNMP v2c）
- 使用 WMI 采集方式，对安装机器最低版本要求是不低于 Windows 2012

### Windows SNMP 服务开启方式 {#install-windows-snmp}

通过 PowerShell 在 Windows 系统上安装 SNMP 服务、配置访问权限及防火墙规则的完整流程。完整命令如下：

```shell
# 安装服务
Install-WindowsFeature -Name 'SNMP-Service'

# 配置社区字符串
$regPath = "HKLM:\SYSTEM\CurrentControlSet\services\SNMP\Parameters\ValidCommunities"
New-ItemProperty -Path $regPath -Name "datakit" -PropertyType DWord -Value 4 -Force

# 配置允许主机
$allowedHostsPath = "HKLM:\SYSTEM\CurrentControlSet\services\SNMP\Parameters\PermittedManagers"
New-ItemProperty -Path $allowedHostsPath -Name "1" -PropertyType String -Value "192.168.1.0" -Force

# 防火墙配置
netsh advfirewall firewall add rule name="datakit-snmp" dir=in action=allow protocol=UDP localport=161

# 重启服务
Restart-Service SNMP
```

注意以下几点：

1. 需在以管理员身份运行的 PowerShell 中执行
1. 社区字符串要跟采集器配置文件中的 `inputs.windows_remote.snmp.community` 一致，建议写 `datakit`
1. 配置允许访问的主机，将 `192.168.1.0` 需替换为实际客户端 IP/网段（例如 10.10.1.0/24），如果要允许所有 IP 可以写成 `-Name "*"`
1. 配置防火墙规则，开放 UDP 161 端口（SNMP 标准端口），规则名称 `datakit-snmp` 可自定义
1. 最后需要重启服务使配置生效

通过以下 PowerShell 命令验证服务状态：

```shell
# 检查服务状态，确认服务状态为 Running
Get-Service -Name "SNMP Service"

# 查看防火墙规则
netsh advfirewall firewall show rule name="datakit-snmp"

# 使用 SNMP 客户端工具验证连通性（可选） ：
snmpwalk -v 2c -c datakit <Localhost IP> .1.3.6.1.2.1.1.1.0
```

### 采集器配置 {#input-config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
<!-- markdownlint-enable -->

## WMI 数据采集配置指南 {#wmi}

一：前置条件

确保以下服务已开启：

1. Windows Management Instrumentation (WMI) 服务名：`winmgmt`
2. Remote Procedure Call (RPC) 服务名：`RpcSs`
3. PowerShell 版本在 4.0 及以上。
4. 被采集的 Windows Server 版本在 2008 及以上。
5. 作为安装 DataKit 采集器的 Windows 服务器需要在 2012R2 及以上。


二：防火墙配置

```powershell
# 允许 WMI DCOM 通信
New-NetFirewallRule -DisplayName "WMI_DCOM" -Direction Inbound -Protocol TCP -LocalPort 135 -Action Allow -Program "%SystemRoot%\system32\svchost.exe" -Service rpcss

# 允许 WMI 服务通信（动态端口）
New-NetFirewallRule -DisplayName "WMI_Service" -Direction Inbound -Protocol TCP -LocalPort Any -Action Allow -Program "%SystemRoot%\system32\svchost.exe" -Service winmgmt

# 允许 WMI 无安全应用通信
New-NetFirewallRule -DisplayName "WMI_UnsecApp" -Direction Inbound -Action Allow -Program "%SystemRoot%\system32\wbem\unsecapp.exe"
```


三：用户权限配置

运行 PowerShell 脚本创建专用用户并授权访问 WMI 命名空间：

```shell
<#
Copy and modify from https://github.com/grbray/PowerShell/blob/main/Windows/Set-WMINameSpaceSecurity.ps1
#>
$UserName = "datakit"
$password = "xxxxxxxxxxxx" | ConvertTo-SecureString -AsPlainText -ForceS
$namespaces = "root/cimv2", "root/wmi"
$allowInherit = $false
$deny = $false
$computerName = "."

#region
function Remove-LegacyUser {
    param($UserName)
    try {
        
        $user = Get-WmiObject -Class Win32_UserAccount -Filter "Name='$UserName' AND LocalAccount='True'"
        if ($user) {
            Write-Host "[1] delete $UserName ..."
            $adsi = [ADSI]"WinNT://$env:COMPUTERNAME"
            $adsi.Delete("User", $UserName)
            Start-Sleep -Seconds 1
        }
    } catch {
        throw "delete user error: $_"
    }
}

function New-LegacyUser {
    param($UserName, $Password, $Description)
    try {
        $adsi = [ADSI]"WinNT://$env:COMPUTERNAME"
        $user = $adsi.Create("User", $UserName)
        $user.SetPassword([Runtime.InteropServices.Marshal]::PtrToStringAuto([Runtime.InteropServices.Marshal]::SecureStringToBSTR($password)))
        $user.Put("Description", $Description)
        $user.Put("UserFlags", 0x10000)  # ADS_UF_DONT_EXPIRE_PASSWD
        $user.SetInfo()
    } catch {
        throw "creat user error: $_"
    }
}

function Add-LegacyGroupMember {
    param($Group, $User)
    try {
        $group = [ADSI]"WinNT://$env:COMPUTERNAME/$Group,group"
        $group.Add("WinNT://$env:COMPUTERNAME/$User")
    } catch {
        throw "add user to group error: $_"
    }
}
#endregion

Remove-LegacyUser -UserName $UserName

#creat user
New-LegacyUser -UserName $UserName -Password $password -Description "user datakit for remote_wmi"

#add user to group
$groups = "Administrators", "Distributed COM Users", "Performance Monitor Users", "Event Log Readers"
foreach ($group in $groups) {
    try {
        Add-LegacyGroupMember -Group $group -User $UserName
        Write-Host "add user to: $group"
    } catch {
        Write-Warning "add to $group error: $_"
    }
}

if ($PSBoundParameters.ContainsKey("Credential")) {
    $remoteparams = @{ComputerName=$computerName;Credential=$credential}
} else {
    $remoteparams = @{ComputerName=$computerName}
}

$computerName = (Get-WmiObject @remoteparams Win32_ComputerSystem).Name
$accessMask = 0x23
$domain = $computerName

foreach ($namespace in $namespaces){
    $invokeparams = @{Namespace=$namespace;Path="__systemsecurity=@"} + $remoteParams
    $output = Invoke-WmiMethod @invokeparams -Name GetSecurityDescriptor
    if ($output.ReturnValue -ne 0) {
        throw "GetSecurityDescriptor failed: $($output.ReturnValue)"
    }
    
    $acl = $output.Descriptor
    $OBJECT_INHERIT_ACE_FLAG = 0x1
    $CONTAINER_INHERIT_ACE_FLAG = 0x2

    $getparams = @{Class="Win32_Account";Filter="Domain='$domain' and Name='$UserName'"}
    $win32account = Get-WmiObject @getparams

    if ($win32account -eq $null) {
        throw "Account was not found: $UserName"
    }

    $ace = (New-Object System.Management.ManagementClass("win32_Ace")).CreateInstance()
    $ace.AccessMask = $accessMask
    Write-Host "add Enable,MethodExecute,RemoteAccess for ns:'${namespace}' user:'${UserName}'"
    
    $ace.AceFlags = if ($allowInherit) { $CONTAINER_INHERIT_ACE_FLAG } else { 0 }

    $trustee = (New-Object System.Management.ManagementClass("win32_Trustee")).CreateInstance()
    $trustee.SidString = $win32account.Sid
    $ace.Trustee = $trustee

    $ace.AceType = if ($deny) { 0x1 } else { 0x0 }
    $acl.DACL += $ace.psobject.immediateBaseObject

    $setparams = @{Name="SetSecurityDescriptor";ArgumentList=$acl.psobject.immediateBaseObject} + $invokeParams
    $output = Invoke-WmiMethod @setparams
    if ($output.ReturnValue -ne 0) {
        throw "SetSecurityDescriptor failed: $($output.ReturnValue)"
    }
}
```


四：配置验证

执行以下命令验证 WMI 连通性：

```powershell
wmic /node:"<IP>" /user:".\datakit" /password:"<password>" os get Name
```

✅ 预期结果：返回目标主机的 `Name` 信息。

至此，作为**被采集**的机器已经配置完毕。

---

### DataKit 部署流程 {#install-datakit}

**安装 DataKit：**

通过官方脚本安装 DataKit，安装完成后：

修改采集器配置文件中的 `mode` 为 `wmi` 配置 IP 地址。
   （默认配置文件路径：`C:\Program Files\datakit\conf.d\wmi.conf`）

配置服务权限：

DataKit 安装后没有采集远程服务器的权限，需要修改权限才可以。

同样需要使用上面的脚本，添加用户和赋予权限。但是 需要再增加一个权限：“作为服务登录”。

   > 注意：此用户为 DataKit 服务的启动用户，虽然名称都是一样的，但并非采集用户。

**方法一**：手动配置图形界面添加用户和权限

1. 打开服务管理器（`services.msc`）
2. 找到 `datakit` 服务 → 右键选择 属性
3. 切换至 登录 标签页 → 选择 此账户 → 账户名格式：`.\datakit`（本地用户）或 `domain\username`
4. 输入密码并确认（包括弹框确定）



修改完成之后会重启 DataKit ，就可以采集 WMI 指标和日志了。
