// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rum

import "github.com/gobwas/glob"

var CDNList = &CDNPool{
	literal: map[string]CDN{
		"15cdn.com": {
			Name:    "腾正安全加速（原 15CDN）",
			Website: "https://www.15cdn.com",
		},
		"tzcdn.cn": {
			Name:    "腾正安全加速（原 15CDN）",
			Website: "https://www.15cdn.com",
		},
		"cedexis.net": {
			Name:    "Cedexis GSLB",
			Website: "https://www.cedexis.com/",
		},
		"cdxcn.cn": {
			Name:    "Cedexis GSLB (For China)",
			Website: "https://www.cedexis.com/",
		},
		"qhcdn.com": {
			Name:    "360 云 CDN (由奇安信运营)",
			Website: "https://cloud.360.cn/doc?name=cdn",
		},
		"qh-cdn.com": {
			Name:    "360 云 CDN (由奇虎 360 运营)",
			Website: "https://cloud.360.cn/doc?name=cdn",
		},
		"qihucdn.com": {
			Name:    "360 云 CDN (由奇虎 360 运营)",
			Website: "https://cloud.360.cn/doc?name=cdn",
		},
		"360cdn.com": {
			Name:    "360 云 CDN (由奇虎 360 运营)",
			Website: "https://cloud.360.cn/doc?name=cdn",
		},
		"360cloudwaf.com": {
			Name:    "奇安信网站卫士",
			Website: "https://wangzhan.qianxin.com",
		},
		"360anyu.com": {
			Name:    "奇安信网站卫士",
			Website: "https://wangzhan.qianxin.com",
		},
		"360safedns.com": {
			Name:    "奇安信网站卫士",
			Website: "https://wangzhan.qianxin.com",
		},
		"360wzws.com": {
			Name:    "奇安信网站卫士",
			Website: "https://wangzhan.qianxin.com",
		},
		"akamai.net": {
			Name:    "Akamai CDN",
			Website: "https://www.akamai.com",
		},
		"akamaiedge.net": {
			Name:    "Akamai CDN",
			Website: "https://www.akamai.com",
		},
		"ytcdn.net": {
			Name:    "Akamai CDN",
			Website: "https://www.akamai.com",
		},
		"edgesuite.net": {
			Name:    "Akamai CDN",
			Website: "https://www.akamai.com",
		},
		"akamaitech.net": {
			Name:    "Akamai CDN",
			Website: "https://www.akamai.com",
		},
		"akamaitechnologies.com": {
			Name:    "Akamai CDN",
			Website: "https://www.akamai.com",
		},
		"edgekey.net": {
			Name:    "Akamai CDN",
			Website: "https://www.akamai.com",
		},
		"tl88.net": {
			Name:    "易通锐进（Akamai 中国）由网宿承接",
			Website: "https://www.akamai.com",
		},
		"cloudfront.net": {
			Name:    "AWS CloudFront",
			Website: "https://aws.amazon.com/cn/cloudfront/",
		},
		"worldcdn.net": {
			Name:    "CDN.NET",
			Website: "https://cdn.net",
		},
		"worldssl.net": {
			Name:    "CDN.NET / CDNSUN / ONAPP",
			Website: "https://cdn.net",
		},
		"cdn77.org": {
			Name:    "CDN77",
			Website: "https://www.cdn77.com/",
		},
		"panthercdn.com": {
			Name:    "CDNetworks",
			Website: "https://www.cdnetworks.com",
		},
		"cdnga.net": {
			Name:    "CDNetworks",
			Website: "https://www.cdnetworks.com",
		},
		"cdngc.net": {
			Name:    "CDNetworks",
			Website: "https://www.cdnetworks.com",
		},
		"gccdn.net": {
			Name:    "CDNetworks",
			Website: "https://www.cdnetworks.com",
		},
		"gccdn.cn": {
			Name:    "CDNetworks",
			Website: "https://www.cdnetworks.com",
		},
		"akamaized.net": {
			Name:    "Akamai CDN",
			Website: "https://www.akamai.com",
		},
		"126.net": {
			Name:    "网易云 CDN",
			Website: "https://www.163yun.com/product/cdn",
		},
		"163jiasu.com": {
			Name:    "网易云 CDN",
			Website: "https://www.163yun.com/product/cdn",
		},
		"amazonaws.com": {
			Name:    "AWS Cloud",
			Website: "https://aws.amazon.com/cn/cloudfront/",
		},
		"cdn77.net": {
			Name:    "CDN77",
			Website: "https://www.cdn77.com/",
		},
		"cdnify.io": {
			Name:    "CDNIFY",
			Website: "https://cdnify.com",
		},
		"cdnsun.net": {
			Name:    "CDNSUN",
			Website: "https://cdnsun.com",
		},
		"bdydns.com": {
			Name:    "百度云 CDN",
			Website: "https://cloud.baidu.com/product/cdn.html",
		},
		"ccgslb.com.cn": {
			Name:    "蓝汛 CDN",
			Website: "https://cn.chinacache.com/",
		},
		"ccgslb.net": {
			Name:    "蓝汛 CDN",
			Website: "https://cn.chinacache.com/",
		},
		"ccgslb.com": {
			Name:    "蓝汛 CDN",
			Website: "https://cn.chinacache.com/",
		},
		"ccgslb.cn": {
			Name:    "蓝汛 CDN",
			Website: "https://cn.chinacache.com/",
		},
		"c3cache.net": {
			Name:    "蓝汛 CDN",
			Website: "https://cn.chinacache.com/",
		},
		"c3dns.net": {
			Name:    "蓝汛 CDN",
			Website: "https://cn.chinacache.com/",
		},
		"chinacache.net": {
			Name:    "蓝汛 CDN",
			Website: "https://cn.chinacache.com/",
		},
		"wswebcdn.com": {
			Name:    "网宿 CDN",
			Website: "https://www.wangsu.com/",
		},
		"lxdns.com": {
			Name:    "网宿 CDN",
			Website: "https://www.wangsu.com/",
		},
		"wswebpic.com": {
			Name:    "网宿 CDN",
			Website: "https://www.wangsu.com/",
		},
		"cloudflare.net": {
			Name:    "Cloudflare",
			Website: "https://www.cloudflare.com",
		},
		"akadns.net": {
			Name:    "Akamai CDN",
			Website: "https://www.akamai.com",
		},
		"chinanetcenter.com": {
			Name:    "网宿 CDN",
			Website: "https://www.wangsu.com",
		},
		"customcdn.com.cn": {
			Name:    "网宿 CDN",
			Website: "https://www.wangsu.com",
		},
		"customcdn.cn": {
			Name:    "网宿 CDN",
			Website: "https://www.wangsu.com",
		},
		"51cdn.com": {
			Name:    "网宿 CDN",
			Website: "https://www.wangsu.com",
		},
		"wscdns.com": {
			Name:    "网宿 CDN",
			Website: "https://www.wangsu.com",
		},
		"cdn20.com": {
			Name:    "网宿 CDN",
			Website: "https://www.wangsu.com",
		},
		"wsdvs.com": {
			Name:    "网宿 CDN",
			Website: "https://www.wangsu.com",
		},
		"wsglb0.com": {
			Name:    "网宿 CDN",
			Website: "https://www.wangsu.com",
		},
		"speedcdns.com": {
			Name:    "网宿 CDN",
			Website: "https://www.wangsu.com",
		},
		"wtxcdn.com": {
			Name:    "网宿 CDN",
			Website: "https://www.wangsu.com",
		},
		"wsssec.com": {
			Name:    "网宿 WAF CDN",
			Website: "https://www.wangsu.com",
		},
		"fastly.net": {
			Name:    "Fastly",
			Website: "https://www.fastly.com",
		},
		"fastlylb.net": {
			Name:    "Fastly",
			Website: "https://www.fastly.com/",
		},
		"hwcdn.net": {
			Name:    "Stackpath (原 Highwinds)",
			Website: "https://www.stackpath.com/highwinds",
		},
		"incapdns.net": {
			Name:    "Incapsula CDN",
			Website: "https://www.incapsula.com",
		},
		"kxcdn.com.": {
			Name:    "KeyCDN",
			Website: "https://www.keycdn.com/",
		},
		"lswcdn.net": {
			Name:    "LeaseWeb CDN",
			Website: "https://www.leaseweb.com/cdn",
		},
		"mwcloudcdn.com": {
			Name:    "QUANTIL (网宿)",
			Website: "https://www.quantil.com/",
		},
		"mwcname.com": {
			Name:    "QUANTIL (网宿)",
			Website: "https://www.quantil.com/",
		},
		"azureedge.net": {
			Name:    "Microsoft Azure CDN",
			Website: "https://azure.microsoft.com/en-us/services/cdn/",
		},
		"msecnd.net": {
			Name:    "Microsoft Azure CDN",
			Website: "https://azure.microsoft.com/en-us/services/cdn/",
		},
		"mschcdn.com": {
			Name:    "Microsoft Azure CDN",
			Website: "https://azure.microsoft.com/en-us/services/cdn/",
		},
		"v0cdn.net": {
			Name:    "Microsoft Azure CDN",
			Website: "https://azure.microsoft.com/en-us/services/cdn/",
		},
		"azurewebsites.net": {
			Name:    "Microsoft Azure App Service",
			Website: "https://azure.microsoft.com/en-us/services/app-service/",
		},
		"azurewebsites.windows.net": {
			Name:    "Microsoft Azure App Service",
			Website: "https://azure.microsoft.com/en-us/services/app-service/",
		},
		"trafficmanager.net": {
			Name:    "Microsoft Azure Traffic Manager",
			Website: "https://azure.microsoft.com/en-us/services/traffic-manager/",
		},
		"cloudapp.net": {
			Name:    "Microsoft Azure",
			Website: "https://azure.microsoft.com",
		},
		"chinacloudsites.cn": {
			Name:    "世纪互联旗下上海蓝云（承载 Azure 中国）",
			Website: "https://www.21vbluecloud.com/",
		},
		"spdydns.com": {
			Name:    "云端智度融合 CDN",
			Website: "https://www.isurecloud.net/index.html",
		},
		"jiashule.com": {
			Name:    "知道创宇云安全加速乐CDN",
			Website: "https://www.yunaq.com/jsl/",
		},
		"jiasule.org": {
			Name:    "知道创宇云安全加速乐CDN",
			Website: "https://www.yunaq.com/jsl/",
		},
		"365cyd.cn": {
			Name:    "知道创宇云安全创宇盾（政务专用）",
			Website: "https://www.yunaq.com/cyd/",
		},
		"huaweicloud.com": {
			Name:    "华为云WAF高防云盾",
			Website: "https://www.huaweicloud.com/product/aad.html",
		},
		"cdnhwc1.com": {
			Name:    "华为云 CDN",
			Website: "https://www.huaweicloud.com/product/cdn.html",
		},
		"cdnhwc2.com": {
			Name:    "华为云 CDN",
			Website: "https://www.huaweicloud.com/product/cdn.html",
		},
		"cdnhwc3.com": {
			Name:    "华为云 CDN",
			Website: "https://www.huaweicloud.com/product/cdn.html",
		},
		"dnion.com": {
			Name:    "帝联科技",
			Website: "http://www.dnion.com/",
		},
		"ewcache.com": {
			Name:    "帝联科技",
			Website: "http://www.dnion.com/",
		},
		"globalcdn.cn": {
			Name:    "帝联科技",
			Website: "http://www.dnion.com/",
		},
		"tlgslb.com": {
			Name:    "帝联科技",
			Website: "http://www.dnion.com/",
		},
		"fastcdn.com": {
			Name:    "帝联科技",
			Website: "http://www.dnion.com/",
		},
		"flxdns.com": {
			Name:    "帝联科技",
			Website: "http://www.dnion.com/",
		},
		"dlgslb.cn": {
			Name:    "帝联科技",
			Website: "http://www.dnion.com/",
		},
		"newdefend.cn": {
			Name:    "牛盾云安全",
			Website: "https://www.newdefend.com",
		},
		"ffdns.net": {
			Name:    "CloudXNS",
			Website: "https://www.cloudxns.net",
		},
		"aocdn.com": {
			Name:    "可靠云 CDN (贴图库)",
			Website: "http://www.kekaoyun.com/",
		},
		"bsgslb.cn": {
			Name:    "白山云 CDN",
			Website: "https://zh.baishancloud.com/",
		},
		"qingcdn.com": {
			Name:    "白山云 CDN",
			Website: "https://zh.baishancloud.com/",
		},
		"bsclink.cn": {
			Name:    "白山云 CDN",
			Website: "https://zh.baishancloud.com/",
		},
		"trpcdn.net": {
			Name:    "白山云 CDN",
			Website: "https://zh.baishancloud.com/",
		},
		"anquan.io": {
			Name:    "牛盾云安全",
			Website: "https://www.newdefend.com",
		},
		"cloudglb.com": {
			Name:    "快网 CDN",
			Website: "http://www.fastweb.com.cn/",
		},
		"fastweb.com": {
			Name:    "快网 CDN",
			Website: "http://www.fastweb.com.cn/",
		},
		"fastwebcdn.com": {
			Name:    "快网 CDN",
			Website: "http://www.fastweb.com.cn/",
		},
		"cloudcdn.net": {
			Name:    "快网 CDN",
			Website: "http://www.fastweb.com.cn/",
		},
		"fwcdn.com": {
			Name:    "快网 CDN",
			Website: "http://www.fastweb.com.cn/",
		},
		"fwdns.net": {
			Name:    "快网 CDN",
			Website: "http://www.fastweb.com.cn/",
		},
		"hadns.net": {
			Name:    "快网 CDN",
			Website: "http://www.fastweb.com.cn/",
		},
		"hacdn.net": {
			Name:    "快网 CDN",
			Website: "http://www.fastweb.com.cn/",
		},
		"cachecn.com": {
			Name:    "快网 CDN",
			Website: "http://www.fastweb.com.cn/",
		},
		"qingcache.com": {
			Name:    "青云 CDN",
			Website: "https://www.qingcloud.com/products/cdn/",
		},
		"qingcloud.com": {
			Name:    "青云 CDN",
			Website: "https://www.qingcloud.com/products/cdn/",
		},
		"frontwize.com": {
			Name:    "青云 CDN",
			Website: "https://www.qingcloud.com/products/cdn/",
		},
		"msscdn.com": {
			Name:    "美团云 CDN",
			Website: "https://www.mtyun.com/product/cdn",
		},
		"800cdn.com": {
			Name:    "西部数码",
			Website: "https://www.west.cn",
		},
		"tbcache.com": {
			Name:    "阿里云 CDN",
			Website: "https://www.aliyun.com/product/cdn",
		},
		"aliyun-inc.com": {
			Name:    "阿里云 CDN",
			Website: "https://www.aliyun.com/product/cdn",
		},
		"aliyuncs.com": {
			Name:    "阿里云 CDN",
			Website: "https://www.aliyun.com/product/cdn",
		},
		"alikunlun.net": {
			Name:    "阿里云 CDN",
			Website: "https://www.aliyun.com/product/cdn",
		},
		"alikunlun.com": {
			Name:    "阿里云 CDN",
			Website: "https://www.aliyun.com/product/cdn",
		},
		"alicdn.com": {
			Name:    "阿里云 CDN",
			Website: "https://www.aliyun.com/product/cdn",
		},
		"aligaofang.com": {
			Name:    "阿里云盾高防",
			Website: "https://www.aliyun.com/product/ddos",
		},
		"yundunddos.com": {
			Name:    "阿里云盾高防",
			Website: "https://www.aliyun.com/product/ddos",
		},
		"cdngslb.com": {
			Name:    "阿里云 CDN",
			Website: "https://www.aliyun.com/product/cdn",
		},
		"yunjiasu-cdn.net": {
			Name:    "百度云加速",
			Website: "https://su.baidu.com",
		},
		"momentcdn.com": {
			Name:    "魔门云 CDN",
			Website: "https://www.cachemoment.com",
		},
		"aicdn.com": {
			Name:    "又拍云",
			Website: "https://www.upyun.com",
		},
		"qbox.me": {
			Name:    "七牛云",
			Website: "https://www.qiniu.com",
		},
		"qiniu.com": {
			Name:    "七牛云",
			Website: "https://www.qiniu.com",
		},
		"qiniudns.com": {
			Name:    "七牛云",
			Website: "https://www.qiniu.com",
		},
		"jcloudcs.com": {
			Name:    "京东云 CDN",
			Website: "https://www.jdcloud.com/cn/products/cdn",
		},
		"jdcdn.com": {
			Name:    "京东云 CDN",
			Website: "https://www.jdcloud.com/cn/products/cdn",
		},
		"qianxun.com": {
			Name:    "京东云 CDN",
			Website: "https://www.jdcloud.com/cn/products/cdn",
		},
		"jcloudlb.com": {
			Name:    "京东云 CDN",
			Website: "https://www.jdcloud.com/cn/products/cdn",
		},
		"jcloud-cdn.com": {
			Name:    "京东云 CDN",
			Website: "https://www.jdcloud.com/cn/products/cdn",
		},
		"maoyun.tv": {
			Name:    "猫云融合 CDN",
			Website: "https://www.maoyun.com/",
		},
		"maoyundns.com": {
			Name:    "猫云融合 CDN",
			Website: "https://www.maoyun.com/",
		},
		"xgslb.net": {
			Name:    "WebLuker (蓝汛)",
			Website: "http://www.webluker.com",
		},
		"ucloud.cn": {
			Name:    "UCloud CDN",
			Website: "https://www.ucloud.cn/site/product/ucdn.html",
		},
		"ucloud.com.cn": {
			Name:    "UCloud CDN",
			Website: "https://www.ucloud.cn/site/product/ucdn.html",
		},
		"cdndo.com": {
			Name:    "UCloud CDN",
			Website: "https://www.ucloud.cn/site/product/ucdn.html",
		},
		"zenlogic.net": {
			Name:    "Zenlayer CDN",
			Website: "https://www.zenlayer.com",
		},
		"ogslb.com": {
			Name:    "Zenlayer CDN",
			Website: "https://www.zenlayer.com",
		},
		"uxengine.net": {
			Name:    "Zenlayer CDN",
			Website: "https://www.zenlayer.com",
		},
		"tan14.net": {
			Name:    "TAN14 CDN",
			Website: "http://www.tan14.cn/",
		},
		"verycloud.cn": {
			Name:    "VeryCloud 云分发",
			Website: "https://www.verycloud.cn/",
		},
		"verycdn.net": {
			Name:    "VeryCloud 云分发",
			Website: "https://www.verycloud.cn/",
		},
		"verygslb.com": {
			Name:    "VeryCloud 云分发",
			Website: "https://www.verycloud.cn/",
		},
		"xundayun.cn": {
			Name:    "SpeedyCloud CDN",
			Website: "https://www.speedycloud.cn/zh/Products/CDN/CloudDistribution.html",
		},
		"xundayun.com": {
			Name:    "SpeedyCloud CDN",
			Website: "https://www.speedycloud.cn/zh/Products/CDN/CloudDistribution.html",
		},
		"speedycloud.cc": {
			Name:    "SpeedyCloud CDN",
			Website: "https://www.speedycloud.cn/zh/Products/CDN/CloudDistribution.html",
		},
		"mucdn.net": {
			Name:    "Verizon CDN (Edgecast)",
			Website: "https://www.verizondigitalmedia.com/platform/edgecast-cdn/",
		},
		"nucdn.net": {
			Name:    "Verizon CDN (Edgecast)",
			Website: "https://www.verizondigitalmedia.com/platform/edgecast-cdn/",
		},
		"alphacdn.net": {
			Name:    "Verizon CDN (Edgecast)",
			Website: "https://www.verizondigitalmedia.com/platform/edgecast-cdn/",
		},
		"systemcdn.net": {
			Name:    "Verizon CDN (Edgecast)",
			Website: "https://www.verizondigitalmedia.com/platform/edgecast-cdn/",
		},
		"edgecastcdn.net": {
			Name:    "Verizon CDN (Edgecast)",
			Website: "https://www.verizondigitalmedia.com/platform/edgecast-cdn/",
		},
		"zetacdn.net": {
			Name:    "Verizon CDN (Edgecast)",
			Website: "https://www.verizondigitalmedia.com/platform/edgecast-cdn/",
		},
		"coding.io": {
			Name:    "Coding Pages",
			Website: "https://coding.net/pages",
		},
		"coding.me": {
			Name:    "Coding Pages",
			Website: "https://coding.net/pages",
		},
		"gitlab.io": {
			Name:    "GitLab Pages",
			Website: "https://docs.gitlab.com/ee/user/project/pages/",
		},
		"github.io": {
			Name:    "GitHub Pages",
			Website: "https://pages.github.com/",
		},
		"herokuapp.com": {
			Name:    "Heroku SaaS",
			Website: "https://www.heroku.com",
		},
		"googleapis.com": {
			Name:    "Google Cloud Storage",
			Website: "https://cloud.google.com/storage/",
		},
		"netdna.com": {
			Name:    "Stackpath (原 MaxCDN)",
			Website: "https://www.stackpath.com/maxcdn/",
		},
		"netdna-cdn.com": {
			Name:    "Stackpath (原 MaxCDN)",
			Website: "https://www.stackpath.com/maxcdn/",
		},
		"netdna-ssl.com": {
			Name:    "Stackpath (原 MaxCDN)",
			Website: "https://www.stackpath.com/maxcdn/",
		},
		"cdntip.com": {
			Name:    "腾讯云 CDN",
			Website: "https://cloud.tencent.com/product/cdn-scd",
		},
		"dnsv1.com": {
			Name:    "腾讯云 CDN",
			Website: "https://cloud.tencent.com/product/cdn-scd",
		},
		"tencdns.net": {
			Name:    "腾讯云 CDN",
			Website: "https://cloud.tencent.com/product/cdn-scd",
		},
		"dayugslb.com": {
			Name:    "腾讯云大禹 BGP 高防",
			Website: "https://cloud.tencent.com/product/ddos-advanced",
		},
		"tcdnvod.com": {
			Name:    "腾讯云视频 CDN",
			Website: "https://lab.skk.moe/cdn",
		},
		"tdnsv5.com": {
			Name:    "腾讯云 CDN",
			Website: "https://cloud.tencent.com/product/cdn-scd",
		},
		"ksyuncdn.com": {
			Name:    "金山云 CDN",
			Website: "https://www.ksyun.com/post/product/CDN",
		},
		"ks-cdn.com": {
			Name:    "金山云 CDN",
			Website: "https://www.ksyun.com/post/product/CDN",
		},
		"ksyuncdn-k1.com": {
			Name:    "金山云 CDN",
			Website: "https://www.ksyun.com/post/product/CDN",
		},
		"netlify.com": {
			Name:    "Netlify",
			Website: "https://www.netlify.com",
		},
		"zeit.co": {
			Name:    "ZEIT Now Smart CDN",
			Website: "https://zeit.co",
		},
		"zeit-cdn.net": {
			Name:    "ZEIT Now Smart CDN",
			Website: "https://zeit.co",
		},
		"b-cdn.net": {
			Name:    "Bunny CDN",
			Website: "https://bunnycdn.com/",
		},
		"lsycdn.com": {
			Name:    "蓝视云 CDN",
			Website: "https://cloud.lsy.cn/",
		},
		"scsdns.com": {
			Name:    "逸云科技云加速 CDN",
			Website: "http://www.exclouds.com/navPage/wise",
		},
		"quic.cloud": {
			Name:    "QUIC.Cloud",
			Website: "https://quic.cloud/",
		},
		"flexbalancer.net": {
			Name:    "FlexBalancer - Smart Traffic Routing",
			Website: "https://perfops.net/flexbalancer",
		},
		"gcdn.co": {
			Name:    "G - Core Labs",
			Website: "https://gcorelabs.com/cdn/",
		},
		"sangfordns.com": {
			Name:    "深信服 AD 系列应用交付产品  单边加速解决方案",
			Website: "http://www.sangfor.com.cn/topic/2011adn/solutions5.html",
		},
		"stspg-customer.com": {
			Name:    "StatusPage.io",
			Website: "https://www.statuspage.io",
		},
		"turbobytes.net": {
			Name:    "TurboBytes Multi-CDN",
			Website: "https://www.turbobytes.com",
		},
		"turbobytes-cdn.com": {
			Name:    "TurboBytes Multi-CDN",
			Website: "https://www.turbobytes.com",
		},
		"att-dsa.net": {
			Name:    "AT&T Content Delivery Network",
			Website: "https://www.business.att.com/products/cdn.html",
		},
		"azioncdn.net": {
			Name:    "Azion Tech | Edge Computing PLatform",
			Website: "https://www.azion.com",
		},
		"belugacdn.com": {
			Name:    "BelugaCDN",
			Website: "https://www.belugacdn.com",
		},
		"cachefly.net": {
			Name:    "CacheFly CDN",
			Website: "https://www.cachefly.com/",
		},
		"inscname.net": {
			Name:    "Instart CDN",
			Website: "https://www.instart.com/products/web-performance/cdn",
		},
		"insnw.net": {
			Name:    "Instart CDN",
			Website: "https://www.instart.com/products/web-performance/cdn",
		},
		"internapcdn.net": {
			Name:    "Internap CDN",
			Website: "https://www.inap.com/network/content-delivery-network",
		},
		"footprint.net": {
			Name:    "CenturyLink CDN (原 Level 3)",
			Website: "https://www.centurylink.com/business/networking/cdn.html",
		},
		"llnwi.net": {
			Name:    "Limelight Network",
			Website: "https://www.limelight.com",
		},
		"llnwd.net": {
			Name:    "Limelight Network",
			Website: "https://www.limelight.com",
		},
		"unud.net": {
			Name:    "Limelight Network",
			Website: "https://www.limelight.com",
		},
		"lldns.net": {
			Name:    "Limelight Network",
			Website: "https://www.limelight.com",
		},
		"stackpathdns.com": {
			Name:    "Stackpath CDN",
			Website: "https://www.stackpath.com",
		},
		"stackpathcdn.com": {
			Name:    "Stackpath CDN",
			Website: "https://www.stackpath.com",
		},
		"mncdn.com": {
			Name:    "Medianova",
			Website: "https://www.medianova.com",
		},
		"rncdn1.com": {
			Name:    "Reelected Networks",
			Website: "https://reflected.net/globalcdn",
		},
		"simplecdn.net": {
			Name:    "Reelected Networks",
			Website: "https://reflected.net/globalcdn",
		},
		"swiftserve.com": {
			Name:    "Conversant - SwiftServe CDN",
			Website: "https://reflected.net/globalcdn",
		},
		"bitgravity.com": {
			Name:    "Tata communications CDN",
			Website: "https://cdn.tatacommunications.com",
		},
		"zenedge.net": {
			Name:    "Oracle Dyn Web Application Security suite (原 Zenedge CDN)",
			Website: "https://cdn.tatacommunications.com",
		},
		"biliapi.com": {
			Name:    "Bilibili 业务 GSLB",
			Website: "https://lab.skk.moe/cdn",
		},
		"hdslb.net": {
			Name:    "Bilibili 高可用负载均衡",
			Website: "https://github.com/bilibili/overlord",
		},
		"hdslb.com": {
			Name:    "Bilibili 高可用地域负载均衡",
			Website: "https://github.com/bilibili/overlord",
		},
		"xwaf.cn": {
			Name:    "极御云安全（浙江壹云云计算有限公司）",
			Website: "https://www.stopddos.cn",
		},
		"shifen.com": {
			Name:    "百度旗下业务地域负载均衡系统",
			Website: "https://lab.skk.moe/cdn",
		},
		"sinajs.cn": {
			Name:    "新浪静态域名",
			Website: "https://lab.skk.moe/cdn",
		},
		"tencent-cloud.net": {
			Name:    "腾讯旗下业务地域负载均衡系统",
			Website: "https://lab.skk.moe/cdn",
		},
		"elemecdn.com": {
			Name:    "饿了么静态域名与地域负载均衡",
			Website: "https://lab.skk.moe/cdn",
		},
		"sinaedge.com": {
			Name:    "新浪科技融合CDN负载均衡",
			Website: "https://lab.skk.moe/cdn",
		},
		"sina.com.cn": {
			Name:    "新浪科技融合CDN负载均衡",
			Website: "https://lab.skk.moe/cdn",
		},
		"sinacdn.com": {
			Name:    "新浪云 CDN",
			Website: "https://www.sinacloud.com/doc/sae/php/cdn.html",
		},
		"sinasws.com": {
			Name:    "新浪云 CDN",
			Website: "https://www.sinacloud.com/doc/sae/php/cdn.html",
		},
		"saebbs.com": {
			Name:    "新浪云 SAE 云引擎",
			Website: "https://www.sinacloud.com/doc/sae/php/cdn.html",
		},
		"websitecname.cn": {
			Name:    "美橙互联旗下建站之星",
			Website: "https://www.sitestar.cn",
		},
		"cdncenter.cn": {
			Name:    "美橙互联CDN",
			Website: "https://www.cndns.com",
		},
		"vhostgo.com": {
			Name:    "西部数码虚拟主机",
			Website: "https://www.west.cn",
		},
		"jsd.cc": {
			Name:    "上海云盾YUNDUN",
			Website: "https://www.yundun.com",
		},
		"powercdn.cn": {
			Name:    "动力在线CDN",
			Website: "http://www.powercdn.com",
		},
		"21vokglb.cn": {
			Name:    "世纪互联云快线业务",
			Website: "https://www.21vianet.com",
		},
		"21vianet.com.cn": {
			Name:    "世纪互联云快线业务",
			Website: "https://www.21vianet.com",
		},
		"21okglb.cn": {
			Name:    "世纪互联云快线业务",
			Website: "https://www.21vianet.com",
		},
		"21speedcdn.com": {
			Name:    "世纪互联云快线业务",
			Website: "https://www.21vianet.com",
		},
		"21cvcdn.com": {
			Name:    "世纪互联云快线业务",
			Website: "https://www.21vianet.com",
		},
		"okcdn.com": {
			Name:    "世纪互联云快线业务",
			Website: "https://www.21vianet.com",
		},
		"okglb.com": {
			Name:    "世纪互联云快线业务",
			Website: "https://www.21vianet.com",
		},
		"cdnetworks.net": {
			Name:    "北京同兴万点网络技术",
			Website: "http://www.txnetworks.cn/",
		},
		"txnetworks.cn": {
			Name:    "北京同兴万点网络技术",
			Website: "http://www.txnetworks.cn/",
		},
		"cdnnetworks.com": {
			Name:    "北京同兴万点网络技术",
			Website: "http://www.txnetworks.cn/",
		},
		"txcdn.cn": {
			Name:    "北京同兴万点网络技术",
			Website: "http://www.txnetworks.cn/",
		},
		"cdnunion.net": {
			Name:    "宝腾互联旗下上海万根网络（CDN 联盟）",
			Website: "http://www.cdnunion.com",
		},
		"cdnunion.com": {
			Name:    "宝腾互联旗下上海万根网络（CDN 联盟）",
			Website: "http://www.cdnunion.com",
		},
		"mygslb.com": {
			Name:    "宝腾互联旗下上海万根网络（YaoCDN）",
			Website: "http://www.vangen.cn",
		},
		"cdnudns.com": {
			Name:    "宝腾互联旗下上海万根网络（YaoCDN）",
			Website: "http://www.vangen.cn",
		},
		"sprycdn.com": {
			Name:    "宝腾互联旗下上海万根网络（YaoCDN）",
			Website: "http://www.vangen.cn",
		},
		"chuangcdn.com": {
			Name:    "创世云融合 CDN",
			Website: "https://www.chuangcache.com/index.html",
		},
		"aocde.com": {
			Name:    "创世云融合 CDN",
			Website: "https://www.chuangcache.com",
		},
		"ctxcdn.cn": {
			Name:    "中国电信天翼云CDN",
			Website: "https://www.ctyun.cn/product2/#/product/10027560",
		},
		"yfcdn.net": {
			Name:    "云帆加速CDN",
			Website: "https://www.yfcloud.com",
		},
		"mmycdn.cn": {
			Name:    "蛮蛮云 CDN（中联利信）",
			Website: "https://www.chinamaincloud.com/cloudDispatch.html",
		},
		"chinamaincloud.com": {
			Name:    "蛮蛮云 CDN（中联利信）",
			Website: "https://www.chinamaincloud.com/cloudDispatch.html",
		},
		"cnispgroup.com": {
			Name:    "中联数据（中联利信）",
			Website: "http://www.cnispgroup.com/",
		},
		"cdnle.com": {
			Name:    "新乐视云联（原乐视云）CDN",
			Website: "http://www.lecloud.com/zh-cn",
		},
		"gosuncdn.com": {
			Name:    "高升控股CDN技术",
			Website: "http://www.gosun.com",
		},
		"mmtrixopt.com": {
			Name:    "mmTrix性能魔方（高升控股旗下）",
			Website: "http://www.mmtrix.com",
		},
		"cloudfence.cn": {
			Name:    "蓝盾云CDN",
			Website: "https://www.cloudfence.cn/#/cloudWeb/yaq/yaqyfx",
		},
		"ngaagslb.cn": {
			Name:    "新流云（新流万联）",
			Website: "https://www.ngaa.com.cn",
		},
		"p2cdn.com": {
			Name:    "星域云P2P CDN",
			Website: "https://www.xycloud.com",
		},
		"00cdn.com": {
			Name:    "星域云P2P CDN",
			Website: "https://www.xycloud.com",
		},
		"sankuai.com": {
			Name:    "美团云（三快科技）负载均衡",
			Website: "https://www.mtyun.com",
		},
		"lccdn.org": {
			Name:    "领智云 CDN（杭州领智云画）",
			Website: "http://www.linkingcloud.com",
		},
		"nscloudwaf.com": {
			Name:    "绿盟云 WAF",
			Website: "https://cloud.nsfocus.com",
		},
		"2cname.com": {
			Name:    "网堤安全",
			Website: "https://www.ddos.com",
		},
		"ucloudgda.com": {
			Name:    "UCloud 罗马 Rome 全球网络加速",
			Website: "https://www.ucloud.cn/site/product/rome.html",
		},
		"google.com": {
			Name:    "Google Web 业务",
			Website: "https://lab.skk.moe/cdn",
		},
		"1e100.net": {
			Name:    "Google Web 业务",
			Website: "https://lab.skk.moe/cdn",
		},
		"ncname.com": {
			Name:    "NodeCache",
			Website: "https://www.nodecache.com",
		},
		"alipaydns.com": {
			Name:    "蚂蚁金服旗下业务地域负载均衡系统",
			Website: "https://lab.skk.moe/cdn/",
		},
		"wscloudcdn.com": {
			Name:    "全速云（网宿）CloudEdge 云加速",
			Website: "https://www.quansucloud.com/product.action?product.id=270",
		},
	},
	glob: map[*glob.Glob]CDN{
		&kunlunCDNGlob: {
			Name:    "阿里云 CDN",
			Website: "https://www.aliyun.com/product/cdn",
		},
	},
}
