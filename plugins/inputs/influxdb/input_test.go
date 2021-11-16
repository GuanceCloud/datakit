package influxdb

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func TestParseData(t *testing.T) {
	// 指定指标名的映射 map
	if fc, err := DebugVarsDataParse2Point([]byte(dataInfluxDebugVars1v8),
		MetricMap); err != nil {
		t.Error("parse failed", err)
	} else {
		for {
			_, err := fc()
			if err != nil {
				if reflect.TypeOf(err) == reflect.TypeOf(NoMoreDataError{}) || err.Error() == "no more data" {
					break
				} else {
					t.Error(err)
				}
			}
		}
	}

	// 不指定指标名的映射 map
	if fc, err := DebugVarsDataParse2Point([]byte(dataInfluxDebugVars1v8),
		nil); err != nil {
		t.Error("parse failed", err)
	} else {
		for {
			_, err := fc()
			if err != nil {
				if reflect.TypeOf(err) == reflect.TypeOf(NoMoreDataError{}) || err.Error() == "no more data" {
					break
				} else {
					t.Error(err)
				}
			}
		}
	}
}

func TestCollect(t *testing.T) {
	i := Input{
		Interval: datakit.Duration{Duration: time.Second * 15},
		Timeout:  datakit.Duration{Duration: time.Second * 5},
	}
	i.client = &http.Client{
		Transport: &http.Transport{
			ResponseHeaderTimeout: i.Timeout.Duration,
			TLSClientConfig:       nil,
		},
		Timeout: i.Timeout.Duration,
	}
	i.URL = "http://localhost:8086/debug/vars"
	if err := i.Collect(); err != nil {
		i.collectCache = make([]inputs.Measurement, 0)
	}
}

