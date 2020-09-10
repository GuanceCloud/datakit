package wechatminiprogram

import "gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"

const (
	grantType            = "client_credential"
	baseURL              = "https://api.weixin.qq.com/"
	accessTokenURL       = "cgi-bin/token" //nolint:gosec
	DailySummaryURL      = "datacube/getweanalysisappiddailysummarytrend"
	DailyVisitTrendURL   = "datacube/getweanalysisappiddailyvisittrend"
	UserPortraitURL      = "datacube/getweanalysisappiduserportrait"
	VisitDistributionURL = "datacube/getweanalysisappidvisitdistribution"
	VisitPageURL         = "datacube/getweanalysisappidvisitpage"
	jsErrSearchURL       = "wxaapi/log/jserr_search"
	PerformanceURL       = "wxaapi/log/get_performance"
	offect               = 8 * 3600
	errcode              = 40000
	daySecond            = 24 * 3600
)

var (
	l         *logger.Logger
	inputName = "wechatminiprogram"
	sampleCfg = `
[[inputs.wechatminiprogram]]

## Small program unique unique credentials   AppID
 appid = ""
## Small program unique credential key   AppSecret
 secret = ""
[inputs.wechatminiprogram.analysis]
 runtime = "18:13"
 name = [
	"DailySummary",
	"VisitDistribution",
	"UserPortrait",
	"DailyVisitTrend",
	"VisitPage"]

#[inputs.wechatminiprogram.operation]
# name = [
#	"JsErrSearch",
#	"Performance"]
`
)

var Config = map[string]map[string]string{
	"access_source_session_cnt": {
		"1":  "小程序历史列表",
		"2":  "搜索",
		"3":  "会话",
		"4":  "扫一扫二维码",
		"5":  "公众号主页",
		"6":  "聊天顶部",
		"7":  "系统桌面",
		"8":  "小程序主页",
		"9":  "附近的小程序",
		"10": "其他",
		"11": "模板消息",
		"12": "客服消息",
		"13": "公众号菜单",
		"14": "APP分享",
		"15": "支付完成页",
		"16": "长按识别二维码",
		"17": "相册选取二维码",
		"18": "公众号文章",
		"19": "钱包",
		"20": "卡包",
		"21": "小程序内卡券",
		"22": "其他小程序",
		"23": "其他小程序返回",
		"24": "卡券适用门店列表",
		"25": "搜索框快捷入口",
		"26": "小程序客服消息",
		"27": "公众号下发",
		"28": "系统会话菜单",
		"29": "任务栏-最近使用",
		"30": "长按小程序菜单圆点",
		"31": "连wifi成功页",
		"32": "城市服务",
		"33": "微信广告",
		"34": "其他移动应用",
		"35": "发现入口-我的小程序",
		"36": "任务栏-我的小程序",
		"37": "微信圈子",
		"38": "手机充值",
		"39": "H5",
		"40": "插件",
		"41": "大家在用",
		"42": "发现页",
		"43": "浮窗",
		"44": "附近的人",
		"45": "看一看",
		"46": "朋友圈",
		"47": "企业微信",
		"48": "视频",
		"49": "收藏",
		"50": "微信红包",
		"51": "微信游戏中心",
		"52": "摇一摇",
		"53": "公众号导购消息",
		"54": "识物",
		"55": "小程序订单",
		"56": "小程序直播",
		"57": "群工具",
	},
	"access_staytime_info": {
		"1": "0-2s",
		"2": "3-5s",
		"3": "6-10s",
		"4": "11-20s",
		"5": "20-30s",
		"6": "30-50s",
		"7": "50-100s",
		"8": ">100s",
	},
	"access_depth_info": {
		"1": "1 页",
		"2": "2 页",
		"3": "3 页",
		"4": "4 页",
		"5": "5 页",
		"6": "6-10 页",
		"7": ">10 页",
	},
	"cost_time_type": {
		"1": "启动总耗时",
		"2": "下载耗时",
		"3": "初次渲染耗时",
	},
}
