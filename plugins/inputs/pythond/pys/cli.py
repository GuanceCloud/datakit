#encoding: utf-8

import os
import sys
import time
import importlib
import threading
import argparse
import logging
from logging.handlers import RotatingFileHandler

import sys
sys.path.append(${PythonCorePath})
sys.path.extend(${CustomerDefinedScriptRoot})

from datakit_framework import DataKitFramework

PY2 = sys.version_info[0] == 2
PY3 = sys.version_info[0] == 3

logger = logging.getLogger('pythond_cli')

def init_log():
    log_path = os.path.join(os.path.expanduser('~'), "_datakit_pythond_cli.log")
    print(log_path)
    logger.setLevel(logging.DEBUG)
    handler = RotatingFileHandler(log_path, maxBytes=100000, backupCount=10)
    logger.addHandler(handler)

def mylog(msg, *args, **kwargs):
    logger.debug(time.strftime("%Y-%m-%d %H:%M:%S ", time.localtime()) + msg, *args, **kwargs)

class RunThread (threading.Thread):
	__plugin = DataKitFramework()
	__interval = 10

	def __init__(self, plugin):
		threading.Thread.__init__(self)
		self.__plugin = plugin
		if self.__plugin.interval:
			self.__interval = self.__plugin.interval

	def run(self):
		if self.__plugin:
			while True:
				try:
					self.__plugin.run()
				except:
					mylog("Unexpected error: info = %s, script = '%s'", sys.exc_info(), self.__plugin.name)
				time.sleep(self.__interval)

def search_plugin(plugin_path):
	try:
		mod = importlib.import_module(plugin_path)
	except ModuleNotFoundError:
		mylog(plugin_path + " not found.")
		return

	plugins = []

	for _, v in mod.__dict__.items():
		if v is not DataKitFramework and type(v).__name__ == 'type' and issubclass(v, DataKitFramework):
			plugin = v()
			# return plugin
			plugins.append(plugin)

	return plugins

def main(*args):
    plugins = []
    threads = []

    for arg in args:
        plg = search_plugin(arg)
        if plg and len(plg) > 0:
            plugins.extend(plg)

    for plg in plugins:
        thd = RunThread(plg)
        thd.start()
        threads.append(thd)

    for t in threads:
        t.join()

if __name__ == '__main__':
	parser = argparse.ArgumentParser(description="datakit framework")
	parser.add_argument('--logname', '-l', help='indicates datakit framework log name, required')
	args = parser.parse_args()
	if args.logname:
		DataKitFramework.log_name = args.logname
	else:
		print("need logname")
		sys.exit(-1)

	init_log()
	main(${CustomerDefinedScriptName})