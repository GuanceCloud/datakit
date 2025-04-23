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

The Windows Remote Collector gathers various Windows basic metrics and objects remotely via WMI or SNMP, and uploads the data to Guance Cloud to help monitor and analyze Windows system anomalies.

## Configuration {#config}

### Prerequisites {#requirements}

- Windows 2008 R2 or later  
- For SNMP collection:  
    - SNMP service must be enabled on Windows (Only SNMP v2c is supported by the collector)

### Enabling Windows SNMP Service {#install-windows-snmp}

Complete steps to install the SNMP service, configure access permissions, and firewall rules via PowerShell. Full commands:  

```shell
# Install service
Install-WindowsFeature -Name 'SNMP-Service'

# Configure community string
$regPath = "HKLM:\SYSTEM\CurrentControlSet\services\SNMP\Parameters\ValidCommunities"
New-ItemProperty -Path $regPath -Name "datakit" -PropertyType DWord -Value 4 -Force

# Configure allowed hosts
$allowedHostsPath = "HKLM:\SYSTEM\CurrentControlSet\services\SNMP\Parameters\PermittedManagers"
New-ItemProperty -Path $allowedHostsPath -Name "1" -PropertyType String -Value "192.168.1.0" -Force

# Configure firewall
netsh advfirewall firewall add rule name="datakit-snmp" dir=in action=allow protocol=UDP localport=161

# Restart service
Restart-Service SNMP
```

Key Notes:  

1. Must be executed in PowerShell with Administrator privileges
1. Community string must match `inputs.windows_remote.snmp.community` in the collector's configuration file. Recommended value: `datakit`  
1. Replace `192.168.1.0` with actual client IP/subnet (e.g., `10.10.1.0/24`). To allow all IPs: Use `-Name "*"`  
1. Firewall rule: Opens UDP port 161 (standard SNMP port). Rule name `datakit-snmp` can be customized  
1. Restart service to apply configurations  

Use the following PowerShell commands to validate configurations:  

```shell
# Check service status (ensure status is "Running")
Get-Service -Name "SNMP Service"

# View firewall rules
netsh advfirewall firewall show rule name="datakit-snmp"

# Optional: Verify connectivity using SNMP client tools
snmpwalk -v 2c -c datakit <Local_IP> .1.3.6.1.2.1.1.1.0
```

"In `conf.d/windows_remote.conf`, set the `mode` configuration option to `snmp` to enable the SNMP collector functionality."

### Collector Configuration {#input-config}

<!-- markdownlint-disable MD046 -->
=== "Host installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    After configuration, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can now be turned on by [configMap injection collector configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).



## WMI Collection Guide {#wmi}

1 Prerequisites

Ensure the following services are running:

1. Service name for Windows Management Instrumentation (WMI): `winmgmt`
2. Service name for Remote Procedure Call (RPC): `RpcSs`
3. PowerShell version is 4.0 or above.
4. The Windows Server version from which data is collected is 2008 or above.
5. The Windows server used for installing the DataKit collector must be `2012R2` or above.

2 Firewall Configuration

Run PowerShell as Administrator and execute these commands:

```shell
# Allow WMI DCOM communication
New-NetFirewallRule -DisplayName "WMI_DCOM" -Direction Inbound -Protocol TCP -LocalPort 135 -Action Allow -Program "%SystemRoot%\system32\svchost.exe" -Service rpcss

# Allow WMI service communication (dynamic ports)
New-NetFirewallRule -DisplayName "WMI_Service" -Direction Inbound -Protocol TCP -LocalPort Any -Action Allow -Program "%SystemRoot%\system32\svchost.exe" -Service winmgmt

# Allow WMI Unsecured Application communication
New-NetFirewallRule -DisplayName "WMI_UnsecApp" -Direction Inbound -Action Allow -Program "%SystemRoot%\system32\wbem\unsecapp.exe"
```

3 User Permission Setup

