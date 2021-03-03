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

def _mysql_schema_size(measurement, res, config, cols=None):
	lines = []
	time = datetime.datetime.utcnow()
	for record in res:
		tags = {
			"server": config.get('host') + ":" + str(config.get('port')),
			"host": config.get("hostname", config.get("host")),
			"table_schema": record[0],
		}
		fields = {
			"schema_size_data_size": record[1],
			"schema_size_index_size": record[2],
			"schema_size_total": record[3]
		}
		line = make_line(measurement, tags, fields, time)
		lines.append(line)
	return lines

def _mysql_noindex_table(measurement, res, config, cols=None):
	lines = []
	time = datetime.datetime.utcnow()
	for record in res:
		tags = {
			"server": config.get('host') + ":" + str(config.get('port')),
			"host": config.get("hostname", config.get("host")),
			"table_schema": record[0],
			"table_name": record[1],
		}
		fields = {
			"noindex_table_data_length": record[2]
		}
		line = make_line(measurement, tags, fields, time)
		lines.append(line)
	return lines

def _mysql_slave_status(measurement, res, config, cols):
	measurement = "mysql"
	cols = [c[0] for c in cols]
	lines = []
	time = datetime.datetime.utcnow()
	for record in res:
		tags = {
			"server": config.get('host') + ":" + str(config.get('port')),
			"host": config.get("hostname", config.get("host")),
		}
		tags["name"] = config.get("name", tags["server"])
		fields = {col: val for col, val in zip(cols, record)}
		line = make_line(measurement, tags, fields, time)
		lines.append(line)
	return lines

def run(c, mock=None):
	c.log.info("mysql starting....")
	measurement = "mysqlMonitor"

	if mock is not None:
		mysqls = mock 
	else:
		mysqls = c.cfg.get("mysql")

	for config in mysqls:
		metric_info_list = [
			{
				"sql": """
					SELECT
					table_schema,
					sum( data_length ) AS data_size, 
					sum( index_length ) AS index_size,
					sum( data_length )+sum( index_length ) AS total
					FROM
					information_schema.TABLES 
					GROUP BY
					table_schema
				""",
				"type": "mysql_schema_size",
				"get_lines_func": _mysql_schema_size
			},
			{
				"sql": """
					show slave status
				""",
				"is_object": True,
				"type": "mysql_slave_status",
				"get_lines_func": _mysql_slave_status
			},
			{
				"sql": """
					SELECT t.TABLE_SCHEMA,t.table_name,t.data_length
					FROM information_schema.tables AS t
					LEFT JOIN (SELECT DISTINCT table_schema, table_name
								FROM information_schema.KEY_COLUMN_USAGE) AS kt
						ON kt.table_schema = t.table_schema
					AND kt.table_name = t.table_name
					WHERE t.table_schema NOT IN
						('mysql', 'information_schema', 'performance_schema','sys')
					AND kt.table_name IS NULL
					order by t.data_length desc limit 20
				""",
				"type": "mysql_noindex_table",
				"get_lines_func": _mysql_noindex_table
			}
		]

		for metric_info in metric_info_list:
			cursor = None
			connection = None
			try:
				connection = pymysql.connect(host=config.get("host"), port=config.get("port", 3306), user=config.get('user'), password=config.get("password"))
				cursor = connection.cursor()
				cursor.execute(metric_info.get("sql"))
				res = cursor.fetchall()
				cols = cursor.description
				get_lines_func = metric_info.get("get_lines_func")
				if not get_lines_func:
					continue
				lines = get_lines_func(measurement, res, config, cols)
				is_object = metric_info.get("is_object")
				if not any(lines):
					c.log.info("empty lines")
				else:
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

	c.log.info("mysql ok")