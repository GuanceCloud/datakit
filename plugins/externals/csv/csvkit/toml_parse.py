import toml
import logging
from csvkit.utils import exit

def parse_toml_cfg(cfg_str):
    try:
        return toml.loads(cfg_str)
    except Exception as e:
        logging.critical("parse toml exception {}".format(e))
        exit(0)

