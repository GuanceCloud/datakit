# -*- encoding: utf8 -*-
START = "StartRow"
BATCH = "BatchSize"

UUID  = "DatawayUuid"
AUTH  = "DatawayAuth"
URL   = "DatawayUrl"
PK    = "Ak"
SK    = "Sk"

FILES = "Files"

RULES = "Rules"

PRMK  = "PrimaryKey"

MEMENT = "Measurement"
TAG    = "Tags"
FIELD  = "Fields"
TS     = "Timestamp"

COLUMN  = "ColumnIndex"
CELL    = "Cell"
TYPE    = "Type"
NAME    = "Name"
NACTION = "NaAction"
TUNIT   = "TimeUnit"
TIME_FORMAT = "TimeFormat"



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

