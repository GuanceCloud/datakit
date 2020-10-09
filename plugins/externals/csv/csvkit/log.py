# -*- encoding: utf8 -*-
import logging

def en_logging(log_file, log_level):
    level = 0
    if log_level == "debug":
        level = logging.DEBUG
    elif log_level == "info":
        level = logging.INFO
    elif log_level == "warn":
        level = logging.WARN
    elif log_level == "error":
        level = logging.ERROR
    elif log_level == "fatal":
        level = logging.FATAL
    else:
        level = logging.INFO
    logging.basicConfig(filename=log_file, format='%(asctime)s:%(filename)s:%(lineno)d:%(levelname)s: %(message)s',
                        filemode='w', level=level)
