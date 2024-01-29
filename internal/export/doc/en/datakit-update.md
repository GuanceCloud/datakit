
# DataKit Upgrade
---

DataKit supports both manual and automatic upgrade.

## Preconditions {#req}

- Automatic upgrade require DataKit version >= 1.1.6-rc1
- There is no version requirement for manual upgrade

## Manually Upgrade {#manual}

Directly execute the following command to view the current DataKit version. If the latest version is available online, the corresponding upgrade
command will be prompted, such as:

> - For remote upgrade, you must upgrade Datakit to [1.5.9](changelog.md#cl-1.5.9)+
> - If [DataKit < 1.2.7](changelog.md#cl-1.2.7), you can only use `datakit --version`
> - If DataKit < 1.2.0, [use the upgrade command directly](changelog.md#cl-1.2.0-break-changes)
<!-- markdownlint-disable MD046 -->
=== "Linux/macOS"

    ``` shell
    $ datakit version
    
           Version: 1.2.8
            Commit: e9ccdfbae4
            Branch: testing
     Build At(UTC): 2022-03-11 11:07:06
    Golang Version: go version go1.18.3 linux/amd64
          Uploader: xxxxxxxxxxxxx/xxxxxxx/xxxxxxx
    ReleasedInputs: all
    ---------------------------------------------------
    
    Online version available: 1.2.9, commit 9f5ac898be (release at 2022-03-10 12:03:12)
    
    Upgrade:
{{ InstallCmd 4 (.WithPlatform "unix") (.WithUpgrade true) }}
    ```

=== "Windows"

    ``` powershell
    $ datakit.exe version
    
           Version: 1.2.8
            Commit: e9ccdfbae4
            Branch: testing
     Build At(UTC): 2022-03-11 11:07:36
    Golang Version: go version go1.18.3 linux/amd64
          Uploader: xxxxxxxxxxxxx/xxxxxxx/xxxxxxx
    ReleasedInputs: all
    ---------------------------------------------------
    
    Online version available: 1.2.9, commit 9f5ac898be (release at 2022-03-10 12:03:12)
    
    Upgrade:
{{ InstallCmd 4 (.WithPlatform "windows") (.WithUpgrade true) }}
    ```
---

If the DataKit is currently in proxy mode, the proxy settings will be automatically added to the prompt command of automatic upgrade:

=== "Linux/macOS"

    ```shell
    HTTPS_PROXY=http://10.100.64.198:9530 DK_UPGRADE=1 ...
    ```

=== "Windows"

    ``` powershell
    $env:HTTPS_PROXY="http://10.100.64.198:9530"; $env:DK_UPGRADE="1" ...
    ```
<!-- markdownlint-enable -->
## Auto Upgrade {#auto}

In Linux, in order to facilitate the automatic upgrade of DataKit, tasks can be added through crontab to realize regular upgrade.

> Note: Currently, automatic upgrade only support Linux, and proxy mode is not supported for the time being.

### Prepare to Upgrade Script {#prepare}

Copy the following script contents to the installation directory of the machine where the DataKit is located and save `datakit-upgrade.sh` (name optional).

```bash
#!/usr/bin/env bash
# Upgrade DataKit if new version available

echo "Checking for available upgrade..."

if [ ! -x /usr/local/datakit/datakit ]; then
  echo "/usr/local/datakit/datakit cmd not found, has datakit been installed?" >&2
  exit 1
fi

out_lines=()
while IFS='' read -r line; do out_lines+=("$line"); done < <(/usr/local/datakit/datakit version)

if [ ${#out_lines[@]} -lt 4 ]; then
  echo "invalid version output" >&2
  exit 1
fi

for ((i=0;i<${#out_lines[@]};i++))
do
  line="${out_lines[$i]}"
  if [[ "$line" =~ Upgrade: ]] && [ $((i+1)) -lt ${#out_lines[@]} ]; then
    cmd="${out_lines[$((i+1))]}"
    break
  fi
done

if [ -z "$cmd" ]; then
  echo "already up-to-date!" >&2
  exit 0
fi

if [[ "$cmd" =~ DK_UPGRADE ]]; then
  if ! bash -c "$cmd"; then
    echo "fail to upgrade" >&2
    exit 2
  fi
else
  printf "get invalid upgrade cmd: %s\n" "$cmd" >&2
  exit 3
fi

echo "successfully upgrade!"
```

### Add Crontab Task {#add-crontab}

Execute the following command to enter the crontab rule addition interface:

```shell
crontab -u root -e
```

Add the following rule:

```shell
# Mean to try the new version upgrade every morning
0 0 * * * bash /path/to/datakit-upgrade.sh >>/var/log/datakit/auto-upgrade.log 2>&1
```

Tips: crontab, The basic syntax is as follows

```txt
*   *   *   *   *     <command to be execute>
^   ^   ^   ^   ^
|   |   |   |   |
|   |   |   |   +----- day of week(0 - 6) (Sunday=0)
|   |   |   +--------- month (1 - 12)   
|   |   +------------- day of month (1 - 31)
|   +----------------- hour (0 - 23)   
+--------------------- minute (0 - 59)
```

Execute the following command to ensure that crontab is installed successfully:

```shell
crontab -u root -l
```

Make sure the crontab service starts:

```shell
service crond restart
```

If the installation is successful and an upgrade is attempted, you can see logs like the following in `upgrade_log`:

```txt
2021-05-10T09:49:06.083+0800 DEBUG  ota-update datakit/main.go:201  get online version...
2021-05-10T09:49:07.728+0800 DEBUG  ota-update datakit/main.go:216  online version: datakit 1.1.6-rc0/9bc4b960, local version: datakit 1.1.6-rc0-62-g7a1d0956/7a1d0956
2021-05-10T09:49:07.728+0800 INFO  ota-update datakit/main.go:224  Up to date(1.1.6-rc0-62-g7a1d0956)
```

If an upgrade does occur, you will see an upgrade log similar to the following:

```txt
2021-05-10T09:52:18.352+0800 DEBUG ota-update datakit/main.go:201 get online version...
2021-05-10T09:52:18.391+0800 DEBUG ota-update datakit/main.go:216 online version: datakit 1.1.6-rc0/9bc4b960, local version: datakit 1.0.1/7a1d0956
2021-05-10T09:52:18.391+0800 INFO  ota-update datakit/main.go:219 New online version available: 1.1.6-rc0, commit 9bc4b960 (release at 2021-04-30 14:31:27)
...
```

## Remote Upgrade {#remote}

[:octicons-tag-24: Version-1.5.9](changelog.md#cl-1.5.9) · [:octicons-beaker-24: Experimental](index.md#experimental)

If there are many Datakit need to upgrade, we can use remote upgrade via HTTP request. Before we use remote upgrade, we first need to upgrade or install with option `DK_UPGRADE_MANAGER=1`:

```shell
DK_UPGRADE=1 \
  DK_UPGRADE_MANAGER=1 \
  bash -c "$(curl -L https://static.guance.com/datakit/install.sh)"
```

The remote upgrade service currently provides two http APIs:

- **View the current Datakit version and available upgrade versions**

| API                                                   | Method |
|-------------------------------------------------------|--------|
| `http://<datakit-ip-or-host>:9542/v1/datakit/version` | `GET`  |


Example：

```shell
$ curl 'http://127.0.0.1:9542/v1/datakit/version'
{
    "Version": "1.5.7",
    "Commit": "1a9xxxxxxx",
    "Branch": "master",
    "BuildAtUTC": "2023-03-29 07:03:35",
    "GoVersion": "go version go1.18.3 darwin/arm64",
    "Uploader": "someone",
    "ReleasedInputs": "all",
    "AvailableUpgrades": [
        {
            "version": "1.5.8",
            "commit": "d8d2218354",
            "date_utc": "2023-03-24 11:12:54",
            "download_url": "https://static.guance.com/datakit/install.sh",
            "version_type": "Online"
        }
    ]
}
```


- **Upgrade the current Datakit to the latest version**

| API                                                   | Method |
|-------------------------------------------------------|--------|
| `http://<datakit-ip-or-host>:9542/v1/datakit/upgrade` | `POST` |

Example：

```shell
$ curl -X POST 'http://127.0.0.1:9542/v1/datakit/upgrade'
{"msg":"success"}
```
<!-- markdownlint-disable MD046 -->
???+ info

    The upgrade process may take a long time.
<!-- markdownlint-enable -->
## DataKit Version Downgrade {#downgrade}

If the new version is unsatisfactory and eager to roll back the recovery function of the old version, you can directly reverse upgrade in the following ways:
<!-- markdownlint-disable MD046 -->
=== "Linux/macOS"

    ```shell
{{ InstallCmd 4 (.WithPlatform "unix") (.WithUpgrade true) (.WithVersion "1.2.3") }}
    ```
=== "Windows"

    ```powershell
{{ InstallCmd 4 (.WithPlatform "windows") (.WithUpgrade true) (.WithVersion "1.2.3") }}
    ```
<!-- markdownlint-enable -->
The version number here can be found on the [DataKit release history](changelog.md) page. Currently, only rollback to [1.2.0](changelog.md#cl-1.2.0) is supported, and previous rc versions do not recommend rollback. After rolling back the version, you may encounter some configurations that are only available in the new version, which cannot be resolved in the rolled back version. For the time being, you can only manually adjust the configuration to adapt to the old version of DataKit.

## Version Detection Failed Processing {#version-check-failed}

During the DataKit installation/upgrade process, the installer detects the currently running version of the DataKit to ensure that the version is the upgraded version.

However, in some cases, the older version of the DataKit service did not uninstall successfully, resulting in the detection process discovering that the current running DataKit version number is still the older version number:

```shell
2022-09-22T21:20:35.967+0800    ERROR   installer  installer/main.go:374  checkIsNewVersion: current version: 1.4.13, expect 1.4.16
```

At this point, we can force the old version of DataKit to stop and restart the DataKit:

``` shell
datakit service -T # Stop service
datakit service -S # Start a new service

# If not, uninstall the DataKit service and then reinstall the service
datakit service -U # uninstall service
datakit service -I # reinstall service

# After the above operations are completed, confirm whether the next DataKit version is the latest version

datakit version # Confirm that the current running DataKit is the latest version

       Version: 1.4.16
        Commit: 1357544bd6
        Branch: master
 Build At(UTC): 2022-09-20 11:43:20
Golang Version: go version go1.18.3 linux/amd64
      Uploader: zy-infra-gitlab-prod-runner/root/xxx
ReleasedInputs: checked
```

## Offline Upgrade {#offline-upgrade}

Please refer to [Offline Install](datakit-offline-install.md) related sections.
