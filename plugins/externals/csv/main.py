# -*- encoding: utf8 -*-
import os
import time
import argparse
import shutil

from csvkit.network import Downloder
from csvkit.utils import is_http_url, exit
from csvkit.yaml_parse import YamlParser
from csvkit.worker import Worker
from csvkit.const import FILE
from csvkit.convert import Csv2Excel

EXCEL_SUFFIX = ".xlsx"
TEMP_DIR = ""

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
    parser.add_argument('-y','--yaml', help='YAML cfg info', default="")
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
if args.yaml == "":
    exit(0)

yaml_data = YamlParser(args.yaml).parse()
succ_csv= get_file(yaml_data[FILE])
succ_csv = csv2excel(succ_csv)
Worker(yaml_data, *[succ_csv]).run()
if TEMP_DIR != "":
    shutil.rmtree(TEMP_DIR)
exit(0)