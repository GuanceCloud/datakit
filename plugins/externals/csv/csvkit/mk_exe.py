# -*- encoding: utf8 -*-

import os
cmd = "pyinstaller --onefile -F --clean main.py -n csvkit"
os.system(cmd)
