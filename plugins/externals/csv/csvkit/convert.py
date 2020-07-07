# -*- encoding: utf8 -*-

import threading
import xlwt
import csv

DEFAULT_SHEET_NAME = "Sheet1"

class Csv2Excel:
    def __init__(self, *conv_files):
        self.conv_files = conv_files

    def convert(self):
        tasks =[]
        for file_path, conv_path in self.conv_files:
            t = threading.Thread(target=self._conv_task, args=(file_path, conv_path))
            tasks.append(t)
            t.start()

        for t in tasks:
            t.join()

    def _conv_task(self, file_path, conv_path):
        with open(file_path) as f:
            book = xlwt.Workbook(encoding='utf-8')
            sheet = book.add_sheet(DEFAULT_SHEET_NAME)
            f_csv = csv.reader(f)

            for r, row in enumerate(f_csv):
                for c, column in enumerate(row):
                    sheet.write(r, c, column)

            book.save(conv_path)
