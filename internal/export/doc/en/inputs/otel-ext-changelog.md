---
skip: 'not-searchable-on-index-page'
title: 'OTEL Plugins ChangeLog'
---

## Intro {#intro}

OpenTelemetry native agent does not cover every mainstream framework equally.
This extension set improves compatibility and trace data quality for additional frameworks.

Current extensions are available for:

<!-- markdownlint-disable MD046 MD030 -->
<div class="grid cards" markdown>

-   :material-language-java: **Java**

    ---

    [Download SDK :material-download:](https://static.<<<custom_key.brand_main_domain>>>/dd-image/opentelemetry-javaagent.jar){:target="_blank"}

</div>
<!-- markdownlint-enable MD046 MD030 -->

> V1 is no longer maintained; V2 is in production.

## Changelog {#changelog}

## 2.20.0-ext (2025/9/24) {#cl-2.20.0-ext}

### New {#cl-2.20.0-ext-new}

- Merge OpenTelemetry `v2.20.0`.
- Merge SQL obfuscation from upstream.

## 1.28.0-ext (2023/7/7) {#cl-1.28.0-ext}

### New {#cl-1.28.0-ext-new}

- Merge OpenTelemetry latest version for `v1.28.0`.

## 1.26.2-ext (2023/6/15) {#cl-1.26.2-ext}
Download jar: [v1.26.2-ext](https://static.<<<custom_key.brand_main_domain>>>/dd-image/opentelemetry-javaagent-1.26.2-ext.jar){:target="_blank"}

### New {#cl-1.26.2-ext-new}

- Add DB statement obfuscation logic.

## 1.26.1-ext (2023/6/9) {#cl-1.26.1-ext}

### New {#cl-1.26.1-ext-new}

- Support non-invasive method argument capture.
- Integrate Alibaba HSF framework.

## 1.26.0-ext (2023/6/1) {#cl-1.26.0-ext}

### New {#cl-1.26.0-ext-new}

- Merge OpenTelemetry `v1.26.0`.
- Support Dameng database.

## 1.25.0-ext (2023/5/10) {#cl-1.25.0-ext}

### New {#cl-1.25.0-ext-new}

- Merge OpenTelemetry latest version for `v1.25.0`.
- Support xxl-job 2.3.
- Support Alibaba Dubbo and Dubbox frameworks.
- Add Thrift framework support.
