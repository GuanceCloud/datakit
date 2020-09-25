import logging
import xlrd
import time
import traceback

from xlrd.sheet import Cell

from csvkit.utils import get_file, exit, str2bool
from csvkit.const import *
from csvkit.exceptions import DropException, IgnoreException, AbortException
from csvkit.network import MetricSender

DEFAULT_METRIC = "csvmetric"
BATCH_SIZE     = 10

class MetricCfgChecker:
    def __init__(self, sheet, parsed_toml):
        self.sheet = sheet
        self.toml  = parsed_toml
        self.cfg   = {}
        self.tag_ts = []
        self.title = self.sheet.row_values(self.toml[ROWS])

    def check(self):
        self._ck_common()
        self._ck_metric()
        self._ck_timestamp()
        self._ck_tags()
        self._ck_fields()
        logging.debug("checked configuration {}".format(self.cfg))
        return self.cfg

    def _ck_common(self):
        self._ck_item(self.toml, ROWS, False, self.cfg, 0)

    def _ck_metric(self):
        self._ck_item(self.toml, METRIC, False, self.cfg, DEFAULT_METRIC)

    def _ck_tags(self):
        tags = []
        if TAGS in self.toml:
            for tag in self.toml[TAGS]:
                if tag not in self.title:
                    raise Exception()
                t          = {}
                t[COLUMN] = tag
                t[INDEX]  = self.title.index(tag)
                t[TYPE]    = CELL_STR
                t[NULL_OP] = NULL_OP_IGNORE
                tags.append(t)
                self.tag_ts.append(tag)
        self.cfg[TAGS] = tags

    def _ck_fields(self):
        fields = []
        if FIELD in self.toml:
            for field in self.toml[FIELD]:
                f = {}
                self._ck_item(field, COLUMN, True, f)
                column = f[COLUMN]
                if column not in self.title:
                    raise Exception("field {} not found in {} row".format(column, self.toml[FILE]))
                f[INDEX]   = self.title.index(column)
                self._ck_item(field, TYPE, False, f, CELL_STR, CELL_TYPE)
                self._ck_item(field, NULL_OP, False, f, NULL_OP_IGNORE, FIELD_OP)
                if NULL_OP in f and f[NULL_OP] == NULL_OP_FILL:
                    self._ck_item(field, NULL_FILL, True, None)
                    f[NULL_FILL] = self._conv_fill_value(field[NULL_FILL], field[TYPE])
                fields.append(f)
        else:
            for index, field in enumerate(self.title):
                f = {}
                if field in self.tag_ts:
                    continue
                f[COLUMN] = field
                f[INDEX]  = index
                f[TYPE]   = CELL_STR
                f[NULL_OP] = NULL_OP_IGNORE
                fields.append(f)
        if len(fields) == 0:
            raise Exception("none of valid {} configuration".format(FIELD))
        self.cfg[FIELD] = fields

    def _ck_timestamp(self):
        if TS not in self.toml:
            return
        ts = self.toml[TS]
        t = {}
        self._ck_item(ts, COLUMN, True, t)
        if TS_TF not in ts and TS_P not in ts:
            raise Exception("both {} and {} missed in {} configuration {}".format(TS_TF, TS_P, TS))
        self._ck_item(ts, TS_TF, False, t)
        self._ck_item(ts, TS_P, False, t, None, TS_P_TYPE)
        self.tag_ts.append(ts[COLUMN])

    def _ck_item(self, obj_dict, key, required, store_dict=None, default_val=None, valid_list=[]):
        if key not in obj_dict and required:
            raise Exception("missed {} configuration".format(key))

        if store_dict is None:
            return

        if key in obj_dict:
            val = obj_dict[key]
            if valid_list and val not in valid_list:
                raise Exception("{} not supported, only support {}".format(val, valid_list))
            store_dict[key] = val

        if key not in store_dict and default_val:
            store_dict[key] = default_val

    def _conv_fill_value(self, val_str, typ):
        try:
            if typ == CELL_STR:
                return val_str
            elif typ == CELL_INT:
                return int(val_str)
            elif typ == CELL_FLOAT:
                return float(val_str)
            elif typ == CELL_BOOL:
                return str2bool(val_str)
            else:
                raise
        except:
            raise DropException("{} cannot convert to {}".format(val_str, typ))


