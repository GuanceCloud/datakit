---
title     : 'Browser Dialtesting'
summary   : 'Simulate browser page access, interactions, assertions, and screenshots'
tags:
  - 'Dialtesting'
  - 'Network'
__int_icon      : 'icon/dialtesting'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

[:octicons-tag-24: Version-2.1.0](../datakit/changelog-2026.md#cl-2.1.0)

---

Browser dialtesting is a `BROWSER` task type under the `inputs.dialtesting` collector. It simulates real browser access, including opening pages, clicking elements, entering text, waiting for selectors, and asserting titles or text. It reports page performance, step details, failure reasons, and failure screenshots.

For basic dialtesting node configuration, see [Network Dialtesting](dialtesting.md). This page only describes browser-specific configuration, deployment, and troubleshooting.

## Enable Browser Dialtesting {#enable}

Browser dialtesting is enabled by default. To configure it explicitly, set it in `dialtesting.conf`:

```toml
[[inputs.dialtesting]]
  # Other dialtesting node settings are omitted.

  [inputs.dialtesting.browser]
    enabled = true
```

Browser dialtesting currently supports Linux dialtesting nodes only. On non-Linux platforms, DataKit service mode does not run `BROWSER` tasks even when `enabled = true` is configured, except when using local debug verification mode.

DataKit runs `BROWSER` tasks with the embedded browser runner. The node environment must provide Chrome/Chromium. DataKit resolves the browser in the following order:

1. `[inputs.dialtesting.browser].chrome_path`
1. `CHROME_EXECUTABLE_PATH`
1. `chromium`, `google-chrome`, or `chrome` from `PATH`

Use `max_concurrency` to limit browser tasks running at the same time. `0` means no limit. On resource-limited nodes, `1` is recommended.

## Kubernetes Deployment {#kubernetes}

For Kubernetes, use the dialtesting image directly:

```text
<<<% if custom_key.brand_key == 'guance' %>>>
pubrepo.<<<custom_key.brand_main_domain>>>/datakit/datakit:<version>-dialtesting
<<<% else %>>>
pubrepo.<<<custom_key.brand_main_domain>>>/truewatch/datakit:<version>-dialtesting
<<<% endif %>>>
```

No extra `chrome_path` setting is needed when using this image.

## Host Deployment {#host}

For host deployment, install Chrome/Chromium on the dialtesting node first, then confirm browser dialtesting remains enabled. The following examples use a Linux host.

### Install Chrome/Chromium {#host-install-chrome}

On Debian/Ubuntu, install Chromium:

```shell
sudo apt-get update
sudo apt-get install -y \
  chromium \
  ca-certificates \
  fonts-liberation \
  fonts-noto-cjk \
  libatk-bridge2.0-0 \
  libgbm1 \
  libgtk-3-0 \
  libnss3
```

On Red Hat/CentOS/Rocky Linux, install Chromium:

```shell
sudo dnf install -y \
  chromium \
  google-noto-sans-cjk-fonts \
  gtk3 \
  libgbm \
  liberation-fonts \
  nss
```

If Chromium is not available from the system repository, install Google Chrome stable instead. After installation, verify the browser path and version:

```shell
for bin in chromium chromium-browser google-chrome chrome; do
  if command -v "$bin" >/dev/null 2>&1; then
    CHROME_EXECUTABLE_PATH="$(command -v "$bin")"
    break
  fi
done

echo "$CHROME_EXECUTABLE_PATH"
"$CHROME_EXECUTABLE_PATH" --version
```

If `echo "$CHROME_EXECUTABLE_PATH"` prints nothing, Chrome/Chromium is not installed correctly and must be fixed first.

### Configure DataKit {#host-config-datakit}

Copy the dialtesting input configuration:

```shell
cd /usr/local/datakit/conf.d/samples
sudo cp dialtesting.conf.sample ../dialtesting.conf
```

Edit `/usr/local/datakit/conf.d/dialtesting.conf`, confirm browser dialtesting remains enabled, and set the browser path explicitly:

```toml
[[inputs.dialtesting]]
  server = "https://dflux-dial.<<<custom_key.brand_main_domain>>>"
  region_id = "<your-private-node-id>"
  ak = "<your-ak>"
  sk = "<your-sk>"
  pull_interval = "1m"
  time_out = "30s"

  [inputs.dialtesting.browser]
    enabled = true
    chrome_path = "/usr/bin/chromium"
    max_concurrency = 1

  [inputs.dialtesting.tags]
    region = "<your-region>"
```

If you do not want to set `chrome_path` in the configuration file, set the browser path with an environment variable:

```shell
export CHROME_EXECUTABLE_PATH=/usr/bin/chromium
```

If DataKit runs as a systemd service, exporting the variable in the current shell usually does not pass it to the DataKit service process. For host deployment, setting `chrome_path` in `dialtesting.conf` is recommended. If you prefer the environment variable, write `CHROME_EXECUTABLE_PATH` into the DataKit service environment configuration and restart the service.

Restart DataKit after updating the configuration:

```shell
sudo datakit service -R
```

### Verify with a Local Task {#host-local-test}

If no BROWSER task is available from the console yet, use a local JSON task to verify the browser execution path first. `browser_config` is a YAML string. It is easier to write the browser script as YAML first, then put it into the JSON task.

Browser script example:

```yaml
name: browser-homepage
target: https://example.com
timeout_ms: 60000
viewport:
  width: 1280
  height: 720
steps:
  - name: open page
    action: goto
    url: https://example.com
  - name: assert title
    action: assert_title
    contains: Example
```

Create `/tmp/dialtesting-browser-task.json`. When writing JSON, put the YAML above into `browser_config` as a string and represent line breaks with `\n`:

```json
{
  "BROWSER": [
    {
      "name": "browser-homepage",
      "url": "https://example.com",
      "status": "OK",
      "frequency": "1m",
      "post_url": "https://openway.<<<custom_key.brand_main_domain>>>?token=<your-token>",
      "advance_options": {
        "screenshot_on_failure": true
      },
      "browser_config": "name: browser-homepage\ntarget: https://example.com\ntimeout_ms: 60000\nviewport:\n  width: 1280\n  height: 720\nsteps:\n  - name: open page\n    action: goto\n    url: https://example.com\n  - name: assert title\n    action: assert_title\n    contains: Example\n"
    }
  ]
}
```

Temporarily set `server` in `dialtesting.conf` to the local file URL and keep browser dialtesting enabled:

```toml
[[inputs.dialtesting]]
  server = "file:///tmp/dialtesting-browser-task.json"
  pull_interval = "10s"
  time_out = "30s"

  [inputs.dialtesting.browser]
    enabled = true
    chrome_path = "/usr/bin/chromium"
    max_concurrency = 1
```

After verification, restore `server`, `region_id`, `ak`, `sk`, and other settings to the real dialtesting node configuration.

Then verify with debug mode or by restarting DataKit:

```shell
datakit debug --input-conf /usr/local/datakit/conf.d/dialtesting.conf
```

If DataKit runs as a service:

```shell
sudo datakit service -R
```

After 1~2 pull intervals, check metrics:

```shell
curl -s http://127.0.0.1:9529/metrics | grep datakit_dialtesting
```

Normally, `datakit_dialtesting_task_number{protocol="BROWSER"}` is greater than 0, and `datakit_dialtesting_worker_send_points_number{protocol="BROWSER",status="ok"}` keeps increasing.

## BROWSER Task Example {#task}

In custom dialtesting tasks, a `BROWSER` task uses `browser_config` to define the browser script. `browser_config` is a YAML string that describes page navigation, interactions, and assertions.

Common `browser_config` fields:

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `name` | string | N | Script name |
| `target` | string | N | Default target URL, used when a `goto` step does not configure a URL |
| `timeout_ms` | int | N | Script timeout in milliseconds |
| `viewport.width` | int | N | Browser viewport width |
| `viewport.height` | int | N | Browser viewport height |
| `tags` | object | N | Custom tags |
| `steps` | array | Y | Browser execution steps |

`steps` can use actions and assertions such as `goto`, `click`, `input`, `wait_for_selector`, `assert_title`, `assert_url`, and `assert_text`.

In the full task JSON, `browser_config` is inside the `BROWSER` task object:

```json
{
  "BROWSER": [
    {
      "name": "browser-homepage",
      "url": "https://example.com",
      "status": "OK",
      "frequency": "1m",
      "post_url": "https://openway.<<<custom_key.brand_main_domain>>>?token=<your-token>",
      "advance_options": {
        "screenshot_on_failure": true
      },
      "browser_config": "<browser_config YAML string>"
    }
  ]
}
```

## Failure Screenshot {#screenshot}

When `advance_options.screenshot_on_failure = true` is configured, browser dialtesting generates a screenshot for the failed step. Before reporting the result, DataKit uploads the screenshot to the Dataway specified by task `post_url` and writes the uploaded screenshot object into the dialtesting result.

After a successful upload, `steps[].screenshot` is replaced from the local path with an object:

```json
{
  "id": "run_789_step_2",
  "date": "20260528",
  "file": "run_789_step_2.png",
  "size": 12345,
  "type": "image/png"
}
```

If upload fails, DataKit does not keep the local screenshot path and records the `screenshot_upload_error` field. The dialtesting result is still reported.

## Troubleshooting {#troubleshooting}

Use DataKit metrics on the dialtesting node to check task and reporting status:

```shell
curl -s http://127.0.0.1:9529/metrics | grep datakit_dialtesting
```

Check these metrics first:

```text
datakit_dialtesting_task_number
datakit_dialtesting_worker_send_points_number
datakit_dialtesting_dataway_send_failed_number
datakit_dialtesting_worker_cached_points_number
datakit_dialtesting_worker_dropped_points_number
```

Check Chrome/Chromium availability with:

```shell
echo $CHROME_EXECUTABLE_PATH
$CHROME_EXECUTABLE_PATH --version
command -v chromium || command -v chromium-browser || command -v google-chrome || command -v chrome
```

Troubleshoot common issues as follows:

- No tasks are pulled: check `server`, `region_id`, `ak`, and `sk`, and confirm that `datakit_dialtesting_task_number{protocol="BROWSER"}` is greater than 0.
- BROWSER tasks are available in the console but not executed on the node: confirm that `[inputs.dialtesting.browser].enabled = false` is not explicitly configured, and check whether the DataKit log contains `browser.enabled is false or unsupported`.
- Results are not reported: check that task `post_url` is reachable, and that `datakit_dialtesting_dataway_send_failed_number`, `datakit_dialtesting_worker_cached_points_number`, and `datakit_dialtesting_worker_dropped_points_number` do not keep increasing.
- Browser fails to start: check that `chrome_path`, `CHROME_EXECUTABLE_PATH`, or Chrome/Chromium from `PATH` is accessible to the DataKit process.
- Browser dependencies are missing: check NSS, GTK, GBM, fonts, and certificates. In Kubernetes, use the `datakit:<version>-dialtesting` image directly.
- Screenshot is not uploaded: confirm that `advance_options.screenshot_on_failure = true` is enabled and `has_screenshot` is `true` in the failed result. If upload fails, `screenshot_upload_error` appears in the result fields or step details.

Normally, the node can pull `BROWSER` tasks, `datakit_dialtesting_worker_send_points_number{status="ok"}` keeps increasing, and `datakit_dialtesting_dataway_send_failed_number`, `datakit_dialtesting_worker_cached_points_number`, and `datakit_dialtesting_worker_dropped_points_number` do not keep increasing.
