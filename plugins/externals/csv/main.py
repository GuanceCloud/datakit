# -*- encoding: utf8 -*-

import os
import time
import argparse

from csvkit.network import Downloder, test_init
from csvkit.utils import is_http_url, exit
from csvkit.yaml_parse import YamlParser
from csvkit.worker import Worker
from csvkit.const import FILES
from csvkit.convert import Csv2Excel
from csvkit.log import log_file_init, log_console_init

VERSION_STR = "V0.0.2"
PATCH_STR = ""
EXCEL_SUFFIX = ".xlsx"


def mk_tmp_dir():
    try:
        os.makedirs(temp_dir)
    except FileExistsError:
        pass
    except Exception as ex:
        raise ex


def parse_cli():
    """
    解析命令行参数
    :return: 解析后参数
    """
    parser = argparse.ArgumentParser(description="Convert CSV/EXCEL file to influxdb line protocol by http transmission")

    parser.add_argument('-y','--yaml', help='YAML file url path', default=yaml_path)
    parser.add_argument('-v', '--version', action='store_true', help='Show version info', default=False)
    parser.add_argument('-c', '--console', action='store_true', help='Show line protocol data on the console during the transformation', default=False)
    parser.add_argument('-l', '--log', help='Log file name', default="")
    parser.add_argument('-t', '--test', action='store_true', help='Only convert EXCEL/CSV file to line protocol data without uploading to the dataway', default=False)
    parser.add_argument('-d', '--tmpdir', action='store', help='specify __tmp dir, default to <cwd>/__tmp', default="")

    args = parser.parse_args()
    return args

def get_yaml(url_yaml):
    """
    加载或下载YAML配置文件
    :param url_yaml: 本地磁盘路径或网络路径YAML文件url地址
    :return:
    """
    # 网络文件
    if is_http_url(url_yaml):
        file_name = os.path.basename(url_yaml)
        yamlpath = os.path.join(temp_dir, file_name)

        yaml_downloer = Downloder((url_yaml, yamlpath))
        _, fail = yaml_downloer.down()
        if len(fail):
            raise ValueError("{} download fail".format(url_yaml))
        return yamlpath
    # 本地文件
    if not os.path.isfile(url_yaml):
        raise ValueError("{} not found".format(url_yaml))
    return url_yaml


def get_file(*files):
    """
    加载或下载文件
    :param files: 本地磁盘路径或网络路径CSV/excel文件url地址列表
    :return:
    """
    down_file = []
    succ_file = []
    not_found_file = []
    for file_url in files:
        if is_http_url(file_url):
            file_name = os.path.basename(file_url)
            file_path = os.path.join(temp_dir, file_name)
            down_file.append((file_url, file_path))
            continue
        real_path = os.path.realpath(file_url)
        if os.path.exists(real_path) and os.path.isfile(real_path):
            succ_file.append((file_url, real_path))
        else:
            not_found_file.append(real_path)

    downloader = Downloder(*down_file)
    succ, fail = downloader.down()

    for file_url, _ in fail:
        not_found_file.append(file_url)

    for file_url, path in succ:
        succ_file.append((file_url, path))

    if not_found_file:
        raise ValueError("File {} not founded.".format(", ".join(not_found_file)))

    return succ_file

def csv2excel(*files):
    succ_files = []
    fail_files = []
    conv_files = []

    for url_file, file_path in files:
        if not os.path.exists(file_path) or not os.path.isfile(file_path):
            fail_files.append((url_file, file_path))
        if file_path.endswith(".csv"):
            file_name = os.path.basename(file_path)
            file_name = file_name.split(".")[0] + EXCEL_SUFFIX
            conv_path = os.path.join(temp_dir, file_name)
            conv_files.append((file_path, conv_path))
            succ_files.append((url_file, conv_path))
        elif file_path.endswith(".xlsx") or file_path.endswith(".xls"):
            succ_files.append((url_file, file_path))

    convtor = Csv2Excel(*conv_files)
    convtor.convert()
    return succ_files


yaml_path = "config.yaml"
args = parse_cli()

if args.version:
    ver_str = VERSION_STR + PATCH_STR
    print(ver_str)
    os._exit(0)


temp_dir = os.path.join(os.getcwd() , "__tmp")
if args.tmpdir:
    temp_dir = os.path.join(args.tmpdir , "__tmp")

mk_tmp_dir()

if args.console:
    log_console_init(args.console)

if args.log != "":
    log_path = os.path.join(temp_dir, args.log)
    log_file_init(log_path)

if args.test:
    test_init(args.test)

yamlpath = get_yaml(args.yaml)
yaml_data = YamlParser(yamlpath).parse()
succ_csv = get_file(*(yaml_data[FILES]))
succ_csv = csv2excel(*succ_csv)
Worker(yaml_data, *succ_csv).run()

exit(0)
