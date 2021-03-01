import yaml
import logging
import os
import es_index
import traceback


class Cfg(object):
    def __init__(self, path):
        self.cfg = self.init_cfg(path)
        self.log = self.init_log()

    def init_cfg(self, path):
        config = {}
        if os.path.exists(path):
            with open(path) as f:
                config = yaml.safe_load(f)

        return config

    def init_log(self):
        level_relations = {
            'debug': logging.DEBUG,
            'info': logging.INFO,
            'warning': logging.WARNING,
            'error': logging.ERROR,
            'crit': logging.CRITICAL
        }
        logging.basicConfig(
            level=level_relations.get(self.cfg.get("log").get("level"), "debug"),
            filename=self.cfg.get("log").get("path"),
            format='%(asctime)s - %(pathname)s[line:%(lineno)d] - %(levelname)s: %(message)s'
        )
        log = logging.getLogger("quan_yuan_tang")
        return log


if __name__ == '__main__':
    c = Cfg("config.yaml")
    run_dict = {
        "es_index": es_index.run,
    }
    for k, v in run_dict.items():
        try:
            v(c)
        except Exception as e:
            c.log.error(traceback.format_exc())
