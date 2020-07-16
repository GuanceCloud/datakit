# -*- encoding: utf8 -*-

import threading

log_handler = None
log_console = False

mutex = threading.Lock()

def log_file_init(file_path):
    global log_handler
    log_handler = open(file_path, "w")

def log_console_init(is_enable):
    global log_console
    log_console = is_enable

def log(line_data):
    with mutex:
        if log_console:
            print(line_data, end = '')

        if log_handler:
            log_handler.write(line_data)

def log_flush():
    if log_handler:
        log_handler.close()
