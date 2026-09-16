#!/usr/bin/env python3
# Unless explicitly stated otherwise all files in this repository are licensed
# under the MIT License.
# This product includes software developed at Guance Cloud (https://www.guance.com/).
# Copyright 2021-present Guance, Inc.

"""Check the published chart's Partner profile locally (Helm 3 and PyYAML)."""

from pathlib import Path
import re
import shutil
import subprocess
import tarfile
import tempfile

import yaml

ROOT = Path(__file__).resolve().parents[2]
PRESET = "values-gke-autopilot-partner.yaml"
SYNC = "gke-autopilot-partner-allowlist.yaml"
ALLOWLIST = yaml.safe_load(
    (ROOT / "docs/howtos/gke-autopilot/truewatch-datakit-autopilot-v2.yaml").read_text()
)


def helm(*args, error=None):
    result = subprocess.run(["helm", *map(str, args)], capture_output=True, text=True)
    if error is not None:
        assert result.returncode != 0, "Expected Helm to reject incompatible values"
        assert error in result.stderr, result.stderr
    elif result.returncode:
        raise RuntimeError(result.stderr)
    return result.stdout


def prepare(chart, version):
    # Render only the existing build templates in a temporary chart. Never run
    # make or overwrite the developer's generated chart metadata and values.
    params = {
        "Version": version,
        "HelmVersion": version.split("-")[0],
        "BrandDomain": "truewatch.com",
        "BrandChartRepo": "truewatch.com/chartrepo/truewatch",
        "DockerImageRepo": "pubrepo.truewatch.com/truewatch",
    }
    for source, target in [
        ("charts-Chart.template.yaml", "Chart.yaml"),
        ("charts-values.template.yaml", "values.yaml"),
        ("charts-values-gke-autopilot.template.yaml", "values-gke-autopilot.yaml"),
        ("charts-readme.template.md", "README.md"),
    ]:
        text = (ROOT / "templates" / source).read_text()
        for key, value in params.items():
            text = text.replace("{{." + key + "}}", value)
        assert "{{" not in text, f"Unhandled build template expression in {source}"
        (chart / target).write_text(text)


def render(chart, *args):
    text = helm("template", "datakit", chart, "--namespace", "datakit", "--no-hooks", *args)
    return [doc for doc in yaml.safe_load_all(text) if doc]


def resource(docs, kind):
    matches = [doc for doc in docs if doc["kind"] == kind]
    assert len(matches) == 1, f"Expected one {kind}, found {len(matches)}"
    return matches[0]


