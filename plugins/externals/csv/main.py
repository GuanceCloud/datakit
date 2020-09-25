# -*- encoding: utf8 -*-
import os
import time
import argparse
import shutil
import base64
import logging

from csvkit.network import Downloder, set_http_addr
from csvkit.utils import is_http_url, exit
from csvkit.yaml_parse import YamlParser
from csvkit.worker import Worker
from csvkit.const import FILE
from csvkit.convert import Csv2Excel
from csvkit.log import en_logging

EXCEL_SUFFIX = ".xlsx"
TEMP_DIR = ""
HTTP_ADDR = ""

def tmp_dir():
    global TEMP_DIR

    if TEMP_DIR != "":
        return TEMP_DIR

    temp_dir = "tmp_" + time.strftime("%Y-%m-%d_%H-%M-%S", time.localtime(time.time()))
    TEMP_DIR = os.path.join(os.getcwd(), "data", "csv", temp_dir)
    try:
        os.makedirs(TEMP_DIR)
    except FileExistsError:
        pass
    except Exception as ex:
        raise ex

    return TEMP_DIR

def parse_cli():
    """
    解析命令行参数
    :return: 解析后参数
    """
    parser = argparse.ArgumentParser(description="Convert CSV/EXCEL file to influxdb line protocol by http transmission")
    parser.add_argument('--yaml', help='YAML cfg info', default="")
    parser.add_argument('--http', help='Http host address', default="http://127.0.0.1:9529")
    parser.add_argument('--log_file', help='Log file name', default="")
    parser.add_argument('--log_level', help='Log level', default="info")
    args = parser.parse_args()
    return args
#
def get_file(file):
    """
    加载或下载文件
    :param files: 本地磁盘路径或网络路径CSV/excel文件url地址
    :return:
    """
    if is_http_url(file):
        file_name = os.path.basename(file)
        file_path = os.path.join(tmp_dir(), file_name)
        downloader = Downloder([(file, file_path)])
        succ, fail = downloader.down()
        if len(fail) != 0:
            raise("download file `{}` failed".format(file))
        return

    real_path = os.path.realpath(file)
    if os.path.exists(real_path) and os.path.isfile(real_path):
        return (file, real_path)
    else:
        raise Exception("file `{}` not founded.".format(file))

def read_local_cfg():
    if not os.path.exists("config.yaml"):
        return ""
    with open("1.txt", "r") as f:
        cfg_data = f.read()
    return cfg_data

def csv2excel(files):
    url_file, file_path = files

    if file_path.endswith(".xlsx") or file_path.endswith(".xls"):
        return (url_file, file_path)

    if file_path.endswith(".csv"):
        file_name = os.path.basename(file_path)
        file_name = file_name.split(".")[0] + EXCEL_SUFFIX
        conv_path = os.path.join(tmp_dir(), file_name)
        convtor = Csv2Excel(*[(url_file, conv_path)])
        convtor.convert()
        return (url_file, conv_path)

args = parse_cli()
en_logging(args.log_file, args.log_level)

try:
    logging.info("http addr: {}".format(args.http))
    set_http_addr(args.http)

    logging.info("start read yaml cfg file")
    yaml_cfg = ""
    if args.yaml != "":
        yaml_cfg =  base64.standard_b64decode(args.yaml)
    else:
        yaml_cfg = read_local_cfg()

    if yaml_cfg == "":
        logging.critical("yaml cfg empty")
        exit(0)

    logging.info("start parse yaml cfg file")
    yaml_data = YamlParser(yaml_cfg).parse()
    logging.info("parsed yaml cfg data: {}".format(yaml_data))

    succ_csv= get_file(yaml_data[FILE])
    succ_csv = csv2excel(succ_csv)
    Worker(yaml_data, *[succ_csv]).run()
    if TEMP_DIR != "":
        shutil.rmtree(TEMP_DIR)
    exit(0)
except Exception as e:
    logging.critical("{}".format(e))