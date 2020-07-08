# -*- encoding: utf8 -*-

import threading
import time
import xlrd
from xlrd.sheet import Cell
from csvkit.network import Uploder
from csvkit.exceptions import IndexException, AbortException, DropException, IgnoreException
from csvkit.const import *
from csvkit.utils import exit
from csvkit.log import log

class SheetWorker:
    def __init__(self, yaml, sheet, uploader):
        self.yaml = yaml
        self.sheet = sheet
        self.uploader = uploader
        self.rule_cell = []
        self.primary_key = set()

    def run(self):
        for rule in self.yaml[RULES]:
            rule_cell = self._mk_cell_info(rule)
            self.rule_cell.append(rule_cell)

        for r in range(self.yaml[START], self.sheet.nrows):
            row_data = self.sheet.row(r)
            for rule, rule_c in zip(self.yaml[RULES], self.rule_cell):
                try:
                    data = self.procc_row(r, row_data, rule, rule_c)
                    log(data)
                    self.uploader.upload(data)
                except DropException:
                    continue
                except AbortException:
                    exit(1)
                except:
                    raise

    def procc_row(self, r_index, row_data, rule, rule_c):
        self.check_pk(row_data, rule)
        measurement = self._build_measurement(row_data, rule)
        tags = self._build_tags(r_index,row_data, rule, rule_c)
        fields = self._build_fields(r_index,row_data, rule, rule_c)
        timestamp = self._build_timestamp(row_data, rule)
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

    def _build_measurement(self, row, rule):
        return rule[MEMENT]

    def _build_tags(self, r_index, row, rule, rule_c):
        tag_info = rule.get(TAG, [])
        return self._build_tag_fields(tag_info, rule_c, r_index, row)

    def _build_fields(self, r_index, row, rule, rule_c):
        field_info = rule[FIELD]
        fields = self._build_tag_fields(field_info, rule_c, r_index, row)
        if not fields:
            raise DropException()
        return fields

    def _build_tag_fields(self, item_info, rule_c, r_index, row):
        tf = {}
        for info in item_info:
            r = r_index
            c = 0
            if COLUMN in info:
                c = info[COLUMN]
                val_cell =  row[c]
                if val_cell.ctype == 0:
                    val_cell = self._get_merge_val(r, c)  # 尝试获取合并值
            elif CELL in info:
                val_cell = rule_c[info[CELL]]
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

    def _build_timestamp(self, row, rule):
        # 未指定时间戳
        if TS not in rule:
            return int(time.time()*1E9)
        ts = rule.get(TS)
        t_index = ts[COLUMN]
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

    def _mk_cell_info(self, rule):
        cell = []
        tag_info = rule.get(TAG, [])
        for tag in tag_info:
            if CELL in tag:
                cell.append(tag[CELL])
        field_info = rule[FIELD]
        for field in field_info:
            if CELL in field:
                cell.append(field[CELL])
        cell_grp = {}
        for r, c in cell:
            if c not in cell_grp:
                cell_grp[c] = []
            cell_grp[c].append((r,c))

        cell_strip = {}
        for c in cell_grp.keys():
            grp = cell_grp[c]
            grp_sort = sorted(grp, key=lambda item: item[0], reverse=True)
            for index in range(len(grp_sort)-1):
                i = grp_sort[index][0]
                j = grp_sort[index][1]
                pi = grp_sort[index+1][0]
                pj = grp_sort[index+1][1]
                cell_c = self.sheet.cell(i,j)
                if cell_c.ctype == 0:
                    cell_c = self._get_merge_val(i,j)
                cell_p = self.sheet.cell(pi, pj)
                if cell_p.ctype == 0:
                    cell_p = self._get_merge_val(pi, pj)
                if cell_c.ctype == 1 and cell_p.ctype == 1 and cell_c.value.startswith(cell_p.value):
                    s = cell_c.value.lstrip(cell_p.value)
                    ncell = Cell(1, s)
                    cell_strip[(i,j)] = ncell
                else:
                    cell_strip[(i,j)] = cell_c
            if grp_sort:
                i = grp_sort[-1][0]
                j = grp_sort[-1][1]
                cell_c = self.sheet.cell(i, j)
                if cell_c.ctype == 0:
                    cell_c = self._get_merge_val(i,j)
                cell_strip[(i, j)] = cell_c
        return cell_strip

    def _get_merge_val(self, r, c):
        for r_min, rmax, c_min, c_max in self.sheet.merged_cells:
            if r in range(r_min, rmax) and c in range(c_min, c_max):
                return self.sheet.cell(r_min, c_min)
        return Cell(0, None)

    def _cell_strip(self, rule):
        pass

    def _check_column_index(self, row):
        max_column = len(row)
        tags = self.yaml.get(TAG, [])
        for i, tag in enumerate(tags, 1):
            if tag[COLUMN] >= max_column:
                IndexException("{} {}th element {} out of range {}".format(TAG, i, COLUMN, max_column-1))

        fields = self.yaml.get(FIELD, [])
        for i, field in enumerate(fields, 1):
            if field[COLUMN] >= max_column:
                IndexException("{} {}th element {} out of range {}".format(FIELD, i, COLUMN, max_column-1))

        ts = self.yaml.get(TS)
        if ts[COLUMN] >= max_column:
            IndexException("{} {} out of range {}".format(TS, COLUMN, max_column-1))


class Worker:
    def __init__(self, yaml, *files):
        self.yaml = yaml
        self.files  = files
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
        #uploader = Uploder(self.yaml)
        sender = Sender(rpc_server=self.yaml)

        with xlrd.open_workbook(file_path) as wbook:
            for name in wbook.sheet_names():
                sheet = wbook.sheet_by_name(name)
                if sheet.nrows == 0 or sheet.ncols == 0:
                    continue
                s_worker = SheetWorker(self.yaml, sheet, uploader)
                s_worker.run()
        uploader.flush()