def check_partner(docs, version):
    assert {d["kind"] for d in docs} == {
        "ServiceAccount", "ClusterRole", "ClusterRoleBinding", "Secret", "Service", "DaemonSet"
    }, "YAML deployment must contain all resources and no hooks or synchronizer"
    ds = resource(docs, "DaemonSet")
    assert ds["metadata"]["labels"]["app.kubernetes.io/version"] == version
    pod = ds["spec"]["template"]
    assert pod["metadata"]["labels"]["cloud.google.com/matching-allowlist"] == ALLOWLIST["metadata"]["name"]
    spec = pod["spec"]
    assert all(spec[k] is False for k in ("hostNetwork", "hostPID", "hostIPC"))
    assert spec["dnsPolicy"] == "ClusterFirst"
    assert len(spec["containers"]) == 1 and not spec.get("initContainers")
    container = spec["containers"][0]
    approved = ALLOWLIST["matchingCriteria"]["containers"][0]
    assert container["name"] == approved["name"]
    assert container["image"] == approved["image"] + ":" + version
    for key, value in approved["securityContext"].items():
        assert container["securityContext"][key] == value
    assert container["securityContext"]["allowPrivilegeEscalation"] is False
    patterns = [entry["name"] for entry in approved["env"]]
    env = {entry["name"]: entry for entry in container["env"]}
    assert len(env) == len(container["env"]), "Duplicate environment variables"
    assert all(any(re.fullmatch(p, name) for p in patterns) for name in env)
    assert env["ENV_DATAWAY"]["valueFrom"]["secretKeyRef"] == {
        "name": "datakit-dataway", "key": "ENV_DATAWAY"
    }
    assert env["ENV_INPUT_CONTAINER_GCP_CLOUD_API_ENABLED"]["value"] == "false"
    assert env["ENV_INPUT_CONTAINER_ENABLE_K8S_NODE_LOCAL"]["value"] == "true"
    assert env["ENV_INPUT_CONTAINER_ENDPOINTS"]["value"] == "unix:///var/run/containerd/containerd.sock"
    assert env["HOST_ROOT"]["value"] == "/rootfs"
    mounts = {m["name"]: m for m in container["volumeMounts"]}
    assert set(mounts) == {m["name"] for m in approved["volumeMounts"]}
    for allowed in approved["volumeMounts"]:
        actual = mounts[allowed["name"]]
        assert all(actual[k] == v for k, v in allowed.items())
    volumes = {v["name"]: v for v in spec["volumes"]}
    approved_volumes = ALLOWLIST["matchingCriteria"]["volumes"]
    assert set(volumes) == {v["name"] for v in approved_volumes}
    for allowed in approved_volumes:
        actual = volumes[allowed["name"]]
        if "hostPath" in allowed:
            assert actual["hostPath"]["path"] == allowed["hostPath"]["path"]
            assert mounts[allowed["name"]]["readOnly"] is True
        else:
            assert actual == {"name": "cache", "emptyDir": {}}
    for rule in resource(docs, "ClusterRole")["rules"]:
        assert set(rule["verbs"]) <= {"get", "list", "watch"}
        assert not ({"pods/exec", "pods/log", "clusterroles"} & set(rule.get("resources", [])))
        assert "nonResourceURLs" not in rule
    assert resource(docs, "Secret")["metadata"]["name"] == "datakit-dataway"


def check_tolerations(chart, work):
    cases = [
        ("empty", []),
        ("single-field", [{"operator": "Exists"}]),
        ("multiple-fields", [
            {"key": "dedicated", "operator": "Equal", "value": "observability", "effect": "NoSchedule"},
        ]),
        ("no-execute", [
            {"key": "node.kubernetes.io/not-ready", "operator": "Exists",
             "effect": "NoExecute", "tolerationSeconds": 300},
        ]),
        ("multiple-entries", [
            {"key": "dedicated", "operator": "Exists", "effect": "NoSchedule"},
            {"key": "node.kubernetes.io/not-ready", "operator": "Exists",
             "effect": "NoExecute", "tolerationSeconds": 300},
        ]),
    ]
    for profile, kind, args in [
        ("regular", "DaemonSet", []),
        ("Cloud API", "Deployment", ["-f", chart / "values-gke-autopilot.yaml"]),
        ("Partner V2", "DaemonSet", ["-f", chart / PRESET]),
    ]:
        for name, tolerations in cases:
            values = work / f"tolerations-{name}.yaml"
            values.write_text(yaml.safe_dump({"tolerations": tolerations}))
            docs = render(chart, *args, "-f", values)
            pod_spec = resource(docs, kind)["spec"]["template"]["spec"]
            assert pod_spec.get("tolerations", []) == tolerations, (profile, name)