const (
	dataInfluxDebugVars1v8 = `
	{
		"system": {
			"currentTime": "2021-07-13T03:08:55.252414311Z",
			"started": "2021-07-12T05:17:28.227881632Z",
			"uptime": 78687
		},
		"cmdline": [
			"influxd"
		],
		"memstats": {
			"Alloc": 11126424,
			"TotalAlloc": 974101960,
			"Sys": 73141649,
			"Lookups": 0,
			"Mallocs": 4265076,
			"Frees": 4152542,
			"HeapAlloc": 11126424,
			"HeapSys": 65961984,
			"HeapIdle": 53043200,
			"HeapInuse": 12918784,
			"HeapReleased": 52584448,
			"HeapObjects": 112534,
			"StackInuse": 1146880,
			"StackSys": 1146880,
			"MSpanInuse": 150280,
			"MSpanSys": 196608,
			"MCacheInuse": 13888,
			"MCacheSys": 16384,
			"BuckHashSys": 1581902,
			"GCSys": 2394112,
			"OtherSys": 1843779,
			"NextGC": 11907584,
			"LastGC": 1626145619238680929,
			"PauseTotalNs": 14059643,
			"PauseNs": [
				12175,
				15996,
				8669,
				12540,
				25868,
				12174,
				12902,
				161358,
				94636,
				53814,
				25549,
				12302,
				90496,
				97312,
				48615,
				23469,
				17439,
				93169,
				32004,
				12966,
				29573,
				100305,
				114766,
				9243,
				27447,
				27841,
				227238,
				50900,
				56619,
				32721,
				137615,
				16415,
				86112,
				27445,
				9666,
				33969,
				35644,
				27049,
				32744,
				26838,
				125049,
				31495,
				176623,
				16874,
				62656,
				18205,
				78449,
				146675,
				18374,
				81630,
				29969,
				7596,
				85357,
				41355,
				19159,
				31921,
				13388,
				36018,
				27206,
				13264,
				45047,
				25588,
				167160,
				79779,
				10859,
				44037,
				15381,
				56509,
				59434,
				63946,
				75072,
				89643,
				22605,
				191640,
				517005,
				31937,
				105093,
				69006,
				80129,
				39060,
				78591,
				33451,
				35573,
				80410,
				35817,
				28114,
				59566,
				89549,
				26113,
				25296,
				65479,
				67103,
				86984,
				89272,
				79978,
				100146,
				95011,
				29270,
				57350,
				56885,
				59346,
				73024,
				68390,
				69394,
				33905,
				23744,
				53360,
				13753,
				33495,
				124722,
				96065,
				34164,
				114122,
				116093,
				180000,
				104283,
				211286,
				85354,
				123573,
				58303,
				99894,
				61321,
				111761,
				36321,
				78592,
				16415,
				30437,
				69201,
				92992,
				84172,
				193094,
				108393,
				53577,
				81452,
				143922,
				29389,
				59965,
				13664,
				100105,
				98874,
				40106,
				66300,
				98596,
				31536,
				78545,
				80683,
				13762,
				64755,
				83528,
				54436,
				25221,
				86375,
				54126,
				59191,
				53846,
				118252,
				57441,
				111813,
				84343,
				103766,
				68545,
				169663,
				61336,
				80148,
				31040,
				77531,
				30053,
				27923,
				63546,
				66976,
				26686,
				60706,
				88693,
				66508,
				39398,
				100940,
				43476,
				118544,
				75158,
				271588,
				85699,
				18323,
				58288,
				36523,
				69015,
				87525,
				37926,
				74294,
				51832,
				17526,
				26510,
				36122,
				27220,
				34698,
				26770,
				54981,
				25844,
				28088,
				46435,
				32009,
				22549,
				60431,
				7753,
				40070,
				40961,
				27385,
				32410,
				29086,
				26434,
				31478,
				77697,
				11718,
				60034,
				23574,
				18946,
				23058,
				30563,
				28075,
				29044,
				40041,
				38223,
				22917,
				31341,
				14314,
				14436,
				27301,
				86047,
				69379,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0
			],
			"PauseEnd": [
				1626067048337043591,
				1626067048650760625,
				1626067048688934032,
				1626067048711473259,
				1626067048726314272,
				1626067048735273029,
				1626067060009910875,
				1626067130007282185,
				1626067200009494455,
				1626067265012800960,
				1626067331012542751,
				1626067400002610736,
				1626067480009166458,
				1626067560007024931,
				1626067633493343476,
				1626067700009614867,
				1626067770011075039,
				1626067850008038519,
				1626067930002866954,
				1626068013509697369,
				1626068090002535082,
				1626068170006401485,
				1626068260004618098,
				1626068340002720885,
				1626068420008099918,
				1626068510009122055,
				1626068600007356645,
				1626068690009680976,
				1626068780004432151,
				1626068870007961742,
				1626068960011320381,
				1626069060004800342,
				1626069150009310168,
				1626069240010226774,
				1626069322302124507,
				1626069390010570466,
				1626069483612754725,
				1626069570010446065,
				1626069640010569959,
				1626069748340814841,
				1626069850008735620,
				1626069960008776605,
				1626070070007215417,
				1626070168339357701,
				1626070270007375096,
				1626070371010615446,
				1626070480003302750,
				1626070590009801407,
				1626070693673800543,
				1626070800010807994,
				1626070910009665176,
				1626071020002873102,
				1626071130006342722,
				1626071240003429843,
				1626071350005436966,
				1626071460009398421,
				1626071573724489232,
				1626071682014672025,
				1626071790012122454,
				1626071900005648047,
				1626072013750658741,
				1626072130009624208,
				1626072200013480642,
				1626072320561072598,
				1626072440578102533,
				1626072560608237369,
				1626072680633841689,
				1626072800664506396,
				1626072920692957055,
				1626073040723644459,
				1626073160753226064,
				1626073280777244151,
				1626073400800038096,
				1626073520830453960,
				1626073640856261178,
				1626073760876180691,
				1626073880908702319,
				1626074000934195842,
				1626074120967518599,
				1626074240997664108,
				1626074361016461899,
				1626074481044903669,
				1626074601070907120,
				1626074721098348914,
				1626074841121436152,
				1626074961146896490,
				1626075081177472652,
				1626075201212130668,
				1626075321237701776,
				1626075441258363102,
				1626075561287589545,
				1626075681313631804,
				1626075801340394548,
				1626075921369774020,
				1626076041402621589,
				1626076161426975098,
				1626076281451918354,
				1626076401474825166,
				1626076521503119618,
				1626076641534784659,
				1626076761558789449,
				1626076881594341548,
				1626077001614049617,
				1626077121651873268,
				1626077241681902158,
				1626077340014860305,
				1626077460729543089,
				1626077580756509605,
				1626077700791971533,
				1626077820813234827,
				1626077940848512905,
				1626078060875901116,
				1626078180894908836,
				1626078300921616738,
				1626078420944894351,
				1626078540975493567,
				1626078660994370331,
				1626078781016955679,
				1626078901054038510,
				1626079021068903224,
				1626079141092344635,
				1626079261129047562,
				1626079381150970601,
				1626079501170845502,
				1626079621203593711,
				1626079741239213530,
				1626079861275030507,
				1626079981290296574,
				1626080101325723270,
				1626080221360215467,
				1626080341389381170,
				1626080461416378226,
				1626080581437697080,
				1626080701478181279,
				1626080821500520846,
				1626080941526128568,
				1626081061547480788,
				1626081181576462065,
				1626081301615331315,
				1626081421639693678,
				1626081541652225027,
				1626081661694582331,
				1626081781723293704,
				1626081901746910306,
				1626082021760341610,
				1626082141797555296,
				1626082261806654075,
				1626082381844634518,
				1626082501869830107,
				1626082621902704074,
				1626082741935312488,
				1626082861960808357,
				1626082981981514504,
				1626083102006002912,
				1626083222018469501,
				1626083342067871974,
				1626083462081453257,
				1626083582118875981,
				1626083702140869255,
				1626083822161465483,
				1626083942199704523,
				1626084062218303557,
				1626084182242486377,
				1626084302266891563,
				1626084422304940454,
				1626084542323066328,
				1626084662353614193,
				1626084782378805150,
				1626084902411494085,
				1626085022439799948,
				1626085142450840130,
				1626085262490275935,
				1626085382508288006,
				1626085502546249881,
				1626085622555088829,
				1626085742598234981,
				1626085862619395268,
				1626085982651033534,
				1626086102672447407,
				1626086222694360284,
				1626086342717800417,
				1626086462737690550,
				1626140941652799370,
				1626141062205407724,
				1626141182216790934,
				1626141302258975145,
				1626141422270712071,
				1626141542298396791,
				1626141608645571724,
				1626141691257841532,
				1626141768647491473,
				1626141848641749920,
				1626141928644096952,
				1626142008647615493,
				1626142098641078503,
				1626142188643749509,
				1626142258644299560,
				1626142349260322944,
				1626142438640312288,
				1626142528645918818,
				1626142618641306866,
				1626142718644126461,
				1626142808639887719,
				1626142908640530468,
				1626143008644581829,
				1626143108644824725,
				1626143208646517558,
				1626143308646014200,
				1626143408644762961,
				1626143488646770125,
				1626143588645548383,
				1626143698641191287,
				1626143818647521507,
				1626143938647347703,
				1626144058641272661,
				1626144178648180171,
				1626144298645623422,
				1626144418646720087,
				1626144538646430003,
				1626144658646694687,
				1626144778647479009,
				1626144898646055499,
				1626145018646273392,
				1626145138641972019,
				1626145258641939175,
				1626145378643313819,
				1626145499211027940,
				1626145619238680929,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0
			],
			"NumGC": 228,
			"NumForcedGC": 0,
			"GCCPUFraction": 0.00001462010751575198,
			"EnableGC": true,
			"DebugGC": false,
			"BySize": [
				{
					"Size": 0,
					"Mallocs": 0,
					"Frees": 0
				},
				{
					"Size": 8,
					"Mallocs": 21429,
					"Frees": 20642
				},
				{
					"Size": 16,
					"Mallocs": 1714612,
					"Frees": 1623053
				},
				{
					"Size": 32,
					"Mallocs": 501900,
					"Frees": 496250
				},
				{
					"Size": 48,
					"Mallocs": 496796,
					"Frees": 492965
				},
				{
					"Size": 64,
					"Mallocs": 292601,
					"Frees": 290525
				},
				{
					"Size": 80,
					"Mallocs": 45781,
					"Frees": 45414
				},
				{
					"Size": 96,
					"Mallocs": 74342,
					"Frees": 73657
				},
				{
					"Size": 112,
					"Mallocs": 23782,
					"Frees": 23575
				},
				{
					"Size": 128,
					"Mallocs": 44196,
					"Frees": 43931
				},
				{
					"Size": 144,
					"Mallocs": 8499,
					"Frees": 8355
				},
				{
					"Size": 160,
					"Mallocs": 9085,
					"Frees": 8754
				},
				{
					"Size": 176,
					"Mallocs": 10850,
					"Frees": 10711
				},
				{
					"Size": 192,
					"Mallocs": 12488,
					"Frees": 12452
				},
				{
					"Size": 208,
					"Mallocs": 27973,
					"Frees": 27699
				},
				{
					"Size": 224,
					"Mallocs": 161435,
					"Frees": 160134
				},
				{
					"Size": 240,
					"Mallocs": 220711,
					"Frees": 219028
				},
				{
					"Size": 256,
					"Mallocs": 22222,
					"Frees": 22062
				},
				{
					"Size": 288,
					"Mallocs": 94734,
					"Frees": 93827
				},
				{
					"Size": 320,
					"Mallocs": 2212,
					"Frees": 2055
				},
				{
					"Size": 352,
					"Mallocs": 26768,
					"Frees": 26720
				},
				{
					"Size": 384,
					"Mallocs": 37507,
					"Frees": 36956
				},
				{
					"Size": 416,
					"Mallocs": 2970,
					"Frees": 2902
				},
				{
					"Size": 448,
					"Mallocs": 6351,
					"Frees": 6327
				},
				{
					"Size": 480,
					"Mallocs": 2922,
					"Frees": 2898
				},
				{
					"Size": 512,
					"Mallocs": 9432,
					"Frees": 9382
				},
				{
					"Size": 576,
					"Mallocs": 27841,
					"Frees": 27657
				},
				{
					"Size": 640,
					"Mallocs": 1504,
					"Frees": 1475
				},
				{
					"Size": 704,
					"Mallocs": 3663,
					"Frees": 3643
				},
				{
					"Size": 768,
					"Mallocs": 1176,
					"Frees": 1162
				},
				{
					"Size": 896,
					"Mallocs": 1636,
					"Frees": 1577
				},
				{
					"Size": 1024,
					"Mallocs": 12476,
					"Frees": 12353
				},
				{
					"Size": 1152,
					"Mallocs": 15082,
					"Frees": 14992
				},
				{
					"Size": 1280,
					"Mallocs": 1477,
					"Frees": 1466
				},
				{
					"Size": 1408,
					"Mallocs": 3459,
					"Frees": 3289
				},
				{
					"Size": 1536,
					"Mallocs": 150,
					"Frees": 125
				},
				{
					"Size": 1792,
					"Mallocs": 177,
					"Frees": 155
				},
				{
					"Size": 2048,
					"Mallocs": 703,
					"Frees": 659
				},
				{
					"Size": 2304,
					"Mallocs": 4386,
					"Frees": 4356
				},
				{
					"Size": 2688,
					"Mallocs": 2657,
					"Frees": 2639
				},
				{
					"Size": 3072,
					"Mallocs": 184,
					"Frees": 182
				},
				{
					"Size": 3200,
					"Mallocs": 8,
					"Frees": 7
				},
				{
					"Size": 3456,
					"Mallocs": 4,
					"Frees": 4
				},
				{
					"Size": 4096,
					"Mallocs": 348,
					"Frees": 333
				},
				{
					"Size": 4864,
					"Mallocs": 592,
					"Frees": 584
				},
				{
					"Size": 5376,
					"Mallocs": 551,
					"Frees": 542
				},
				{
					"Size": 6144,
					"Mallocs": 4927,
					"Frees": 4908
				},
				{
					"Size": 6528,
					"Mallocs": 403,
					"Frees": 400
				},
				{
					"Size": 6784,
					"Mallocs": 289,
					"Frees": 284
				},
				{
					"Size": 6912,
					"Mallocs": 154,
					"Frees": 154
				},
				{
					"Size": 8192,
					"Mallocs": 1254,
					"Frees": 1049
				},
				{
					"Size": 9472,
					"Mallocs": 15,
					"Frees": 2
				},
				{
					"Size": 9728,
					"Mallocs": 3,
					"Frees": 3
				},
				{
					"Size": 10240,
					"Mallocs": 0,
					"Frees": 0
				},
				{
					"Size": 10880,
					"Mallocs": 50,
					"Frees": 49
				},
				{
					"Size": 12288,
					"Mallocs": 2533,
					"Frees": 2512
				},
				{
					"Size": 13568,
					"Mallocs": 6,
					"Frees": 6
				},
				{
					"Size": 14336,
					"Mallocs": 3,
					"Frees": 3
				},
				{
					"Size": 16384,
					"Mallocs": 128,
					"Frees": 125
				},
				{
					"Size": 18432,
					"Mallocs": 2,
					"Frees": 2
				},
				{
					"Size": 19072,
					"Mallocs": 1,
					"Frees": 0
				}
			]
		},
		"runtime": {
			"name": "runtime",
			"tags": {},
			"values": {
				"Alloc": 10880272,
				"Frees": 4152498,
				"HeapAlloc": 10880272,
				"HeapIdle": 53264384,
				"HeapInUse": 12697600,
				"HeapObjects": 112203,
				"HeapReleased": 52584448,
				"HeapSys": 65961984,
				"Lookups": 0,
				"Mallocs": 4264701,
				"NumGC": 228,
				"NumGoroutine": 19,
				"PauseTotalNs": 14059643,
				"Sys": 73141649,
				"TotalAlloc": 973855808
			}
		},
		"queryExecutor": {
			"name": "queryExecutor",
			"tags": null,
			"values": {
				"queriesActive": 0,
				"queriesExecuted": 0,
				"queriesFinished": 0,
				"queryDurationNs": 0,
				"recoveredPanics": 0
			}
		},
		"database:_internal": {
			"name": "database",
			"tags": {
				"database": "_internal"
			},
			"values": {
				"numMeasurements": 12,
				"numSeries": 17
			}
		},
		"shard:/var/lib/influxdb/data/_internal/monitor/1:1": {
			"name": "shard",
			"tags": {
				"database": "_internal",
				"engine": "tsm1",
				"id": "1",
				"indexType": "inmem",
				"path": "/var/lib/influxdb/data/_internal/monitor/1",
				"retentionPolicy": "monitor",
				"walPath": "/var/lib/influxdb/wal/_internal/monitor/1"
			},
			"values": {
				"diskBytes": 103008,
				"fieldsCreate": 115,
				"seriesCreate": 12,
				"writeBytes": 0,
				"writePointsDropped": 0,
				"writePointsErr": 0,
				"writePointsOk": 23406,
				"writeReq": 1951,
				"writeReqErr": 0,
				"writeReqOk": 1951,
				"writeValuesOk": 224308
			}
		},
		"tsm1_engine:/var/lib/influxdb/data/_internal/monitor/1:1": {
			"name": "tsm1_engine",
			"tags": {
				"database": "_internal",
				"engine": "tsm1",
				"id": "1",
				"indexType": "inmem",
				"path": "/var/lib/influxdb/data/_internal/monitor/1",
				"retentionPolicy": "monitor",
				"walPath": "/var/lib/influxdb/wal/_internal/monitor/1"
			},
			"values": {
				"cacheCompactionDuration": 136680896,
				"cacheCompactionErr": 0,
				"cacheCompactions": 1,
				"cacheCompactionsActive": 0,
				"tsmFullCompactionDuration": 0,
				"tsmFullCompactionErr": 0,
				"tsmFullCompactionQueue": 0,
				"tsmFullCompactions": 0,
				"tsmFullCompactionsActive": 0,
				"tsmLevel1CompactionDuration": 0,
				"tsmLevel1CompactionErr": 0,
				"tsmLevel1CompactionQueue": 0,
				"tsmLevel1Compactions": 0,
				"tsmLevel1CompactionsActive": 0,
				"tsmLevel2CompactionDuration": 0,
				"tsmLevel2CompactionErr": 0,
				"tsmLevel2CompactionQueue": 0,
				"tsmLevel2Compactions": 0,
				"tsmLevel2CompactionsActive": 0,
				"tsmLevel3CompactionDuration": 0,
				"tsmLevel3CompactionErr": 0,
				"tsmLevel3CompactionQueue": 0,
				"tsmLevel3Compactions": 0,
				"tsmLevel3CompactionsActive": 0,
				"tsmOptimizeCompactionDuration": 0,
				"tsmOptimizeCompactionErr": 0,
				"tsmOptimizeCompactionQueue": 0,
				"tsmOptimizeCompactions": 0,
				"tsmOptimizeCompactionsActive": 0
			}
		},
		"tsm1_cache:/var/lib/influxdb/data/_internal/monitor/1:1": {
			"name": "tsm1_cache",
			"tags": {
				"database": "_internal",
				"engine": "tsm1",
				"id": "1",
				"indexType": "inmem",
				"path": "/var/lib/influxdb/data/_internal/monitor/1",
				"retentionPolicy": "monitor",
				"walPath": "/var/lib/influxdb/wal/_internal/monitor/1"
			},
			"values": {
				"WALCompactionTimeMs": 136,
				"cacheAgeMs": 7991,
				"cachedBytes": 3604337,
				"diskBytes": 0,
				"memBytes": 0,
				"snapshotCount": 0,
				"writeDropped": 0,
				"writeErr": 0,
				"writeOk": 1951
			}
		},
		"tsm1_filestore:/var/lib/influxdb/data/_internal/monitor/1:1": {
			"name": "tsm1_filestore",
			"tags": {
				"database": "_internal",
				"engine": "tsm1",
				"id": "1",
				"indexType": "inmem",
				"path": "/var/lib/influxdb/data/_internal/monitor/1",
				"retentionPolicy": "monitor",
				"walPath": "/var/lib/influxdb/wal/_internal/monitor/1"
			},
			"values": {
				"diskBytes": 103008,
				"numFiles": 1
			}
		},
		"tsm1_wal:/var/lib/influxdb/data/_internal/monitor/1:1": {
			"name": "tsm1_wal",
			"tags": {
				"database": "_internal",
				"engine": "tsm1",
				"id": "1",
				"indexType": "inmem",
				"path": "/var/lib/influxdb/data/_internal/monitor/1",
				"retentionPolicy": "monitor",
				"walPath": "/var/lib/influxdb/wal/_internal/monitor/1"
			},
			"values": {
				"currentSegmentDiskBytes": 0,
				"oldSegmentsDiskBytes": 0,
				"writeErr": 0,
				"writeOk": 1951
			}
		},
		"shard:/var/lib/influxdb/data/_internal/monitor/2:2": {
			"name": "shard",
			"tags": {
				"database": "_internal",
				"engine": "tsm1",
				"id": "2",
				"indexType": "inmem",
				"path": "/var/lib/influxdb/data/_internal/monitor/2",
				"retentionPolicy": "monitor",
				"walPath": "/var/lib/influxdb/wal/_internal/monitor/2"
			},
			"values": {
				"diskBytes": 1764827,
				"fieldsCreate": 115,
				"seriesCreate": 17,
				"writeBytes": 0,
				"writePointsDropped": 0,
				"writePointsErr": 0,
				"writePointsOk": 8172,
				"writeReq": 481,
				"writeReqErr": 0,
				"writeReqOk": 481,
				"writeValuesOk": 81715
			}
		},
		"tsm1_engine:/var/lib/influxdb/data/_internal/monitor/2:2": {
			"name": "tsm1_engine",
			"tags": {
				"database": "_internal",
				"engine": "tsm1",
				"id": "2",
				"indexType": "inmem",
				"path": "/var/lib/influxdb/data/_internal/monitor/2",
				"retentionPolicy": "monitor",
				"walPath": "/var/lib/influxdb/wal/_internal/monitor/2"
			},
			"values": {
				"cacheCompactionDuration": 0,
				"cacheCompactionErr": 0,
				"cacheCompactions": 0,
				"cacheCompactionsActive": 0,
				"tsmFullCompactionDuration": 0,
				"tsmFullCompactionErr": 0,
				"tsmFullCompactionQueue": 0,
				"tsmFullCompactions": 0,
				"tsmFullCompactionsActive": 0,
				"tsmLevel1CompactionDuration": 0,
				"tsmLevel1CompactionErr": 0,
				"tsmLevel1CompactionQueue": 0,
				"tsmLevel1Compactions": 0,
				"tsmLevel1CompactionsActive": 0,
				"tsmLevel2CompactionDuration": 0,
				"tsmLevel2CompactionErr": 0,
				"tsmLevel2CompactionQueue": 0,
				"tsmLevel2Compactions": 0,
				"tsmLevel2CompactionsActive": 0,
				"tsmLevel3CompactionDuration": 0,
				"tsmLevel3CompactionErr": 0,
				"tsmLevel3CompactionQueue": 0,
				"tsmLevel3Compactions": 0,
				"tsmLevel3CompactionsActive": 0,
				"tsmOptimizeCompactionDuration": 0,
				"tsmOptimizeCompactionErr": 0,
				"tsmOptimizeCompactionQueue": 0,
				"tsmOptimizeCompactions": 0,
				"tsmOptimizeCompactionsActive": 0
			}
		},
		"tsm1_cache:/var/lib/influxdb/data/_internal/monitor/2:2": {
			"name": "tsm1_cache",
			"tags": {
				"database": "_internal",
				"engine": "tsm1",
				"id": "2",
				"indexType": "inmem",
				"path": "/var/lib/influxdb/data/_internal/monitor/2",
				"retentionPolicy": "monitor",
				"walPath": "/var/lib/influxdb/wal/_internal/monitor/2"
			},
			"values": {
				"WALCompactionTimeMs": 0,
				"cacheAgeMs": 4805017,
				"cachedBytes": 0,
				"diskBytes": 0,
				"memBytes": 1335349,
				"snapshotCount": 0,
				"writeDropped": 0,
				"writeErr": 0,
				"writeOk": 481
			}
		},
		"tsm1_filestore:/var/lib/influxdb/data/_internal/monitor/2:2": {
			"name": "tsm1_filestore",
			"tags": {
				"database": "_internal",
				"engine": "tsm1",
				"id": "2",
				"indexType": "inmem",
				"path": "/var/lib/influxdb/data/_internal/monitor/2",
				"retentionPolicy": "monitor",
				"walPath": "/var/lib/influxdb/wal/_internal/monitor/2"
			},
			"values": {
				"diskBytes": 0,
				"numFiles": 0
			}
		},
		"tsm1_wal:/var/lib/influxdb/data/_internal/monitor/2:2": {
			"name": "tsm1_wal",
			"tags": {
				"database": "_internal",
				"engine": "tsm1",
				"id": "2",
				"indexType": "inmem",
				"path": "/var/lib/influxdb/data/_internal/monitor/2",
				"retentionPolicy": "monitor",
				"walPath": "/var/lib/influxdb/wal/_internal/monitor/2"
			},
			"values": {
				"currentSegmentDiskBytes": 1764827,
				"oldSegmentsDiskBytes": 0,
				"writeErr": 0,
				"writeOk": 481
			}
		},
		"write": {
			"name": "write",
			"tags": null,
			"values": {
				"pointReq": 31578,
				"pointReqLocal": 31578,
				"req": 2432,
				"subWriteDrop": 0,
				"subWriteOk": 2432,
				"writeDrop": 0,
				"writeError": 0,
				"writeOk": 2432,
				"writeTimeout": 0
			}
		},
		"subscriber": {
			"name": "subscriber",
			"tags": null,
			"values": {
				"createFailures": 0,
				"pointsWritten": 0,
				"writeFailures": 0
			}
		},
		"cq": {
			"name": "cq",
			"tags": null,
			"values": {
				"queryFail": 0,
				"queryOk": 0
			}
		},
		"httpd::8086": {
			"name": "httpd",
			"tags": {
				"bind": ":8086"
			},
			"values": {
				"authFail": 0,
				"clientError": 1422,
				"fluxQueryReq": 0,
				"fluxQueryReqDurationNs": 0,
				"pingReq": 1,
				"pointsWrittenDropped": 0,
				"pointsWrittenFail": 0,
				"pointsWrittenOK": 0,
				"promReadReq": 0,
				"promWriteReq": 0,
				"queryReq": 0,
				"queryReqDurationNs": 0,
				"queryRespBytes": 0,
				"recoveredPanics": 0,
				"req": 2859,
				"reqActive": 1,
				"reqDurationNs": 4377096764,
				"serverError": 0,
				"statusReq": 0,
				"valuesWrittenOK": 0,
				"writeReq": 1422,
				"writeReqActive": 0,
				"writeReqBytes": 0,
				"writeReqDurationNs": 78005990
			}
		}
	}
	`
)