Run this PowerShell script to create a dedicated user and grant WMI access:

```shell
<#
Copy and modify from https://github.com/grbray/PowerShell/blob/main/Windows/Set-WMINameSpaceSecurity.ps1
#>

$UserName = "datakit"
$password = "xxxxxxxxxxxxx" | ConvertTo-SecureString -AsPlainText -Force
$namespaces = "root/cimv2", "root/wmi"
$allowInherit = $false
$deny = $false
$computerName = "."

<#delete datakit user#>
try {
    $ExistingUser = Get-LocalUser -Name $UserName -ErrorAction SilentlyContinue
    if ($ExistingUser) {
        Write-Host "[1] delete $UserName ..."
        Remove-LocalUser -Name $UserName -Confirm:$false
        Start-Sleep -Seconds 1
    }
} catch {
    throw "delete user err: $_"
}

<#create user datakit.#>
New-LocalUser -Name $UserName -Password $password -Description "user datakit for remote_wmi" -AccountNeverExpires

<# Add groups.#>
Add-LocalGroupMember -Group "Administrators" -Member $UserName
Write-Host "Member Of Administrators"
Add-LocalGroupMember -Group "Distributed COM Users" -Member $UserName
Write-Host "Member Of DCOM"
Add-LocalGroupMember -Group "Performance Monitor Users" -Member $UserName
Write-Host "Member Of Perform"
Add-LocalGroupMember -Group "Event Log Readers" -Member $UserName
Write-Host "Member Of log"

if ($PSBoundParameters.ContainsKey("Credential")) {
    $remoteparams = @{ComputerName=$computerName;Credential=$credential}
} else {
    $remoteparams = @{ComputerName=$computerName}
}


$computerName = (Get-WmiObject @remoteparams Win32_ComputerSystem).Name

<#1:Enable, 2:MethodExecute, 0x20:RemoteAccess#>
$accessMask = 0x23

$domain = $computerName

<# add wmi namespace #>
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
    if ($allowInherit) {
        $ace.AceFlags = $CONTAINER_INHERIT_ACE_FLAG
    } else {
        $ace.AceFlags = 0
    }

    $trustee = (New-Object System.Management.ManagementClass("win32_Trustee")).CreateInstance()
    $trustee.SidString = $win32account.Sid
    $ace.Trustee = $trustee

    $ACCESS_ALLOWED_ACE_TYPE = 0x0
    $ACCESS_DENIED_ACE_TYPE = 0x1
    if ($deny) {
        $ace.AceType = $ACCESS_DENIED_ACE_TYPE
    } else {
        $ace.AceType = $ACCESS_ALLOWED_ACE_TYPE
    }
    $acl.DACL += $ace.psobject.immediateBaseObject

    $setparams = @{Name="SetSecurityDescriptor";ArgumentList=$acl.psobject.immediateBaseObject} + $invokeParams

    $output = Invoke-WmiMethod @setparams
    if ($output.ReturnValue -ne 0) {
        throw "SetSecurityDescriptor failed: $($output.ReturnValue)"
    }
}
```

4 Connectivity Verification

Run this command to test WMI access:

```shell
wmic /node:"<TargetIP>" /user:".\datakit" /password:"<Password>" os get Name
```

✅ Expected Result: Returns the target host's `Name` information.

---

### DataKit Deployment {#datakit-deployment}

**Step 1**: Install DataKit

Install DataKit using the official script, then:

Set `mode` to `wmi` in the collector config file (Config path: `C:\Program Files\datakit\conf.d\wmi.conf`).

**Step 2**: Configure Service Permissions

Method 1: GUI

1. Open Services Manager (`services.msc`)
2. Find `datakit.ext` → Right-click Properties
3. Go to Log On tab → Select This account  • Account format: `.\datakit` (local) or `Domain\User`
4. Enter the password and confirm


After the modification is completed, DataKit will be restarted to collect WMI metrics and logs.
