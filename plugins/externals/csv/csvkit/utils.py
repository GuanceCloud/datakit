# -*- encoding: utf8 -*-
import sys
import argparse
import os
import time
import xlwt
import csv

from csvkit.network import Downloder

EXCEL_SUFFIX = ".xlsx"
TEMP_DIR     = ""
HTTP_ADDR    = ""
DEFAULT_SHEET_NAME = "Sheet1"

class Csv2Excel:
    def __init__(self, conv_files):
        self.conv_files = conv_files

    def convert(self):
        with open(self.conv_files[0]) as f:
            book = xlwt.Workbook(encoding='utf-8')
            sheet = book.add_sheet(DEFAULT_SHEET_NAME)
            f_csv = csv.reader(f)

            for r, row in enumerate(f_csv):
                for c, column in enumerate(row):
                    sheet.write(r, c, column)

            book.save(self.conv_files[1])

def is_http_url(url_path):
    if not isinstance(url_path, str):
        return False

    url_path = url_path.strip()
    if url_path.startswith("http://") or url_path.startswith("https://"):
        return True

    return False

def exit(code):
    sys.exit(code)

def parse_cli():
    """
    解析命令行参数
    :return: 解析后参数
    """
    parser = argparse.ArgumentParser(description="Convert CSV/EXCEL file to influxdb line protocol by http transmission")
    parser.add_argument('--metric', help='Toml cfg info about metric', default="")
    parser.add_argument('--object', help='Toml cfg info about object', default="")
    parser.add_argument('--http', help='Http host address', default="http://127.0.0.1:9529")
    parser.add_argument('--log_file', help='Log file name', default="")
    parser.add_argument('--log_level', help='Log level', default="info")
    args = parser.parse_args()
    return args


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


def down_file(file):
    """
    加载或下载文件
    :param file: 本地磁盘路径或网络路径CSV/excel文件url地址
    :return:
    """
    if is_http_url(file):
        file_name = os.path.basename(file)
        file_path = os.path.join(tmp_dir(), file_name)
        downloader = Downloder((file, file_path))
        downloader.down()
        return file, file_path

    real_path = os.path.realpath(file)
    if os.path.exists(real_path) and os.path.isfile(real_path):
        return (file, real_path)

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


def get_file(file_name):
    file = down_file(file_name)
    return csv2excel(file)

def str2bool(str):
    lower = str.lower().strip()
    if lower in ["true", "t"]:
        return True
    if lower in ["false", "f"]:
        return False
    raise Exception("{} cnannot convert to bool".format(str))