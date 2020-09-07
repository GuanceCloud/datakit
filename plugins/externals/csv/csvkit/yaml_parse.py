# -*- encoding: utf8 -*-

import yaml
from csvkit.const import *

DEFAULT_START_ROW = 0
DEFAULT_MEMENT_NAME = "csv"
DEFAULT_NULL_ACTION = IGNORE
DEFAULT_TYPE = STR

class YamlParser:
    def __init__(self, yaml_cfg):
        self.yaml_cfg    = yaml_cfg
        self.parsed_data  = None

    def parse(self):
        if not self.parsed_data:
            self.parsed_data = self._parse()
        return self.parsed_data

    def _parse(self):
        data = {}
        yaml_data = yaml.load(self.yaml_cfg, Loader=yaml.FullLoader)
        self._parse_common(data, yaml_data)
        if METRICS in yaml_data:
            self._parse_metric(data, yaml_data[METRICS])
        if OBJECTS in yaml_data:
            self._parse_object(data, yaml_data[OBJECTS])
        return data

    def _parse_object(self, data, yaml_data):
        object_data = {}
        self._parse_obj(object_data, yaml_data)
        self._parse_class(object_data, yaml_data)
        self._parse_tag(object_data, yaml_data)
        data[OBJECTS] = object_data


    def _parse_obj(self, object_data, yaml_data):
        is_found = False
        data = {}
        for column in yaml_data[COLUMN]:
            if AS_OBJ in column and column[AS_OBJ] == True:
                data[INDEX] = column[INDEX]
                is_found = True
                break
        if not is_found:
            raise("miss as_object configuration")
        object_data[AS_OBJ] = data

    def _parse_tag(self, object_data, yaml_data):
        tags = []
        for column in yaml_data[COLUMN]:
            if AS_TAG in column and column[AS_TAG] == True:
                data = {}
                data[INDEX] = column[INDEX]
                data[NAME] = column[NAME]
                tags.append(data)
        object_data[AS_TAG] = tags

    def _parse_class(self, object_data, yaml_data):
        data = {}
        for column in yaml_data[COLUMN]:
            if AS_CLASS in column and column[AS_CLASS] == True:
                data[INDEX] = column[INDEX]
                break
        object_data[AS_CLASS] = data

    def _parse_metric(self, data, yaml_data):
        metric_data = {}
        self._parse_measurement(metric_data, yaml_data)
        self._parse_tags(metric_data, yaml_data)
        self._parse_fields(metric_data, yaml_data)
        self._parse_timestamp(metric_data, yaml_data)
        data[METRICS] = metric_data

    def _parse_common(self, data, yaml_data):
        if START not in yaml_data:
            data[START] = DEFAULT_START_ROW
        else:
            data[START] = yaml_data[START]

        if FILE not in yaml_data:
            raise ("Miss required `{}`".format(FILE))
        data[FILE] = yaml_data[FILE]

    def _parse_measurement(self, data, yaml_data):
        if MEMENT not in yaml_data:
            data[MEMENT] = DEFAULT_MEMENT_NAME
        else:
            data[MEMENT] = yaml_data[MEMENT]

    def _parse_tags(self, data, yaml_data):
        tag_ok = []
        for column in yaml_data[COLUMN]:
            if AS_TAG in column and column[AS_TAG] == True:
                t = {}
                t[NAME] = column[NAME]
                t[TYPE] = DEFAULT_TYPE
                t[INDEX]   = column[INDEX]
                t[NACTION] = self._get_na_action(column)
                tag_ok.append(t)
        data[TAG] = tag_ok

    def _parse_fields(self, data, yaml_data):
        fields_ok = []
        for column in yaml_data[COLUMN]:
            if AS_FIELD in column and column[AS_FIELD] == True:
                f = {}
                f[NAME]  = column[NAME]
                f[INDEX] = column[INDEX]
                f[NACTION] = self._get_na_action(column)
                f[TYPE]    = self._get_type(column)
                fields_ok.append(f)
        data[FIELD] = fields_ok

    def _parse_timestamp(self, data, yaml_data):
        for column in yaml_data[COLUMN]:
            if AS_TIME in column and column[AS_TIME] == True:
                ts_ok = {}
                ts_ok[INDEX] = column[INDEX]
                if TUNIT not in column and TIME_FORMAT not in column:
                    raise("Missed `{}` or `{}` configuration".format(TUNIT, TIME_FORMAT))
                if TUNIT in column:
                    ts_ok[TUNIT] = column[TUNIT]
                if TIME_FORMAT in column:
                    ts_ok[TIME_FORMAT] = column[TIME_FORMAT]
                data[TS] = ts_ok

    def _get_type(self, column):
        if TYPE not in column:
            t = DEFAULT_TYPE
        else:
            t = column[TYPE]

        if t not in FieldType:
            raise("Unsuported type `{}`".format(t))
        return t

    def _get_na_action(self, column):
        if NACTION not in column:
            t = DEFAULT_NULL_ACTION
        else:
            t = column[NACTION]

        if t not in NaAction:
            raise ("Unsuported {} `{}`".format(NACTION, t))
        return t