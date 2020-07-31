# -*- encoding: utf8 -*-
START = "start_rows"
FILE = "file"
RULES = "Rules"
PRMK  = "PrimaryKey"

MEMENT = "metric"
TAG    = "Tags"
AS_TAG = "as_tag"
AS_FIELD = "as_field"
AS_TIME  = "as_time"
FIELD  = "Fields"
TS     = "Timestamp"

COLUMN  = "columns"
INDEX  = "index"
CELL    = "Cell"
TYPE    = "type"
NAME    = "name"
NACTION = "na_action"
TUNIT   = "time_precision"
TIME_FORMAT = "time_format"



IGNORE = "ignore"
DROP   = "drop"
ABORT  = "abort"
NaAction = [IGNORE, DROP, ABORT]

INT   = "int"
STR   = "str"
BOOL  = "bool"
FLOAT = "float"
FieldType =  [INT, STR, BOOL, FLOAT]
TsUnit = ["s", "ms", "us", "ns"]
