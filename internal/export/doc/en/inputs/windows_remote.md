---
title     : 'Windows Remote'
summary   : 'Collect metrics and object data via SNMP or WMI'
tags:
  - 'SNMP'
  - 'WMI'
  - 'WINDOWS'
__int_icon      : 'icon/windows'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

{{.AvailableArchs}}

---

The Windows Remote Collector gathers various Windows basic metrics and objects remotely via WMI or SNMP, and uploads the data to help monitor and analyze Windows system anomalies.

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

At this point, the machine being collected has been configured.

---

## DataKit Deployment {#datakit-deployment}

### 1. Create User Permissions {#creat-user}

**Objective**: Create a dedicated user `datakit` with WMI permissions to serve as the service startup account.

1. **Execute the User Creation Script**  
   After running the provided script, the system will automatically create the user `datakit` and grant it WMI operation permissions.

2. **Verify User Permissions**

   Ensure the user has the following permissions:

   - Remote WMI access permissions
   - "Log on as a service" permission (to be configured in later steps)

---

### 2. Installation and Configuration {#config-wmi}

**Objective**: Complete DataKit installation and configure WMI collection mode.

1. **Install via Official Script**  
   After executing the installation command, DataKit will be deployed by default to:  
   `C:\Program Files\datakit`

2. **Modify Collector Configuration**  
   Edit the configuration file:  
   `C:\Program Files\datakit\conf.d\windows_remote\windows_remote.conf`

---

### 3. Configure Service Permissions {#config-user}

**Key Notes:**
The DataKit service must be started using the dedicated account `datakit`, which requires the "Log on as a service" permission.

Method 1: GUI

1. Open Services Manager (`services.msc`)
2. Find `datakit.ext` → Right-click Properties
3. Go to Log On tab → Select This account  • Account format: `.\datakit` (local) or `Domain\User`
4. Enter the password and confirm

Method 2: PowerShell

Add user : `Log On as a service`, and change `datakit` server:

```shell
$username = ".\datakit"
$password = 'xxxxxxxxx'

$service = Get-WmiObject win32_service -filter "DisplayName like 'datakit'"


Set-ExecutionPolicy Bypass -Scope Process


function Add-ServiceLogonRight {
    param (
        [Parameter(Mandatory)]
        [string]$Username
    )

    try {
        $tempPath = [System.IO.Path]::GetTempFileName()
        $tmpFile = New-Item -Path $tempPath -ItemType File -Force

        secedit /export /cfg "$tmpFile.inf" | Out-Null

        $content = Get-Content "$tmpFile.inf" -Encoding ASCII
        
        if ($content -match "^SeServiceLogonRight\s*=") {
            $content = $content -replace "^SeServiceLogonRight\s*=.*", "`$0, $Username"
        } else {
            $content += "`r`nSeServiceLogonRight = $Username"
        }

        $content | Set-Content "$tmpFile.inf" -Encoding ASCII

        secedit /import /cfg "$tmpFile.inf" /db "$tmpFile.sdb" | Out-Null
        secedit /configure /db "$tmpFile.sdb" /cfg "$tmpFile.inf" | Out-Null

        Write-Host " user $Username add to Log on as a server"
    }
    catch {
        Write-Error " error: $_"
    }
    finally {
        Remove-Item "$tmpFile*" -Force -ErrorAction SilentlyContinue
    }
}


Add-ServiceLogonRight -Username "datakit"

# Stop the service if it's running
if ($service.State -eq "Running") {
    Write-Host "Stopping service $($service.DisplayName)..."
    $stopResult = $service.StopService()
    if ($stopResult.ReturnValue -eq "0") {
        Write-Host "$($service.DisplayName) stopped successfully."
    } else {
        Write-Warning "Failed to stop $($service.DisplayName). Return Value: $($stopResult.ReturnValue). Proceeding with credential change."
    }
    # Wait a short time after stopping
    Start-Sleep -Seconds 15
}

Start-Sleep -Seconds 3 # Wait before attempting to change
$returnValue = $service.Change($Null,$Null,$Null,$Null,$Null,$Null,$username,$password,$Null,$Null,$Null)
Write-Host "return value is $($returnValue.ReturnValue)"
if ($returnValue.ReturnValue -eq "0") {
    Write-Host "Successfully changed credentials for $($service.DisplayName)"
} elseif ($returnValue.ReturnValue -eq "15") {
    Write-Warning "Service database is locked for $($service.DisplayName). Retrying in 10 seconds..."
    Start-Sleep -Seconds 10
    $returnValue = $service.Change($Null,$Null,$Null,$Null,$Null,$Null,$username,$password)
    if ($returnValue.ReturnValue -eq "0") {
        Write-Host "Successfully changed credentials for $($service.DisplayName) after retry"
    } else {
        Write-Error "Failed to change credentials for $($service.DisplayName) after retry. Return Value: $($returnValue.ReturnValue)"
    }
} else {
    Write-Error "Failed to change credentials for $($service.DisplayName). Return Value: $($returnValue.ReturnValue)"
}

$service.StartService()

```

---


### Verify Deployment Results {#verify-results}

1. **Check Service Status**

   ```shell
   Get-Service datakit
   ```

   Confirm that the service status is **Running**.

2. **View Collected Metrics**

   ```shell
   datakit monitor
   ```

   The output should include WMI-related metrics (e.g., data from the `windows_remote` collector).

3. **Troubleshoot with Logs**

   Inspect the log files at:
   `C:\Program Files\datakit\log`

## Appendix {#appendix}

When querying host performance metrics via the WMI (Windows Management Instrumentation) protocol, the following classes are utilized:

| Class Name                                     | DataKit            | type    |
|:-----------------------------------------------|:-------------------|:--------|
| Win32_PerfFormattedData_PerfOS_Processor       | cpu                | metric  |
| win32_Processor                                | host_object cpu    | object  |
| Win32_LogicalDisk                              | host_object disk   | object  |
| Win32_OperatingSystem                          | host_object system | object  |
| Win32_OperatingSystem                          | mem                | object  |
| Win32_PerfFormattedData_Tcpip_NetworkInterface | net                | metric  |
| Win32_PerfFormattedData_PerfDisk_PhysicalDisk  | diskio             | metric  |
| Win32_PerfFormattedData_PerfProc_Process       | host_processes     | object  |
| Win32_Process                                  | host_processes     | object  |
| Win32_NTLogEvent                               | log                | logging |

Query these classes to verify metrics:

```shell
# Retrieve CPU metric information
Get-CimInstance -ClassName Win32_PerfFormattedData_PerfOS_Processor | Select-Object *

# Retrieve System information
Get-CimInstance -ClassName Win32_OperatingSystem | Select-Object *

# And so on...
```

Special queries for metrics:

- When querying `Win32_NetworkAdapterConfiguration`, use the query parameter `where IPEnabled = TRUE` to ensure the network adapter status.
- When querying `Win32_LogicalDisk`, specify local disks with `where DriveType=3`.
- When querying disk performance metrics from `Win32_PerfFormattedData_PerfDisk_PhysicalDisk`, use `where Name = '_Total'`.
- For CPU performance metrics from `Win32_PerfFormattedData_PerfOS_Processor`, use `where Name = '_Total'`.

All queries are executed within the namespace `root\cimv2`.
