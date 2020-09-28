# -*- encoding: utf8 -*-
import logging
import traceback
import json

import xlrd
from xlrd.sheet import Cell

from csvkit.const import *
from csvkit.utils import get_file, str2bool
from csvkit.network import ObjectSender
from csvkit.exceptions import DropException, AbortException, IgnoreException

BATCH_SIZE = 10

class ObjectCfgChecker:
    def __init__(self, sheet, parsed_toml):
        self.sheet = sheet
        self.toml  = parsed_toml
        self.cfg   = {}
        self.tag_found = []
        if ROWS in self.toml:
            self.title = self.sheet.row_values(self.toml[ROWS])
        else:
            self.title = self.sheet.row_values(0)

    def check(self):
        self._ck_common()
        self._ck_name()
        self._ck_class()
        self._ck_tags()
        self._ck_fields()
        logging.debug("checked configuration {}".format(self.cfg))
        return self.cfg

    def _ck_common(self):
        self._ck_item(self.toml, ROWS, False, self.cfg, 0)

    def _ck_name(self):
        self._ck_item(self.toml, NAME, True, self.cfg)
        name = self.cfg[NAME]
        if name not in self.title:
            raise Exception("{} configuration about {} not found in {} row{}".format(NAME, name, self.toml[FILE]), self.cfg[ROWS])
        d = {}
        d[COLUMN] = name
        d[INDEX] = self.title.index(name)
        d[TYPE] = CELL_STR
        d[NULL_OP] = NULL_OP_DROP
        self.cfg[NAME] = d
        self.tag_found.append(name)


    def _ck_class(self):
        if CLASS not in self.toml:
            return
        clas = self.toml[CLASS]
        if clas not in self.title:
            raise Exception("{} configuration about {} not found in {} row{}".format(CLASS, clas, self.toml[FILE]), self.cfg[ROWS])
        d = {}
        d[COLUMN] = clas
        d[INDEX] = self.title.index(clas)
        d[TYPE] = CELL_STR
        d[NULL_OP] = NULL_OP_IGNORE
        self.cfg[CLASS] = d
        self.tag_found.append(clas)


    def _ck_tags(self):
        tags = []
        if TAGS in self.toml:
            for tag in self.toml[TAGS]:
                if tag not in self.title:
                    raise Exception("{} configuration about {} not found in {} row{}".format(FIELD, tag, self.toml[FILE]), self.cfg[ROWS])
                t          = {}
                t[COLUMN] = tag
                t[INDEX]  = self.title.index(tag)
                t[TYPE]    = CELL_STR
                t[NULL_OP] = NULL_OP_IGNORE
                tags.append(t)
                self.tag_found.append(tag)
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
                if field in self.tag_found:
                    continue
                f[COLUMN] = field
                f[INDEX]  = index
                f[TYPE]   = CELL_STR
                f[NULL_OP] = NULL_OP_IGNORE
                fields.append(f)
        self.cfg[FIELD] = fields

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

        if key not in store_dict and default_val is not None:
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
                raise Exception("only support {} type".format(CELL_TYPE))
        except:
            raise Exception("{} cannot convert to {}".format(val_str, typ))


class ObjectSheetWorker:
    def __init__(self, toml, sheet, uploader):
        self.toml     = ObjectCfgChecker(sheet, toml).check()
        self.sheet    = sheet
        self.uploader = uploader
        self._objects = []

    def run(self):
        for r in range(self.toml[ROWS]+1, self.sheet.nrows):
            row_data = self.sheet.row(r)
            self._proc_object(r, row_data)
        self._flush()

    def _proc_object(self, r, row_data):
        try:
            self._proc_object_row(r, row_data)
        except DropException as e:
            logging.error("drop object line {} {} for {}".format(r, row_data, e))
        except AbortException:
            logging.critical("abort object line {} {}".format(r, row_data))
            exit(1)
        except:
            raise

    def _proc_object_row(self, r_index, row_data):
        obj = {}
        n = self._proc_name(r_index, row_data)
        t = self._proc_class_tags(r_index, row_data)
        f = self._proc_fields(r_index, row_data)
        obj["__name"] = n
        obj["__tags"] = t
        for k, v in f.items():
            if k not in obj:
                obj[k] = v
        self._objects.append(obj)
        if len(self._objects) >= BATCH_SIZE:
            self._flush_objects()

    def _proc_name(self, r_index, row_data):
        return self._get_item(r_index, row_data, self.toml[NAME])

    def _proc_class_tags(self, r_index, row_data):
        tags = {}
        if CLASS in self.toml:
            try:
                clas = self._get_item(r_index, row_data, self.toml[CLASS])
            except IgnoreException:
                pass
            except:
                raise
            else:
                tags[self.toml[CLASS][COLUMN]] = clas

        tags_info = self.toml.get(TAGS, [])
        for tag in tags_info:
            try:
                val = self._get_item(r_index, row_data, tag)
            except IgnoreException:
                continue
            except:
                raise
            else:
                tags[tag[COLUMN]] = val
        return tags

    def _proc_fields(self, r_index, row_data):
        f = {}
        fields_info = self.toml.get(FIELD, [])
        for field in fields_info:
            try:
                val = self._get_item(r_index, row_data, field)
            except IgnoreException:
                continue
            except:
                raise
            else:
                f[field[COLUMN]] = val
        return f

    def _get_item(self, r_index, row, item_cfg):
        if INDEX not in item_cfg:
            raise DropException()

        c = item_cfg[INDEX]
        val_cell = row[c]
        if val_cell.ctype == 0:
            val_cell = self._get_merge_val(r_index, c)  # 尝试获取合并值

        if val_cell.ctype == 0:
            val = self._process_null(item_cfg)
        else:
            val = self._conv_type(val_cell.value, item_cfg[TYPE])
        return val

    def _process_null(self, info):
        action = info[NULL_OP]
        if action == NULL_OP_ABORT:
            raise AbortException("{} configuration".format(NULL_OP_ABORT))
        elif action == NULL_OP_DROP:
            raise DropException("{} configuration".format(NULL_OP_DROP))
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
            raise DropException("{} convert to {} exception".format(val, type_str))

    def _get_merge_val(self, r, c):
        for r_min, rmax, c_min, c_max in self.sheet.merged_cells:
            if r in range(r_min, rmax) and c in range(c_min, c_max):
                return self.sheet.cell(r_min, c_min)
        return Cell(0, None)

    def _flush_objects(self):
        if len(self._objects) == 0:
            return
        data = json.dumps(self._objects)
        logging.debug("build objects: {}".format(data))
        self.uploader.send(data)
        self._objects = []

    def _flush(self):
        self._flush_objects()


class ObjectWorker:
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
                s_worker = ObjectSheetWorker(self.toml, sheet, ObjectSender())
                s_worker.run()

def collect_object(parsed_cfg):
    try:
        if FILE not in parsed_cfg:
            raise Exception("missed {} cfg".format(FILE))
        file = get_file(parsed_cfg[FILE])

        ObjectWorker(parsed_cfg, file).run()
    except Exception as e:
        logging.critical("{}".format(traceback.format_exc()))
        exit(0)