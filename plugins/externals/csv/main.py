# -*- encoding: utf8 -*-
import base64
import logging

from csvkit.network import set_http_addr
from csvkit.utils import exit, parse_cli
from csvkit.metric import collect_metric
from csvkit.object import collect_object
from csvkit.log import en_logging
from csvkit.toml_parse import parse_toml_cfg

args = parse_cli()
en_logging(args.log_file, args.log_level)

logging.info("http addr: {}".format(args.http))
set_http_addr(args.http)

if args.metric == "" and args.object == "":
    logging.critical("both metric and object cfg info missed")
    exit(0)

if args.metric != "":
    metric_cfg = base64.standard_b64decode(args.metric).decode()
    logging.info("toml cfg info about metric {}".format(metric_cfg))

    parsed_cfg = parse_toml_cfg(metric_cfg)
    logging.info("toml parsed cfg info about metric {}".format(parsed_cfg))

    collect_metric(parsed_cfg)

if args.object != "":
    object_cfg = base64.standard_b64decode(args.object).decode()
    logging.info("toml cfg info about metric {}".format(object_cfg))
    
    parsed_cfg = parse_toml_cfg(object_cfg)
    logging.info("toml parsed cfg info about object {}".format(parsed_cfg))
    collect_object(parsed_cfg)