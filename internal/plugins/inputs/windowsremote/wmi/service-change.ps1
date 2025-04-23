$ServiceName = "datakit"
$NewAccount = ".\datakit"
$Password =  "xxxxxxxxxxxx"

$Service = Get-WmiObject -Class Win32_Service -Filter "Name='$ServiceName'"

if ($Service) {
    #$ReturnValue = $Service.ChangeStartMode("Automatic", $false, $NewAccount, $Password)

    $ReturnValue = $Service.Change($null, $null, $null, $null, $null, $null, $NewAccount, $Password)
    if ($ReturnValue.ReturnValue -eq 0) {
        Write-Host "change service: '$ServiceName' new user: '$NewAccount' pw:$Password "
    } elseif ($ReturnValue.ReturnValue -eq 5) {
        Write-Error "error:"
    } else {
        Write-Error "service: '$ServiceName' filed err:: $($ReturnValue.ReturnValue)"
    }
} else {
    Write-Error "can not find service: '$ServiceName' "
}