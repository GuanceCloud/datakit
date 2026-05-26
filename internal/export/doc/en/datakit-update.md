
# DataKit Upgrade
---

DataKit supports both manual and automatic upgrade.

## Requirements {#req}

- Automatic upgrade require DataKit version >= 1.1.6-rc1
- There is no version requirement for manual upgrade

### Manually Upgrade {#manual}

Directly execute the following command to view the current DataKit version. If the latest version is available online, the corresponding upgrade
command will be prompted, such as:

> - For remote upgrade, you must upgrade DataKit to [1.5.9](changelog.md#cl-1.5.9)+
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

### Remote Update Service {#auto}

[:octicons-tag-24: Version-1.5.9](changelog.md#cl-1.5.9) · [:octicons-beaker-24: Experimental](index.md#experimental)

> Note: The service does not support DataKit installed in k8s.

During the installation of DataKit, an additional upgrade management service is installed by default, which is specifically used to manage and upgrade the DataKit version. If you are using an older version of DataKit, you can specify additional parameters in the DataKit upgrade command to install this service:

<!-- markdownlint-disable MD046 -->

=== "Public Network Installation"

    ```shell hl_lines="2"
    DK_UPGRADE=1 \
      DK_UPGRADE_MANAGER=1 \
      bash -c "$(curl -L https://static.<<<custom_key.brand_main_domain>>>/datakit/install.sh)"
    ```

=== "Offline Update"

    [:octicons-tag-24: Version-1.38.1](changelog.md#cl-1.38.1)

    If you have [synchronized the DataKit installation package offline](datakit-offline-install.md#offline-advanced), assuming the offline installation package address is `http://my.static.com/datakit`, the upgrade command here is

    ```shell hl_lines="3"
    DK_UPGRADE=1 \
      DK_UPGRADE_MANAGER=1 \
      DK_INSTALLER_BASE_URL="http://my.static.com/datakit" \
      bash -c "$(curl -L https://static.<<<custom_key.brand_main_domain>>>/datakit/install.sh)"
    ```

You can use DCA to achieve remote updates, see [DCA Documentation](../dca/index.md).

---

???+ info

    - The upgrade process may take a long time depending on network bandwidth (essentially equivalent to manually invoking the DataKit upgrade command), please wait patiently.
    - Starting from version [Version-1.91.1](changelog-2026.md#cl-1.91.1), the direct upgrade function of DataKit through HTTP service has been removed.

<!-- markdownlint-enable -->

### Offline Upgrade {#offline-upgrade}

Please refer to [Offline Install](datakit-offline-install.md) related sections.

### DataKit Version Downgrade {#downgrade}

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
The version number here can be found on the [DataKit release history](changelog-{{.Year}}.md) page. Currently, only rollback to [1.2.0](changelog.md#cl-1.2.0) is supported, and previous rc versions do not recommend rollback. After rolling back the version, you may encounter some configurations that are only available in the new version, which cannot be resolved in the rolled back version. For the time being, you can only manually adjust the configuration to adapt to the old version of DataKit.

## FAQ {#faq}

### Differences Between Updating and Installing {#upgrade-vs-install}

To upgrade to a newer version of DataKit, you can do so by:

- Reinstallation
- [Executing the upgrade command](datakit-update.md#manual)

On a host where DataKit is already installed, it is recommended to upgrade to a newer version using the upgrade command rather than reinstalling. If you reinstall, all configurations in [*datakit.conf*](datakit-conf.md#maincfg-example) will be reset to their default settings, such as global tag configurations, port settings, and so on. This may not be desirable.

However, whether you reinstall or execute the upgrade command, all the collectors(inputs) configurations are not reset to default.

### Version Detection Failed Processing {#version-check-failed}

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