def main():
    version = yaml.safe_load((ROOT / "gitlab-ci.yml").read_text())["variables"]["CI_VERSION"]
    with tempfile.TemporaryDirectory(prefix="datakit-partner-chart-") as work:
        work = Path(work)
        chart = work / "datakit"
        shutil.copytree(ROOT / "charts/datakit", chart)
        prepare(chart, version)
        preset = ["-f", chart / PRESET]
        helm("lint", chart)
        helm("lint", chart, *preset)
        check_tolerations(chart, work)
        regular = resource(render(chart), "DaemonSet")["spec"]["template"]["spec"]
        assert regular["hostNetwork"] is True
        assert regular["containers"][0]["securityContext"]["privileged"] is True
        cloud_args = ["-f", chart / "values-gke-autopilot.yaml"]
        helm("lint", chart, *cloud_args)
        cloud = resource(render(chart, *cloud_args), "Deployment")["spec"]["template"]["spec"]
        assert all(cloud[k] is False for k in ("hostNetwork", "hostPID", "hostIPC"))
        assert cloud["volumes"] == [{"emptyDir": {}, "name": "cache"}]
        cloud_env = {e["name"]: e for e in cloud["containers"][0]["env"]}
        assert cloud_env["ENV_INPUT_CONTAINER_GCP_CLOUD_API_ENABLED"]["value"] == "true"
        docs = render(chart, *preset)
        check_partner(docs, version)
        assert docs == render(chart, *preset), "Unchanged Partner config should render deterministically"
        annotations = resource(docs, "DaemonSet")["spec"]["template"]["metadata"]["annotations"]
        changed = render(chart, *preset, "--set", "datakit.dataway_url=https://example.invalid/changed")
        new_annotations = resource(changed, "DaemonSet")["spec"]["template"]["metadata"]["annotations"]
        assert annotations["checksum/secret"] != new_annotations["checksum/secret"], "DataWay changes must restart Pods"
        values = work / "user-values.yaml"
        extra_envs = [
            {"name": "ENV_INPUT_DK_INTERVAL", "value": "30s"},
            {"name": "ENV_NAMESPACE", "valueFrom": {"configMapKeyRef": {"name": "settings", "key": "namespace"}}},
            {"name": "ENV_HTTP_PUBLIC_APIS", "valueFrom": {"secretKeyRef": {"name": "settings", "key": "apis"}}},
        ]
        values.write_text(yaml.safe_dump({"extraEnvs": extra_envs}))
        configured = render(chart, *preset, "-f", values)
        check_partner(configured, version)
        env = resource(configured, "DaemonSet")["spec"]["template"]["spec"]["containers"][0]["env"]
        assert all(e in env for e in extra_envs)
        for setting, error in [
            ("gkeAutopilot.enabled=true", "mutually exclusive"),
            ("workload.kind=Deployment", "requires workload.kind=DaemonSet"),
            ("gkeAutopilotPartner.allowlistName=", "requires gkeAutopilotPartner.allowlistName"),
            ("iploc.enable=true", "does not support"),
            ("git_repos.enable=true,git_repos.git_key_path=key", "does not support"),
            ("extraEnvs[0].name=HOST_ETC,extraEnvs[0].value=/etc", "names must match"),
        ]:
            helm("template", "datakit", chart, *preset, "--set", setting, error=error)
        # A subsequent release must select its new image without updating the preset.
        prepare(chart, "9.99.0")
        check_partner(render(chart, *preset), "9.99.0")
        prepare(chart, version)
        helm("package", chart, "--destination", work)
        package = work / f"datakit-{version.split('-')[0]}.tgz"
        unpacked = work / "unpacked"
        with tarfile.open(package) as archive:
            assert f"datakit/{PRESET}" in archive.getnames()
            assert f"datakit/{SYNC}" in archive.getnames()
            archive.extractall(unpacked, filter="data")
        packaged_chart = unpacked / "datakit"
        sync = yaml.safe_load((packaged_chart / SYNC).read_text())
        assert sync["kind"] == "AllowlistSynchronizer"
        assert sync["spec"]["allowlistPaths"] == [
            "TrueWatch/datakit/" + ALLOWLIST["metadata"]["name"] + ".yaml"
        ]
        packaged_docs = render(packaged_chart, "-f", packaged_chart / PRESET)
        assert packaged_docs == docs, "Customer YAML rendered from the package must match local output"
        check_partner(packaged_docs, version)
    print(f"PASS: chart lint, regular/Cloud API modes, Partner V2, CI_VERSION={version}, "
          "tolerations (15 cases), next release image, ENV references, Secret rollout, invalid settings, package and YAML")


if __name__ == "__main__":
    main()
