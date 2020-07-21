# -*- encoding: utf8 -*-
import os
import threading
import time
import xlrd
from xlrd.sheet import Cell
from csvkit.exceptions import AbortException, DropException, IgnoreException
from csvkit.const import *
from csvkit.utils import exit
from csvkit.network import Sender

class SheetWorker:
    def __init__(self, yaml, sheet, uploader):
        self.yaml = yaml
        self.sheet = sheet
        self.uploader = uploader
        self.primary_key = set()


    def run(self):
        for r in range(self.yaml[START], self.sheet.nrows):
            row_data = self.sheet.row(r)
            try:
                data = self.proc_row(r, row_data, self.yaml)
                self.uploader.send(data.encode())
            except DropException:
                continue
            except AbortException:
                exit(1)
            except:
                raise

    def proc_row(self, r_index, row_data, yaml_data):
        measurement = self._build_measurement(row_data, yaml_data)
        tags = self._build_tags(r_index,row_data, yaml_data)
        fields = self._build_fields(r_index,row_data, yaml_data)
        timestamp = self._build_timestamp(row_data, yaml_data)
        return self._mk_line_proto(measurement, tags, fields, timestamp)

    def check_pk(self, row_data, rule):
        if PRMK not in rule:
            return
        cell = row_data[rule[PRMK]]
        if cell.ctype == 0:
            return
        if cell.value in self.primary_key:
            raise DropException
        self.primary_key.add(cell.value)

    def _build_measurement(self, row, yaml_data):
        return yaml_data[MEMENT]

    def _build_tags(self, r_index, row, yaml_data):
        tag_info = yaml_data.get(TAG, [])
        return self._build_tag_fields(tag_info, r_index, row)

    def _build_fields(self, r_index, row, yaml_data):
        field_info = yaml_data[FIELD]
        fields = self._build_tag_fields(field_info, r_index, row)
        if not fields:
            raise DropException()
        return fields

    def _build_tag_fields(self, item_info, r_index, row):
        tf = {}
        for info in item_info:
            r = r_index
            c = 0
            if INDEX in info:
                c = info[INDEX]
                val_cell =  row[c]
                if val_cell.ctype == 0:
                    val_cell = self._get_merge_val(r, c)  # 尝试获取合并值
            else:
                DropException()

            try:
                if val_cell.ctype == 0:
                    self._process_null(info[NACTION])
                val = self._conv_type(val_cell.value, info[TYPE])
            except IgnoreException as e:
                continue
            except:
                raise

            tf[info[NAME]] = val
        return tf

    def _build_timestamp(self, row, yaml_data):
        # 未指定时间戳
        if TS not in yaml_data:
            return int(time.time()*1E9)
        ts = yaml_data.get(TS)
        t_index = ts[INDEX]
        t = row[t_index]
        # 空值
        if t.ctype == 0:
            self._process_null(DROP)
        # excel日期格式
        if t.ctype == 3:
            return int(((t.value-70*365-19)*86400-8*3600)*1E9)

        if TUNIT in ts:
            try:
                return self._build_timestamp_unit(t.value, ts[TUNIT])
            except:
                if TIME_FORMAT in ts:
                    return self._build_timestamp_format(t.value, ts[TIME_FORMAT])
                else:
                    raise DropException()

        if TIME_FORMAT in ts:
            return self._build_timestamp_format(t.value, ts[TIME_FORMAT])

    def _build_timestamp_unit(self, t, unit):
        t = float(t)

        if unit == "s":
            t = int(t * 1E9)
        elif unit == "ms":
            t = int(t * 1E6)
        elif unit == "us":
            t = int(t * 1E3)
        else:
            t = int(t)
        return t

    def _build_timestamp_format(self, t, format):
        t = time.strptime(t, format)
        return int(time.mktime(t)*1E9)

    def _mk_line_proto(self, measurement, tags, fields, timestamp):
        line_proto_data = ""
        line_proto_data += "{}".format(measurement)
        is_frist_field = True
        # tags可选
        if tags:
            for key, val in tags.items():
                line_proto_data += ",{}={}".format(key, val)
        # fields必填
        for key, val in fields.items():
            if is_frist_field:
                prefix = " "
                is_frist_field = False
            else:
                prefix = ","

            line_proto_data += "{}{}={}".format(prefix, key, self._conv_field_str(val))
        if timestamp:
            line_proto_data += " {}".format(timestamp)
        line_proto_data += "\n"
        return line_proto_data

    def _conv_field_str(self, value):
        type_str = type(value).__name__
        if type_str == "int":
            return "{}i".format(value)
        elif type_str == "str":
            return '"{}"'.format(value)
        else:
            return "{}".format(value)

    def _process_null(self, action):
        if action == ABORT:
            raise AbortException()
        elif action == DROP:
            raise DropException()
        elif action == IGNORE:
            raise IgnoreException()
        else:
            raise DropException()

    def _conv_type(self, val, type_str):
        try:
            if type_str == STR:
                return str(val)
            elif type_str == INT:
                return int(val)
            elif type_str == FLOAT:
                return float(val)
            elif type_str == BOOL:
                return bool(val)
            else:
                raise
        except:
            raise DropException()

    def _get_merge_val(self, r, c):
        for r_min, rmax, c_min, c_max in self.sheet.merged_cells:
            if r in range(r_min, rmax) and c in range(c_min, c_max):
                return self.sheet.cell(r_min, c_min)
        return Cell(0, None)


class Worker:
    def __init__(self, yaml, *files):
        self.yaml  = yaml
        self.files = files
        self.max_column = 0

    def run(self):
        thread = []
        for file_url, file_path in self.files:
            t = threading.Thread(target=self.work_task, args=(file_url, file_path))
            thread.append(t)
            t.start()

        for t in thread:
            t.join()

    def work_task(self, file_url, file_path):
        socket = "unix://" + os.path.join(os.getcwd(), "datakit.sock")
        with xlrd.open_workbook(file_path) as wbook:
            for name in wbook.sheet_names():
                sheet = wbook.sheet_by_name(name)
                if sheet.nrows == 0 or sheet.ncols == 0:
                    continue
                s_worker = SheetWorker(self.yaml, sheet, Sender(socket))
                s_worker.run()