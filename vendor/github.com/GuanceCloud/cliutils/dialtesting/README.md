<!--
Unless explicitly stated otherwise all files in this repository are licensed
under the MIT License.
This product includes software developed at Guance Cloud (https://www.guance.com/).
Copyright 2021-present Guance, Inc.
-->

# Dialtesting Browser Tasks

`BROWSER` tasks run browser synthetic checks through the embedded browser runner.
The `dialtesting` package keeps the same task lifecycle as other dial
types:

```text
task JSON -> NewTask -> Run -> GetResults -> report
```

The embedded runner executes the browser script and returns the final
`browser_dial_testing` point through the existing reporting path.

## Runtime Dependencies

Browser tasks require these binaries on the dial node:

- Chrome/Chromium when `engine=chrome`
- Lightpanda when `engine=lightpanda`

Chrome can be configured by `Task.SetOption()["chrome_path"]`. Lightpanda can be
configured by `Task.SetOption()["lightpanda_path"]`.

## Task Fields

Browser tasks use the normal common task fields, including:

- `external_id`
- `name`
- `status`
- `frequency`
- `schedule_type`
- `crontab`
- `post_url`
- `tags`
- `config_vars`
- `owner_external_id`

The browser-specific task field is:

```json
{
  "browser_config": "<browser-dial YAML script>"
}
```

`browser_config` is a YAML string using the native `browser-dial` script format.
It can contain `name`, `target`, `timeout_ms`, `tags`, `metadata`, `auth`,
`config_vars`, and `steps`.

Do not put these browser script fields at the outer task level:

- `target`
- `steps`
- `auth`
- `success_when`
- `success_when_logic`
- `chrome_path`
- `lightpanda_path`

Success rules belong in `browser_config.steps` as browser assertions such as
`assert_title`, `assert_url`, and `assert_text`.

Browser runtime options stay at the outer task level:

```json
{
  "browser_window": {
    "viewports": [
      {
        "width": 1920,
        "height": 1080
      }
    ]
  },
  "advance_options": {
    "engine": "chrome",
    "screenshot_on_failure": true,
    "headers": {
      "User-Agent": "datakit-browser-dial"
    },
    "cookies": [
      {
        "name": "sid",
        "value": "example"
      }
    ],
    "ignore_https_errors": false,
    "proxy_url": ""
  },
  "retry_options": {
    "enabled": true,
    "count": 2,
    "interval_sec": 5
  }
}
```

Current limits:

- `browser_window.viewports` supports at most one viewport.
- Default viewport is `1920x1080`.
- `advance_options.engine` supports `chrome` and `lightpanda`; empty defaults
  to `chrome`.
- `retry_options.count` must be between `0` and `3`.
- `retry_options.interval_sec` must be between `5` and `300` when retry is
  enabled and count is greater than zero.
- `browser_config.timeout_ms` must not exceed `300000`.
- Each `browser_config.steps[*].timeout_ms` and
  `browser_config.auth.steps[*].timeout_ms` must not exceed `60000`.

## Browser Config Schema

`browser_config` can be written as YAML or JSON. DataKit stores it as a string
inside the browser task JSON.

Top-level fields:

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `name` | string | no | Script name. The outer task name is used when omitted. |
| `target` | string | conditional | Default URL for `goto` steps. Can be omitted when each `goto` step has `url`. |
| `post_url` | string | no | DataWay URL for standalone browser-dial usage. DataKit normally uses the outer task `post_url`. |
| `timeout_ms` | int | no | Total run timeout. DataKit normalizes missing or zero values to `300000`. |
| `headers` | map[string]string | no | Extra request headers for the browser context. |
| `cookies` | array | no | Cookies to preload into the browser context. |
| `ignore_https_errors` | bool | no | Ignore HTTPS certificate errors. |
| `proxy_url` | string | no | Browser proxy URL, for example `http://127.0.0.1:7897`. |
| `tags` | map[string]string | no | Extra result tags. |
| `metadata` | map[string]any | no | Extra result metadata. |
| `auth` | object | no | Optional authentication flow. |
| `config_vars` | array | no | Variables referenced by `value_from`. |
| `steps` | array | yes | Main browser steps. Must contain at least one step. |

`config_vars` entries:

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `name` | string | yes | Variable name. Must be unique and non-empty. |
| `value` | string | no | Variable value. |
| `secure` | bool | no | When true, the value is masked in result field `browser_config_vars`. |

`cookies` entries:

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `name` | string | yes | Cookie name. |
| `value` | string | no | Cookie value. |
| `value_from` | string | no | Read cookie value from a `config_vars` entry. Must not be used together with `value`. |
| `domain` | string | no | Cookie domain. |
| `path` | string | no | Cookie path. |
| `secure` | bool | no | Secure cookie flag. |
| `http_only` | bool | no | HTTP-only cookie flag. |
| `same_site` | string | no | `Lax`, `Strict`, or `None`. |

`auth` fields:

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `mode` | string | no | `none` or `form`. Empty means `none`. |
| `steps` | array | conditional | Required when `mode` is `form`. Uses the same step schema as `steps`. |

Step fields:

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `name` | string | no | Human readable step name. |
| `action` | string | yes | One of the supported actions below. |
| `url` | string | conditional | URL for `goto`. Optional if top-level `target` is set. |
| `selector` | string | conditional | CSS selector for selector-based actions. |
| `value` | string | conditional | Literal input value, or JavaScript expression for `eval`. |
| `value_from` | string | no | Read input value from a `config_vars` entry. |
| `text` | string | conditional | Expected text for assertions, or JavaScript expression for `eval`. |
| `contains` | string | conditional | Assertion passes when actual value contains this string. |
| `equals` | string | conditional | Assertion passes when actual value equals this string. |
| `timeout_ms` | int | no | Per-step timeout. DataKit normalizes missing or zero values to `60000`. |
| `sensitive` | bool | no | For `fill`; when true, `value_from` is required and literal `value` is forbidden. |

Supported step actions:

| Action | Required fields | Description |
| --- | --- | --- |
| `goto` | `url` or top-level `target` | Navigate to a page. |
| `wait_for_selector` | `selector` | Wait until the selector appears. |
| `click` | `selector` | Click the selector. |
| `fill` | `selector`, plus `value` or `value_from` | Fill an input. |
| `assert_title` | `contains`, `equals`, or `text` | Assert the page title. |
| `assert_url` | `contains`, `equals`, or `text` | Assert the current URL. |
| `assert_text` | `selector`, plus `contains`, `equals`, or `text` | Assert element text. |
| `eval` | `value` or `text` | Evaluate JavaScript in the page. |

Recorder tools should generate this schema directly. The Chrome extension code
does not need to live in this repository; only the generated `browser_config`
must match this contract.

## Example

```json
{
  "external_id": "bd-homepage",
  "name": "homepage",
  "status": "OK",
  "frequency": "1m",
  "schedule_type": "frequency",
  "tags": {
    "owner": "platform"
  },
  "browser_config": "name: homepage\ntarget: https://example.com\ntimeout_ms: 30000\nsteps:\n  - name: open homepage\n    action: goto\n  - name: check title\n    action: assert_title\n    contains: Example\n  - name: check body\n    action: assert_text\n    selector: body\n    contains: Example Domain\n"
}
```

More script-only examples live under `dialtesting/examples/`:

- `browser-basic.yaml`
- `browser-auth-form.yaml`
- `browser-config-vars.yaml`

## Host Validation

`GetHostName()` returns every explicitly configured browser destination so the
caller can reject illegal dial addresses before execution.

It collects hostnames from:

- top-level `browser_config.target`
- `browser_config.steps` entries where `action: goto` and `url` is set

The returned list is de-duplicated. Runtime redirects, JavaScript navigation,
third-party page assets, and `eval`-generated requests are not statically
expanded.

## Result Mapping

When `browser-dial` returns `run.success=true`:

- tag `status=OK`
- field `success=1`

When `browser-dial` succeeds after retry:

- tag `status=RETRY_OK`
- field `success=1`
- field `retry_count` greater than 0

When `browser-dial` returns `run.success=false`:

- tag `status=FAIL`
- field `success=-1`
- field `fail_reason` from `run.fail_reason` and the error message

Common result fields include:

- `response_time`
- `last_step`
- `browser_run_id`
- `exit_code`
- `message`
- `failure_type`
- `page_url`
- `page_title`
- `trace_id`
- `trace_ids`
- `ttfb`
- `loading_time`
- `lcp`
- `cls`
- `steps`
