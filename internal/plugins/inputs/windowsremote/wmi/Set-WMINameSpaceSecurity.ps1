<#
Copy and modify from https://github.com/grbray/PowerShell/blob/main/Windows/Set-WMINameSpaceSecurity.ps1
#>

$UserName = "datakit"
$password = "xxxxxxxxxxxxx" | ConvertTo-SecureString -AsPlainText -Force
$namespaces = "root/cimv2", "root/wmi"
$allowInherit = $false
$deny = $false
$computerName = "."

# delete datakit user
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

# create user datakit.
New-LocalUser -Name $UserName -Password $password -Description "user datakit for remote_wmi" -AccountNeverExpires

# Add groups.

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

#Write-Host "remoteparams is $remoteparams"

$computerName = (Get-WmiObject @remoteparams Win32_ComputerSystem).Name

# 1:Enable, 2:MethodExecute, 0x20:RemoteAccess
$accessMask = 0x23

$domain = $computerName

## add wmi namespace
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