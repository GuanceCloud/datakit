#!/usr/bin/env python
# -*- encoding: utf8 -*-
#
# date: Wed Apr  1 08:50:58 UTC 2020
#

import unittest
import requests
import hmac
import base64
import time
import hashlib
import json
import toml
import argparse
import sys

import nsq
from email.utils import formatdate

import os
from email.utils import formatdate

api_host = os.getenv("KODO_HOST")
db_uuid = os.getenv("DB_UUID")
measurements=os.getenv("MEASUREMENTS").split(',')

class KodoAPICases(unittest.TestCase):
    Args = {}

    def test_sub_keyevent(self):
        def handler(msg):
            x = json.loads(msg.body)
            print(json.dumps(x, indent=4, sort_keys=True))
            return True

        r = nsq.Reader(message_handler=handler,
                lookupd_http_addresses=[nsq_lookupd],
                topic='dataflux-point-check',
                channel="dataflux-point-check-channel",
                lookupd_poll_interval=1)
        nsq.run()

    def test_read_metrics(self):

        commands = []
        # for m in measurements:
        #     #commands.append('select * from "%s" order by time desc limit 300' % m)
        #     commands.append('select mean(*) from "%s"' % m)

        commands.append("""
                        SELECT last("出院占比")/3,last("出院占比") FROM biz_1354e40a0b3642488a8c939aefb141f9.."中国新冠舆情" WHERE ((time >= 1587109620244ms AND time <= 1587110520244ms)) AND ((time >= 1587109620244ms AND time <= 1587110520244ms)) GROUP BY time(3s) fill(none) TZ('Asia/Shanghai')
		""")
        j = {
                "db_uuid": "ifdb_294ea76aaa5444a9a8cea43688f2422f",
                "commands": commands
                }
        resp = self.do_api(uri='/v1/read/metrics', jsn=j, method='POST', expect_code=200)
        x = json.loads(resp)
        print(json.dumps(x, indent=4, sort_keys=True))
        print(';'.join(commands))

    def test_create_db(self):
        resp = self.do_api(uri='/v1/influx/add_db?default_rp=rp0&cqrp=autogen&rp_list=rp0,rp1,rp2,rp3', method='POST', expect_code=200, pprint=True)
        x = json.loads(resp)
        print(json.dumps(x, indent=4, sort_keys=True))

    def test_rewrite_sql(self):
        j = {"db_uuid": "ifdb_4948f4d06bdb47218b3138bc4f55ed4c",
        "sqls": [{
                "sql":
                """
        
        select a from b where 1>0 group by time(1m); -- bad select
		select a from b where 1>0 limit 10;
		select MEAN(a) from b where 1>0 group by tag_a;
		select MEAN(a) from b where 1>0 group by tag_a, time(1m);

		select LAST(a) from (select a from b where 1>0) where 1>0;
		select x from (select a from b where 1>0) where 1>0 limit 10;

		select x /* this is comment */ from (select a from b where 1>0) where 1>0;
                """,
                "conditions": "a>0 and b<0",
                "measurements": ["a"],
                "dimensions": ["time(5m)"],
                "default_aggr_func": "LAST",
                "max_point": 360,
                }]
        }
        
        j = {
            "db_uuid":"ifdb_78344286d4a04c729464f564be5bb84f",
            "sqls":
                [
                    {
                       "conditions":"(\"$level\" != 'ok') and (\"time\" >= 1588048020654ms and \"time\" <= 1588051650654ms)",
                        #"default_aggr_func":"Last",
                        "dimensions":["time(1h)"],
                        "fill_opt":"none",
                        "max_point":366,
                        "measurements":[],
                        "sql": "select count(\"$ruleName\") from \"$alert\" where (\"$level\" != 'ok') and (\"time\" >= 1588048020654ms and \"time\" <= 1588051650654ms) order by  \"time\" desc",
                        #"sql":"select last(\"\u4eba\u6570\") from \"\u5728\u7ebf\u4e0a\u8bfe\u4eba\u6570\" where (\"time\" >= 1586509200000ms and \"time\" <= 1587115800000ms) group by time(undefined), \"\u5e74\u7ea7\" fill(none)  tz('Europe/London')",
                        "time_range":[1586509200000,1588038800000]
                    }]}


        resp = self.do_api(uri='/v1/rewrite', jsn=j, method="POST", expect_code=200, pprint=True)
        x = json.loads(resp)
        print(json.dumps(x, indent=4, sort_keys=True))

    def test_add_db(self):
        resp = self.do_api(uri='/v1/influx/add_db?default_rp=rp3&cqrp=', method='POST', expect_code=200, pprint=True)
        print(resp)

    def test_add_cqdb(self):
        resp = self.do_api(uri='/v1/cq/create_db?db_uuid=ifdb_4948f4d06bdb47218b3138bc4f55ed4c', method='POST', expect_code=200, pprint=True)
        print(resp)
        

    def test_drop_meas(self):
        j = {
                "db": db_uuid,
                "command": """drop measurement mock_cpu_2"""
                }
        resp = self.do_api(uri='/v1/drop/metrics', jsn=j, method='POST', expect_code=200, pprint=True)
        print(resp)

    #def test_read_metrics(self):
    #    j = {
    #            "db": db_uuid,
    #            #"command": "SHOW MEASUREMENTS limit 100  offset 1"
    #            "command": "select * from processes order by time limit 3"
    #            }
    #    resp = self.do_api(uri='/v1/read/metrics', jsn=j, method='POST', expect_code=200, pprint=True)
    #    print(resp)


    def test_get_meas_rp(self):
        j = {
                "db_uuid": "ifdb_5708c9c73eb04c79a31861d4b5990ba4",
                "measurements": [
                   "kodo_slow_query",
                   
                    ]
                }

        resp = self.do_api(uri='/v1/measurement/get_rp', jsn=j, method='POST', expect_code=200, pprint=True)

        j = {
                "db": "ifdb_not-exists",
                "measurements": [
                    "kodo_slow_query",
                    "mock_cpu",
                    "dataway_self",
                    "field-not-exists",
                    ]
                }

        resp = self.do_api(uri='/v1/measurement/get_rp', jsn=j, method='POST', expect_code=404, pprint=True)

    def test_modify_rp(self):
        j = {
                "db_uuid": db_uuid,
                "default_rp": """rp5""",
                "rp_list": ["rp3","rp4","rp5"]
                }
        resp = self.do_api(uri='v1/modify/dbrp', jsn=j, method='POST', expect_code=200, pprint=True)
        print(resp)
        
    def test_create_cq(self):
        j = {
            "db_uuid": "ifdb_5708c9c73eb04c79a31861d4b5990ba4", 
            "workspace_uuid": "wksp_e2db2d02837a11eaba9a8671df186910", 
            "measurements": [{"measurement":"Nike","aggr_period":"100m","aggr_func":"mean","aggr_every":"1m"}]
        }
        resp = self.do_api(uri='v1/cq/create', jsn=j, method='POST', expect_code=200, pprint=True)
        print(resp)

    def test_modify_cq(self):
        j = {      
            "cq_uuid": "cq_91f8e99690924d04af9796a35d6c32eb",     
            "workspace_uuid": "wksp_e2db2d02837a11eaba9a8671df186910", 
            "aggr_func": "sum",
            "aggr_period": "1m",
            "aggr_every": "5m"
        }
        resp = self.do_api(uri='v1/cq/modify', jsn=j, method='POST', expect_code=200, pprint=True)
        print(resp)
        
    def test_delete_cq(self):
        j = {           
            "cq_uuid": "cq_91f8e99690924d04af9796a35d6c32eb",      
            "workspace_uuid": "wksp_e2db2d02837a11eaba9a8671df186910"
        }
        resp = self.do_api(uri='v1/cq/delete', jsn=j, method='POST', expect_code=200, pprint=True)
        print(resp)
        
    def test_syncupdate_cq(self):
        resp = self.do_api(uri='v1/cq/syncupdate?workspace_uuid=wksp_0bc86b4627c111ea8dd02e762ce1615d', method='POST', expect_code=200, pprint=True)
        print(resp)
        
    def test_drop_ck_metric(self):
        j = {           
            "db_uuid": "ifdb_4948f4d06bdb47218b3138bc4f55ed4c",      
            "table_name": "$alert"
        }
        resp = self.do_api(uri='v1/ck/drop/metrics', jsn=j, method='POST', expect_code=200, pprint=True)
        print(resp)
        
    def test_drop_metric(self):
        j = {           
            "db_uuid": "ifdb_b20b58a5f85f4e4ab54a565efa19bb8e",      
            "measurement": "战斗基数1"
        }
        resp = self.do_api(uri='v1/drop/metrics', jsn=j, method='POST', expect_code=200, pprint=True)
        print(resp)
        
    def test_ck_status(self):
        resp = self.do_api(uri='v1/ck/status', method='GET', expect_code=200, pprint=True)
        print(resp)
        
    def test_ck_createdb(self):
        resp = self.do_api(uri='v1/ck/create_db?db_uuid=joly_test', method='POST', expect_code=200, pprint=True)
        print(resp)
        
    def test_list_db_tscnt(self):
        j = [
                'ifdb_1337970c3b2148c0a56ae10208499ccd',
                'ifdb_1548eed998314bf9a3beff4d68aa1c63',
                'ifdb_1779891eb1364f96a0f058758342e604',
                'ifdb_179bc0a8854e49c7815a399f26ee066e',
                'ifdb_1a0439a2699c454f9ab1827cb3d32283',
                'ifdb_1bc7ef27031944eda40711a55c7d1c84',
                'ifdb_1be3970954ab40f4b50c0cd51409838d',
                'ifdb_1cd9d46572df4f0f91a4dda741017e9a',
                'ifdb_20c8df9827bd4237a19b2b3bbfb6d3d8',
                'ifdb_245bfb992a87471e8b7e271d9d3d3dec',
                'ifdb_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx',
                ]
        resp = self.do_api(uri='/v1/tscnt', jsn=j, method='GET', expect_code=200, pprint=True)
        print(resp)

    def test_clean_dataway_cache(self):
        resp = self.do_api(uri='/v1/cache/dataway/clean?token=tkn_xxxxxxxxxxxxxxx', method='POST', expect_code=200, pprint=True)
        print(resp)

    def test_check_token(self):
        resp = self.do_api(uri='/v1/check/token/tkn_xxxxxxxxxxxxxxx', method='GET', expect_code=200, pprint=True)
        print(resp)
       
    def test_influxql_parse(self):

        j = [
                #''' show tag keys on "db" from "meas-1" where tag_val='abc' and time>now()-1d limit 1 ''',
                #''' show tag keys on "db" from "meas-1" ''',
                #''' SELECT /a/ from abc where a=~/y/ ''',
                #''' SHOW TAG VALUES ON "NOAA_water_database" WITH KEY = "randtag" ''',
                #''' SHOW FIELD KEYS ON "NOAA_water_database" FROM "h2o_feet" ''',
                #''' SHOW MEASUREMENTS ON NOAA_water_database ''',

                #''' SELECT * from xxx where time > now()- 1s''',
                '''
                SELECT MEAN("spills_per_person") AS m_spp FROM
    (SELECT "spilled_coffee"/"passengers" AS "spills_per_person" FROM "schedule"
            WHERE a>0 OR (b>0 AND c<0) GROUP BY "subway")
    WHERE m_spp > 0 LIMIT 42 '''
            ]

        resp = self.do_api(uri='/v1/parse', jsn=j, method='POST', expect_code=200) # get json license response
        x = json.loads(resp)
        print(json.dumps(x, indent=4, sort_keys=True))

    ##################################################################################################################
    #  Helpers
    ##################################################################################################################
    def setUp(self):
        pass

    def tearDown(self):
        pass

    @classmethod
    def cases(cls):
        return sorted([i for i in cls.__dict__.keys() if i.startswith('test_')])

    def do_api(self, uri=None, jsn=None, data=None, method=None, headers=None, expect_code=200, pprint=False):
        if headers is None:
            headers = {}

        f  = getattr(requests, method.lower())
        self.assertTrue(callable(f))

        body = ''
        ctype = ''
        if jsn is not None:
            body = json.dumps(jsn)
            ctype = 'application/json'
        elif data is not None:
            body = data
            ctype = 'plain/text'

        r = f(api_host + uri, json=jsn, data=data, headers=headers)

        if expect_code is not None and expect_code != r.status_code:
            pretty_print(r)
            self.assertTrue(False)
        else:
            if pprint:
                pretty_print(r)

        return r.text.encode('utf8')

