# -*- encoding: utf8 -*-

import sys
from csvkit.log import log_flush

def is_http_url(url_path):
    if not isinstance(url_path, str):
        return False

    url_path = url_path.strip()
    if url_path.startswith("http://") or url_path.startswith("https://"):
        return True

    return False

def exit(code):
    log_flush()
    sys.exit(code)