class MetricSheetWorker:
    def __init__(self, toml, sheet, uploader):
        self.toml     = MetricCfgChecker(sheet, toml).check()
        self.sheet    = sheet
        self.uploader = uploader
        self._metrics = []

    def run(self):
        for r in range(self.toml[ROWS]+1, self.sheet.nrows):
            row_data = self.sheet.row(r)
            self._proc_metric(r, row_data)
        self._flush()

    def _proc_metric(self, r, row_data):
        try:
            self._proc_metric_row(r, row_data)
        except DropException:
            logging.error("drop metrics line {} {}".format(r, row_data))
        except AbortException:
            logging.critical("abort metrics line {} {}".format(r, row_data))
            exit(1)
        except:
            raise

    def _proc_metric_row(self, r_index, row_data):
        measurement = self._build_measurement(row_data, self.toml)
        tags = self._build_tags(r_index,row_data, self.toml)
        fields = self._build_fields(r_index,row_data, self.toml)
        timestamp = self._build_timestamp(row_data, self.toml)
        ln_data = self._mk_line_proto(measurement, tags, fields, timestamp)
        logging.debug("line {} build metrics: {}".format(r_index, row_data))
        self._metrics.append(ln_data)
        if len(self._metrics) >= BATCH_SIZE:
            self._flush_metrics()

    def _build_measurement(self, row, toml_cfg):
        return toml_cfg[METRIC]

    def _build_tags(self, r_index, row, toml_cfg):
        tag_info = toml_cfg.get(TAGS, [])
        return self._build_tag_fields(tag_info, r_index, row)

    def _build_fields(self, r_index, row, toml_cfg):
        field_info = toml_cfg[FIELD]
        fields = self._build_tag_fields(field_info, r_index, row)
        if not fields:
            raise DropException()
        return fields

    def _build_tag_fields(self, item_info, r_index, row):
        tf = {}
        for info in item_info:
            r = r_index
            if INDEX not in info:
                DropException()

            c = info[INDEX]
            val_cell = row[c]
            if val_cell.ctype == 0:
                val_cell = self._get_merge_val(r, c)  # 尝试获取合并值

            try:
                if val_cell.ctype == 0:
                    val = self._process_null(info)
                else:
                    val = self._conv_type(val_cell.value, info[TYPE])
            except IgnoreException as e:
                continue
            except:
                raise

            tf[info[COLUMN]] = val
        return tf

    def _build_timestamp(self, row, toml_cfg):
        # 未指定时间戳
        if TS not in toml_cfg:
            return int(time.time()*1E9)
        ts = toml_cfg.get(TS)
        t_index = ts[INDEX]
        t = row[t_index]
        # 空值
        if t.ctype == 0:
            self._process_null(NULL_OP_DROP)
        # excel日期格式
        if t.ctype == 3:
            return int(((t.value-70*365-19)*86400-8*3600)*1E9)

        if TS_P in ts:
            try:
                return self._build_timestamp_unit(t.value, ts[TS_P])
            except:
                if TS_TF in ts:
                    return self._build_timestamp_format(t.value, ts[TS_TF])
                else:
                    raise DropException()

        if TS_TF in ts:
            return self._build_timestamp_format(t.value, ts[TS_TF])

    def _build_timestamp_unit(self, t, unit):
        t = float(t)

        if unit == TS_P_S:
            t = int(t * 1E9)
        elif unit == TS_P_MS:
            t = int(t * 1E6)
        elif unit == TS_P_US:
            t = int(t * 1E3)
        else:
            t = int(t)
        return t

    def _build_timestamp_format(self, t, format):
        t = time.strptime(t, format)
        return int(time.mktime(t)*1E9)

    def _conv_field_str(self, value):
        type_str = type(value).__name__
        if type_str == "int":
            return "{}i".format(value)
        elif type_str == "str":
            return '"{}"'.format(value)
        else:
            return "{}".format(value)

    def _process_null(self, info):
        action = info[NULL_OP]
        if action == NULL_OP_ABORT:
            raise AbortException()
        elif action == NULL_OP_DROP:
            raise DropException()
        elif action == NULL_OP_IGNORE:
            raise IgnoreException()
        elif action == NULL_OP_FILL:
            return info[NULL_FILL]
        else:
            raise DropException()

    def _conv_type(self, val, type_str):
        try:
            if type_str == CELL_STR:
                return str(val)
            elif type_str == CELL_INT:
                return int(val)
            elif type_str == CELL_FLOAT:
                return float(val)
            elif type_str == CELL_BOOL:
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

    def _flush_metrics(self):
        if len(self._metrics) == 0:
            return
        data = "\n".join(self._metrics)
        logging.debug("build metrics: {}".format(data))
        self.uploader.send(data)
        self._metrics = []

    def _flush(self):
        self._flush_metrics()


class MetricWorker:
    def __init__(self, toml_cfg, files):
        self.toml  = toml_cfg
        self.files = files
        self.max_column = 0

    def run(self):
        self.work_task(self.files[0], self.files[1])

    def work_task(self, file_url, file_path):
        with xlrd.open_workbook(file_path) as wbook:
            for name in wbook.sheet_names():
                sheet = wbook.sheet_by_name(name)
                if sheet.nrows == 0 or sheet.ncols == 0:
                    continue
                s_worker = MetricSheetWorker(self.toml, sheet, MetricSender())
                s_worker.run()

def collect_metric(parsed_cfg):
    try:
        if FILE not in parsed_cfg:
            raise Exception("missed {} cfg".format(FILE))
        logging.critical("{}".format(parsed_cfg))
        file = get_file(parsed_cfg[FILE])

        MetricWorker(parsed_cfg, file).run()
    except Exception as e:
        logging.critical("{}".format(traceback.format_exc()))
        exit(0)