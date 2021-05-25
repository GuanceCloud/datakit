import datetime
import requests
from influxdb.line_protocol import make_line


def get_cluster_info(host, user, password):
    response = requests.get(host, auth=(user, password))
    if response.status_code == 200:
        resp = response.json()
        cluster_name = resp.get("cluster_name")
        cluster_uuid = resp.get("cluster_uuid")
        return cluster_name, cluster_uuid
    return "", ""


def run(c):
    c.log.info("es_index starting...")
    for es_cfg in c.cfg.get("es_index", {}):
        host = es_cfg.get("host")
        user = es_cfg.get("user")
        password = es_cfg.get("password")
        es_url = "{}{}".format(host, "/_stats")
        response = requests.get(es_url, auth=(user, password))
        if response.status_code != 200:
            c.log.error("state_code:%d,resp:%s", response.status_code, response.text)
            continue
        resp = response.json()
        lines = []
        measurement = "es_index"
        t = datetime.datetime.utcnow()
        cluster_name, cluster_uuid = get_cluster_info(host, user, password)
        for k, v in resp.get("indices", {}).items():
            tags = {
                "index": k,
                "cluster_name": cluster_name,
                "cluster_uuid": cluster_uuid,
            }
            fields = {
                "size": v.get("primaries").get("store", {}).get("size_in_bytes", 0)
            }
            line = make_line(measurement, tags, fields, t)
            lines.append(line)
        body = '\n'.join(lines) + '\n'
        dk_url = "http://0.0.0.0:9529/v1/write/metric"
        result = requests.post(dk_url, data=body.encode("utf-8"))
        if result.status_code != 200:
            c.log.error("state_code:%d,resp:%s", result.status_code, result.text)
    c.log.info("es_index ok")
