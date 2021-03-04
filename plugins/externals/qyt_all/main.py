#! /usr/bin/env python3
from gevent import monkey

monkey.patch_all()
import toml
import logging
import os
import es_index
import traceback
import sys
import gevent
import qyt_mysql
import qyt_oracle


class Cfg(object):
    def __init__(self, path):
        self.cfg = self.init_cfg(path)
        self.log = self.init_log()

    def init_cfg(self, path):
        config = {}
        if os.path.exists(path):
            with open(path) as f:
                config = toml.load(f)

        return config.get("input", {}).get("quanyuantang", {})

    def init_log(self):
        level_relations = {
            'debug': logging.DEBUG,
            'info': logging.INFO,
            'warning': logging.WARNING,
            'error': logging.ERROR,
            'crit': logging.CRITICAL
        }
        logging.basicConfig(
            level=level_relations.get(self.cfg.get("log_level"), "debug"),
            filename=self.cfg.get("log_path"),
            format='%(asctime)s - %(pathname)s[line:%(lineno)d] - %(levelname)s: %(message)s'
        )
        log = logging.getLogger("quan_yuan_tang")
        return log


def run(func, c):
    try:
        func(c)
    except Exception:
        c.log.error(traceback.format_exc())


if __name__ == '__main__':
    _, path = sys.argv
    c = Cfg(path)
    run_dict = {
        "es_index": es_index.run,
        "oracle": qyt_oracle.run,
        "mysql": qyt_mysql.run
    }
    func_list = []
    for k, _ in c.cfg.items():
        if k not in run_dict:
            continue
        func_list.append(gevent.spawn(run, run_dict[k], c))
    gevent.joinall(func_list)
