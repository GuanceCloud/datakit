# -*- encoding: utf8 -*-

import yaml
from csvkit.exceptions import ConfException
from csvkit.const import *

DEFAULT_BATCH_SIZE = 10
DEFAULT_START_ROW = 0
DEFAULT_NULL_ACTION = IGNORE
DEFAULT_DATAWAY_UUID = "xxx-xxx-123"
DEFAULT_DATAWAY_AUTH = False
DEFAULT_TYPE = STR

class YamlParser:
    def __init__(self, yaml_path):
        self.yaml_path    = yaml_path
        self.parsed_data  = None

    def parse(self):
        if not self.parsed_data:
            self.parsed_data = self._parse()
        return self.parsed_data

    def _parse(self):
        data = {}
        yaml_data = self._read_yaml()
        self._parse_comm(data, yaml_data)
        rules = self._get_ck(RULES, yaml_data, list, RULES)
        data[RULES] = []
        for rule in rules:
            data_rule = {}
            self._parse_measurement(data_rule, rule)
            self._parse_tags(data_rule, rule)
            self._parse_fields(data_rule, rule)
            self._parse_timestamp(data_rule, rule)
            if PRMK in rule:
                data_rule[PRMK] = self._get_ck(PRMK, rule, int, PRMK)
            data[RULES].append(data_rule)
        return data

    def _parse_comm(self, data, yaml_data):
        data[START] = self._get_ck(START, yaml_data, int, START) if START in yaml_data else DEFAULT_START_ROW
        data[BATCH] = self._get_ck(BATCH, yaml_data, int, BATCH) if BATCH in yaml_data else DEFAULT_BATCH_SIZE
        data[UUID] = self._get_ck(UUID, yaml_data, str, UUID) if UUID in yaml_data else DEFAULT_DATAWAY_UUID
        data[AUTH] = self._get_ck(AUTH, yaml_data, bool, AUTH) if AUTH in yaml_data else DEFAULT_DATAWAY_AUTH
        data[URL] = self._get_ck(URL, yaml_data, str,  URL)
        if data[AUTH]:
            data[PK] = self._get_ck(PK, yaml_data, str, PK)
            data[SK] = self._get_ck(SK, yaml_data, str, SK)
        data[FILES] = self._get_ck(FILES, yaml_data, list, FILES)

    def _parse_measurement(self, data, yaml_data):
        data[MEMENT] = self._get_ck(MEMENT, yaml_data, str, MEMENT)

    def _parse_tags(self, data, yaml_data):
        tags = yaml_data.get(TAG, [])
        tag_ok = []
        for i, tag in enumerate(tags,1):
            t = {}
            is_unvalid = True

            t[NAME] = self._get_ck(NAME, tag, str, NAME)
            t[TYPE]   = tag.get(TYPE, DEFAULT_TYPE)

            if COLUMN in tag:
                t[COLUMN] = self._get_ck(COLUMN, tag, int, COLUMN)
                is_unvalid = False
            if CELL in tag:
                t[CELL] = eval(self._get_ck(CELL, tag, str, CELL))
                is_unvalid = False
            if is_unvalid:
                raise ConfException("Missed both {} and {} configurtion.".format(CELL, COLUMN))

            action = tag.get(NACTION, DEFAULT_NULL_ACTION)
            self._ck_type(action, str, NACTION)
            if action.lower() not in NaAction:
                raise ConfException("{} only supported `{}` configurtion.".format(NACTION, ",".join(NaAction)))
            t[NACTION] = action.lower()

            tag_ok.append(t)
        data[TAG] = tag_ok

    def _parse_fields(self, data, yaml_data):
        fields = yaml_data.get(FIELD, [])
        fields_ok = []
        for i, field in enumerate(fields, 1):
            f = {}
            is_unvalid = True

            f[NAME] = self._get_ck(NAME, field, str, NAME)
            if COLUMN in field:
                f[COLUMN] = self._get_ck(COLUMN, field, int, COLUMN)
                is_unvalid = False
            if CELL in field:
                f[CELL] = eval(self._get_ck(CELL, field, str, CELL))
                is_unvalid = False
            if is_unvalid:
                raise ConfException("Missed both {} and {} configurtion.".format(CELL, COLUMN))

            action = field.get(NACTION, DEFAULT_NULL_ACTION)
            self._ck_type(action, str,  NACTION)
            if action.lower() not in NaAction:
                raise ConfException("{} only supported `{}` configurtion.".format(NACTION, ",".join(NaAction)))
            f[NACTION] = action.lower()

            t_str = field.get(TYPE, DEFAULT_TYPE)
            self._ck_type(t_str, str, TYPE)
            if t_str.lower() not in FieldType:
                raise ConfException("{} only supported `{}` configurtion.".format(TYPE, ",".join(FieldType)))
            f[TYPE] = t_str.lower()

            fields_ok.append(f)

        if not fields_ok:
            raise ConfException("None Valid Configurations in {}".format(FIELD))

        data[FIELD] = fields_ok

    def _parse_timestamp(self, data, yaml_data):
        # 时间戳可选
        if TS in yaml_data:
            ts_ok = {}
            ts = self._get_ck(TS, yaml_data, dict, TS)
            ts_ok[COLUMN] = self._get_ck(COLUMN, ts, int, COLUMN)
            if TUNIT in ts:
                unit = self._get_ck(TUNIT, ts, str, TUNIT)
                if unit.lower() not in TsUnit:
                    raise ConfException("{} only supported `{}` configurtion.".format(TYPE, ",".join(TsUnit)))
                ts_ok[TUNIT] = unit.lower()
            if TIME_FORMAT in ts:
                time_format =  self._get_ck(TIME_FORMAT, ts, str, TIME_FORMAT)
                ts_ok[TIME_FORMAT] = time_format

            if TUNIT not in ts_ok and TIME_FORMAT not in ts_ok:
                raise ConfException("Missed both {} and {} configurtion.".format(TUNIT, TIME_FORMAT))
            data[TS] = ts_ok

    def _ck_type(self, val, ty, extra_str):
        if not isinstance(val, ty):
            raise ConfException("`{}` expected `{}` configuration in YAML file.".format(extra_str, ty.__name__))

    def _get_ck(self, key, yaml, ty, extra_str):
        if key not in yaml:
            raise ConfException("Missed `{}` configuration in YAML file.".format(extra_str))
        val = yaml[key]
        self._ck_type(val, ty, extra_str)
        return val

    def _read_yaml(self):
        yaml_data = None
        with open(self.yaml_path, 'r', encoding="utf-8") as f:
            file_data = f.read()
            yaml_data = yaml.load(file_data, Loader=yaml.FullLoader)
        return yaml_data

# p = YamlParser("./config.yaml")
# d = p.parse()
# print(d)
