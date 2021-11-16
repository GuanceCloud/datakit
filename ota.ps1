# Update DataKit if new version available
# Date: Sat Apr 10 11:40:04     2021
# Author: tanb

# Required: gsudo.exe
# See https://stackoverflow.com/a/58753166/342348

function tryUpdate {
        $check = gsudo.exe C:\Users\coano\Desktop\datakit.exe -check-update -accept-rc-version
        if (-Not $?) {
                Import-Module bitstransfer;
                start-bitstransfer -source https://static.guance.com/datakit/installer-windows-amd64.exe -destination .dk-installer.exe;
                gsudo.exe .dk-installer.exe -upgrade -ota;
                #.dk-installer.exe -download-only
                rm .dk-installer.exe
        } else {
                echo "update to date"
        }
}

tryUpdate
