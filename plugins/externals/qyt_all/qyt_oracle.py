import sys
import datetime
import requests
import cx_Oracle
import pymysql
from influxdb.line_protocol import make_line

def _write_data(lines, is_object=False):
	res = {
		"success": True
	}
	body = '\n'.join(lines) + '\n'
	url = "http://0.0.0.0:9529"
	path = "/v1/write/metric"
	if is_object:
		path = "/v1/write/object"
	dk_url = url + path
	result = requests.post(dk_url, data=body.encode("utf-8"))
	if result.status_code != 200:
		res['success'] = False
		res['result'] = result
	return res

def _oracle_slowlog_info_pt (measurement, res, config, cols=None):
	lines = []
	time = datetime.datetime.utcnow()
	for record in res:
		tags = {
			"type": "oracle_slowlog_info_pt",
			"sql_id": record[0],
			"oracle_server": config.get("oracle_server"),
			"oracle_port": config.get("oracle_port"),
			"host": config.get("host"),
			"service_name": config.get("service_name"),
			"instance_id": config.get("instance_id"),
			"instance_desc": config.get("instance_desc")
		}
		fields = {
			"sql_text": record[1],
			"executions": record[2],
			"total_time": record[3],
			"time_per_exec": record[4],
			"buffer_gets": record[5],
			"disk_reads": record[6]
		}
		line = make_line(measurement, tags, fields, time)
		lines.append(line)
	return lines

def run(c, mock=None):
	c.log.info("oracle starting....")
	if mock is not None:
		oracles = mock 
	else:
		oracles = c.cfg.get("oracle")

	measurement = "oracle_monitor"
	for config in oracles:
		# config {connect_string, }
		metric_info_list = [
			{
				"sql": """
					select sql_id "SQL ID",
					sql_text ,
					executions ,
					round(elapsed_time / 1000000, 2) "TOTAL_TIME",
					decode((round(elapsed_time / 1000000 /
					(decode(EXECUTIONS, 0, 1, EXECUTIONS)),
					2)),
					0,
					0.01,
					(round(elapsed_time / 1000000 /
					(decode(EXECUTIONS, 0, 1, EXECUTIONS)),
					2))) "TIME_PER_EXEC",
					buffer_gets ,
					disk_reads
					from (select * from v$sql where last_active_time> sysdate-1/144 order by elapsed_time/(decode(EXECUTIONS,0,1,EXECUTIONS)) desc)
					where rownum <= 10 and EXECUTIONS>=5
				""",
				"type": "oracle_slowlog_info_pt",
				"get_lines_func": _oracle_slowlog_info_pt
			}
		]

		for metric_info in metric_info_list:
			try:
				connection = cx_Oracle.connect(config.get("connect_string"), encoding="UTF-8")
				cursor = connection.cursor()
				result = cursor.execute(metric_info.get("sql"))
				res = result.fetchall()
				get_lines_func = metric_info.get("get_lines_func")
				if not get_lines_func:
					continue

				lines = get_lines_func(measurement, res, config, cols=None)
				is_object = metric_info.get("is_object")
				write_res = _write_data(lines, is_object=is_object)
				if not write_res['success']:
					c.log.error("state_code:%d,resp:%s", write_res["result"].status_code, write_res["result"].text)

			except Exception as e:
				c.log.error(e)

			finally:
				if cursor:
					cursor.close()

				if connection:
					connection.close()

	c.log.info("oracle ok")