def pretty_print(resp):
    req = resp.request
    print('\n%s\n%s\n%s\n\n%s' % (
        '>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>',
        req.method + ' ' + req.url,
        '\n'.join('{}: {}'.format(k, v) for k, v in req.headers.items()),
        req.body or ""))

    print('\n%s\nHTTP/1.1 %d %s\n%s\n\n%s' % (
        '<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<',
        resp.status_code, resp.reason,
        '\n'.join('{}: {}'.format(k, v) for k, v in resp.headers.items()),
        resp.text.encode('utf8')))

if __name__ == '__main__':

    psr = argparse.ArgumentParser(description="kodo API test cases")

    psr.add_argument('--case', action='store', dest='case', help='select case function name to test', default=None)
    res = psr.parse_args(sys.argv[1:])
    KodoAPICases.Args = res

    ts = unittest.TestSuite()

    print('testing case %s' % KodoAPICases.Args.case)
    if KodoAPICases.Args.case == 'all':
        for c in KodoAPICases.cases():
            ts.addTest(KodoAPICases(c))
    else:
        if not hasattr(KodoAPICases, KodoAPICases.Args.case):
            print("case `%s' not found")
            sys.exit(-1)

        ts.addTest(KodoAPICases(KodoAPICases.Args.case))

    unittest.TextTestRunner(verbosity=3).run(ts)
