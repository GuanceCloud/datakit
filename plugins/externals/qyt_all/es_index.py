import datetime
import sys
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


def run(host, user, password):
    es_url = "{}{}".format(host, "/_stats")
    response = requests.get(es_url, auth=(user, password))
    if response.status_code != 200:
        print(response.status_code, response.json())
        return
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
        print(result.status_code, result.content)


if __name__ == '__main__':
    if len(sys.argv) == 4:
        _, host, user, password = sys.argv
        run(host, user, password)
    else:
        print("bad params")
