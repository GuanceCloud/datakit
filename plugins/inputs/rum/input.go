// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package rum real user monitoring
package rum

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/gobwas/glob"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	dkhttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	ihttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/storage"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/workerpool"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"google.golang.org/protobuf/proto"
)

var (
	_ inputs.HTTPInput = &Input{}
	_ inputs.InputV2   = &Input{}
)

const (
	inputName         = "rum"
	ReplayFileMaxSize = 1 << 22 // 1024 * 1024 * 4  4Mib
	ProxyErrorHeader  = "X-Proxy-Error"
	// nolint: lll
	sampleConfig = `
[[inputs.rum]]
  ## profile Agent endpoints register by version respectively.
  ## Endpoints can be skipped listen by remove them from the list.
  ## Default value set as below. DO NOT MODIFY THESE ENDPOINTS if not necessary.
  endpoints = ["/v1/write/rum"]

  ## use to upload rum screenshot,html,etc...
  session_replay_endpoints = ["/v1/write/rum/replay"]

  ## Android command-line-tools HOME
  android_cmdline_home = "/usr/local/datakit/data/rum/tools/cmdline-tools"

  ## proguard HOME
  proguard_home = "/usr/local/datakit/data/rum/tools/proguard"

  ## android-ndk HOME
  ndk_home = "/usr/local/datakit/data/rum/tools/android-ndk"

  ## atos or atosl bin path
  ## for macOS datakit use the built-in tool atos default
  ## for Linux there are several tools that can be used to instead of macOS atos partially,
  ## such as https://github.com/everettjf/atosl-rs
  atos_bin_path = "/usr/local/datakit/data/rum/tools/atosl"

  ## Threads config controls how many goroutines an agent cloud start to handle HTTP request.
  ## buffer is the size of jobs' buffering of worker channel.
  ## threads is the total number fo goroutines at running time.
  # [inputs.rum.threads]
    # buffer = 100
    # threads = 8

  ## Storage config a local storage space in hard dirver to cache trace data.
  ## path is the local file path used to cache data.
  ## capacity is total space size(MB) used to store data.
  # [inputs.rum.storage]
    # path = "./rum_storage"
    # capacity = 5120

  # Provide a list to resolve CDN of your static resource.
  # Below is the Datakit default built-in CDN list, you can uncomment that and change it to your cdn list, 
  # it's a JSON array like: [{"domain": "CDN domain", "name": "CDN human readable name", "website": "CDN official website"},...],
  # domain field value can contains '*' as wildcard, for example: "kunlun*.com", 
  # it will match "kunluna.com", "kunlunab.com" and "kunlunabc.com" but not "kunlunab.c.com".
  # cdn_map = '[{"domain":"15cdn.com","name":"腾正安全加速(原 15CDN)","website":"https://www.15cdn.com"},{"domain":"tzcdn.cn","name":"腾正安全加速(原 15CDN)","website":"https://www.15cdn.com"},{"domain":"cedexis.net","name":"Cedexis GSLB","website":"https://www.cedexis.com/"},{"domain":"cdxcn.cn","name":"Cedexis GSLB (For China)","website":"https://www.cedexis.com/"},{"domain":"qhcdn.com","name":"360 云 CDN (由奇安信运营)","website":"https://cloud.360.cn/doc?name=cdn"},{"domain":"qh-cdn.com","name":"360 云 CDN (由奇虎 360 运营)","website":"https://cloud.360.cn/doc?name=cdn"},{"domain":"qihucdn.com","name":"360 云 CDN (由奇虎 360 运营)","website":"https://cloud.360.cn/doc?name=cdn"},{"domain":"360cdn.com","name":"360 云 CDN (由奇虎 360 运营)","website":"https://cloud.360.cn/doc?name=cdn"},{"domain":"360cloudwaf.com","name":"奇安信网站卫士","website":"https://wangzhan.qianxin.com"},{"domain":"360anyu.com","name":"奇安信网站卫士","website":"https://wangzhan.qianxin.com"},{"domain":"360safedns.com","name":"奇安信网站卫士","website":"https://wangzhan.qianxin.com"},{"domain":"360wzws.com","name":"奇安信网站卫士","website":"https://wangzhan.qianxin.com"},{"domain":"akamai.net","name":"Akamai CDN","website":"https://www.akamai.com"},{"domain":"akamaiedge.net","name":"Akamai CDN","website":"https://www.akamai.com"},{"domain":"ytcdn.net","name":"Akamai CDN","website":"https://www.akamai.com"},{"domain":"edgesuite.net","name":"Akamai CDN","website":"https://www.akamai.com"},{"domain":"akamaitech.net","name":"Akamai CDN","website":"https://www.akamai.com"},{"domain":"akamaitechnologies.com","name":"Akamai CDN","website":"https://www.akamai.com"},{"domain":"edgekey.net","name":"Akamai CDN","website":"https://www.akamai.com"},{"domain":"tl88.net","name":"易通锐进(Akamai 中国)由网宿承接","website":"https://www.akamai.com"},{"domain":"cloudfront.net","name":"AWS CloudFront","website":"https://aws.amazon.com/cn/cloudfront/"},{"domain":"worldcdn.net","name":"CDN.NET","website":"https://cdn.net"},{"domain":"worldssl.net","name":"CDN.NET / CDNSUN / ONAPP","website":"https://cdn.net"},{"domain":"cdn77.org","name":"CDN77","website":"https://www.cdn77.com/"},{"domain":"panthercdn.com","name":"CDNetworks","website":"https://www.cdnetworks.com"},{"domain":"cdnga.net","name":"CDNetworks","website":"https://www.cdnetworks.com"},{"domain":"cdngc.net","name":"CDNetworks","website":"https://www.cdnetworks.com"},{"domain":"gccdn.net","name":"CDNetworks","website":"https://www.cdnetworks.com"},{"domain":"gccdn.cn","name":"CDNetworks","website":"https://www.cdnetworks.com"},{"domain":"akamaized.net","name":"Akamai CDN","website":"https://www.akamai.com"},{"domain":"126.net","name":"网易云 CDN","website":"https://www.163yun.com/product/cdn"},{"domain":"163jiasu.com","name":"网易云 CDN","website":"https://www.163yun.com/product/cdn"},{"domain":"amazonaws.com","name":"AWS Cloud","website":"https://aws.amazon.com/cn/cloudfront/"},{"domain":"cdn77.net","name":"CDN77","website":"https://www.cdn77.com/"},{"domain":"cdnify.io","name":"CDNIFY","website":"https://cdnify.com"},{"domain":"cdnsun.net","name":"CDNSUN","website":"https://cdnsun.com"},{"domain":"bdydns.com","name":"百度云 CDN","website":"https://cloud.baidu.com/product/cdn.html"},{"domain":"ccgslb.com.cn","name":"蓝汛 CDN","website":"https://cn.chinacache.com/"},{"domain":"ccgslb.net","name":"蓝汛 CDN","website":"https://cn.chinacache.com/"},{"domain":"ccgslb.com","name":"蓝汛 CDN","website":"https://cn.chinacache.com/"},{"domain":"ccgslb.cn","name":"蓝汛 CDN","website":"https://cn.chinacache.com/"},{"domain":"c3cache.net","name":"蓝汛 CDN","website":"https://cn.chinacache.com/"},{"domain":"c3dns.net","name":"蓝汛 CDN","website":"https://cn.chinacache.com/"},{"domain":"chinacache.net","name":"蓝汛 CDN","website":"https://cn.chinacache.com/"},{"domain":"wswebcdn.com","name":"网宿 CDN","website":"https://www.wangsu.com/"},{"domain":"lxdns.com","name":"网宿 CDN","website":"https://www.wangsu.com/"},{"domain":"wswebpic.com","name":"网宿 CDN","website":"https://www.wangsu.com/"},{"domain":"cloudflare.net","name":"Cloudflare","website":"https://www.cloudflare.com"},{"domain":"akadns.net","name":"Akamai CDN","website":"https://www.akamai.com"},{"domain":"chinanetcenter.com","name":"网宿 CDN","website":"https://www.wangsu.com"},{"domain":"customcdn.com.cn","name":"网宿 CDN","website":"https://www.wangsu.com"},{"domain":"customcdn.cn","name":"网宿 CDN","website":"https://www.wangsu.com"},{"domain":"51cdn.com","name":"网宿 CDN","website":"https://www.wangsu.com"},{"domain":"wscdns.com","name":"网宿 CDN","website":"https://www.wangsu.com"},{"domain":"cdn20.com","name":"网宿 CDN","website":"https://www.wangsu.com"},{"domain":"wsdvs.com","name":"网宿 CDN","website":"https://www.wangsu.com"},{"domain":"wsglb0.com","name":"网宿 CDN","website":"https://www.wangsu.com"},{"domain":"speedcdns.com","name":"网宿 CDN","website":"https://www.wangsu.com"},{"domain":"wtxcdn.com","name":"网宿 CDN","website":"https://www.wangsu.com"},{"domain":"wsssec.com","name":"网宿 WAF CDN","website":"https://www.wangsu.com"},{"domain":"fastly.net","name":"Fastly","website":"https://www.fastly.com"},{"domain":"fastlylb.net","name":"Fastly","website":"https://www.fastly.com/"},{"domain":"hwcdn.net","name":"Stackpath (原 Highwinds)","website":"https://www.stackpath.com/highwinds"},{"domain":"incapdns.net","name":"Incapsula CDN","website":"https://www.incapsula.com"},{"domain":"kxcdn.com.","name":"KeyCDN","website":"https://www.keycdn.com/"},{"domain":"lswcdn.net","name":"LeaseWeb CDN","website":"https://www.leaseweb.com/cdn"},{"domain":"mwcloudcdn.com","name":"QUANTIL (网宿)","website":"https://www.quantil.com/"},{"domain":"mwcname.com","name":"QUANTIL (网宿)","website":"https://www.quantil.com/"},{"domain":"azureedge.net","name":"Microsoft Azure CDN","website":"https://azure.microsoft.com/en-us/services/cdn/"},{"domain":"msecnd.net","name":"Microsoft Azure CDN","website":"https://azure.microsoft.com/en-us/services/cdn/"},{"domain":"mschcdn.com","name":"Microsoft Azure CDN","website":"https://azure.microsoft.com/en-us/services/cdn/"},{"domain":"v0cdn.net","name":"Microsoft Azure CDN","website":"https://azure.microsoft.com/en-us/services/cdn/"},{"domain":"azurewebsites.net","name":"Microsoft Azure App Service","website":"https://azure.microsoft.com/en-us/services/app-service/"},{"domain":"azurewebsites.windows.net","name":"Microsoft Azure App Service","website":"https://azure.microsoft.com/en-us/services/app-service/"},{"domain":"trafficmanager.net","name":"Microsoft Azure Traffic Manager","website":"https://azure.microsoft.com/en-us/services/traffic-manager/"},{"domain":"cloudapp.net","name":"Microsoft Azure","website":"https://azure.microsoft.com"},{"domain":"chinacloudsites.cn","name":"世纪互联旗下上海蓝云(承载 Azure 中国)","website":"https://www.21vbluecloud.com/"},{"domain":"spdydns.com","name":"云端智度融合 CDN","website":"https://www.isurecloud.net/index.html"},{"domain":"jiashule.com","name":"知道创宇云安全加速乐CDN","website":"https://www.yunaq.com/jsl/"},{"domain":"jiasule.org","name":"知道创宇云安全加速乐CDN","website":"https://www.yunaq.com/jsl/"},{"domain":"365cyd.cn","name":"知道创宇云安全创宇盾(政务专用)","website":"https://www.yunaq.com/cyd/"},{"domain":"huaweicloud.com","name":"华为云WAF高防云盾","website":"https://www.huaweicloud.com/product/aad.html"},{"domain":"cdnhwc1.com","name":"华为云 CDN","website":"https://www.huaweicloud.com/product/cdn.html"},{"domain":"cdnhwc2.com","name":"华为云 CDN","website":"https://www.huaweicloud.com/product/cdn.html"},{"domain":"cdnhwc3.com","name":"华为云 CDN","website":"https://www.huaweicloud.com/product/cdn.html"},{"domain":"dnion.com","name":"帝联科技","website":"http://www.dnion.com/"},{"domain":"ewcache.com","name":"帝联科技","website":"http://www.dnion.com/"},{"domain":"globalcdn.cn","name":"帝联科技","website":"http://www.dnion.com/"},{"domain":"tlgslb.com","name":"帝联科技","website":"http://www.dnion.com/"},{"domain":"fastcdn.com","name":"帝联科技","website":"http://www.dnion.com/"},{"domain":"flxdns.com","name":"帝联科技","website":"http://www.dnion.com/"},{"domain":"dlgslb.cn","name":"帝联科技","website":"http://www.dnion.com/"},{"domain":"newdefend.cn","name":"牛盾云安全","website":"https://www.newdefend.com"},{"domain":"ffdns.net","name":"CloudXNS","website":"https://www.cloudxns.net"},{"domain":"aocdn.com","name":"可靠云 CDN (贴图库)","website":"http://www.kekaoyun.com/"},{"domain":"bsgslb.cn","name":"白山云 CDN","website":"https://zh.baishancloud.com/"},{"domain":"qingcdn.com","name":"白山云 CDN","website":"https://zh.baishancloud.com/"},{"domain":"bsclink.cn","name":"白山云 CDN","website":"https://zh.baishancloud.com/"},{"domain":"trpcdn.net","name":"白山云 CDN","website":"https://zh.baishancloud.com/"},{"domain":"anquan.io","name":"牛盾云安全","website":"https://www.newdefend.com"},{"domain":"cloudglb.com","name":"快网 CDN","website":"http://www.fastweb.com.cn/"},{"domain":"fastweb.com","name":"快网 CDN","website":"http://www.fastweb.com.cn/"},{"domain":"fastwebcdn.com","name":"快网 CDN","website":"http://www.fastweb.com.cn/"},{"domain":"cloudcdn.net","name":"快网 CDN","website":"http://www.fastweb.com.cn/"},{"domain":"fwcdn.com","name":"快网 CDN","website":"http://www.fastweb.com.cn/"},{"domain":"fwdns.net","name":"快网 CDN","website":"http://www.fastweb.com.cn/"},{"domain":"hadns.net","name":"快网 CDN","website":"http://www.fastweb.com.cn/"},{"domain":"hacdn.net","name":"快网 CDN","website":"http://www.fastweb.com.cn/"},{"domain":"cachecn.com","name":"快网 CDN","website":"http://www.fastweb.com.cn/"},{"domain":"qingcache.com","name":"青云 CDN","website":"https://www.qingcloud.com/products/cdn/"},{"domain":"qingcloud.com","name":"青云 CDN","website":"https://www.qingcloud.com/products/cdn/"},{"domain":"frontwize.com","name":"青云 CDN","website":"https://www.qingcloud.com/products/cdn/"},{"domain":"msscdn.com","name":"美团云 CDN","website":"https://www.mtyun.com/product/cdn"},{"domain":"800cdn.com","name":"西部数码","website":"https://www.west.cn"},{"domain":"tbcache.com","name":"阿里云 CDN","website":"https://www.aliyun.com/product/cdn"},{"domain":"aliyun-inc.com","name":"阿里云 CDN","website":"https://www.aliyun.com/product/cdn"},{"domain":"aliyuncs.com","name":"阿里云 CDN","website":"https://www.aliyun.com/product/cdn"},{"domain":"alikunlun.net","name":"阿里云 CDN","website":"https://www.aliyun.com/product/cdn"},{"domain":"alikunlun.com","name":"阿里云 CDN","website":"https://www.aliyun.com/product/cdn"},{"domain":"alicdn.com","name":"阿里云 CDN","website":"https://www.aliyun.com/product/cdn"},{"domain":"aligaofang.com","name":"阿里云盾高防","website":"https://www.aliyun.com/product/ddos"},{"domain":"yundunddos.com","name":"阿里云盾高防","website":"https://www.aliyun.com/product/ddos"},{"domain":"kunlun*.com","name":"阿里云 CDN","website":"https://www.aliyun.com/product/cdn"},{"domain":"cdngslb.com","name":"阿里云 CDN","website":"https://www.aliyun.com/product/cdn"},{"domain":"yunjiasu-cdn.net","name":"百度云加速","website":"https://su.baidu.com"},{"domain":"momentcdn.com","name":"魔门云 CDN","website":"https://www.cachemoment.com"},{"domain":"aicdn.com","name":"又拍云","website":"https://www.upyun.com"},{"domain":"qbox.me","name":"七牛云","website":"https://www.qiniu.com"},{"domain":"qiniu.com","name":"七牛云","website":"https://www.qiniu.com"},{"domain":"qiniudns.com","name":"七牛云","website":"https://www.qiniu.com"},{"domain":"jcloudcs.com","name":"京东云 CDN","website":"https://www.jdcloud.com/cn/products/cdn"},{"domain":"jdcdn.com","name":"京东云 CDN","website":"https://www.jdcloud.com/cn/products/cdn"},{"domain":"qianxun.com","name":"京东云 CDN","website":"https://www.jdcloud.com/cn/products/cdn"},{"domain":"jcloudlb.com","name":"京东云 CDN","website":"https://www.jdcloud.com/cn/products/cdn"},{"domain":"jcloud-cdn.com","name":"京东云 CDN","website":"https://www.jdcloud.com/cn/products/cdn"},{"domain":"maoyun.tv","name":"猫云融合 CDN","website":"https://www.maoyun.com/"},{"domain":"maoyundns.com","name":"猫云融合 CDN","website":"https://www.maoyun.com/"},{"domain":"xgslb.net","name":"WebLuker (蓝汛)","website":"http://www.webluker.com"},{"domain":"ucloud.cn","name":"UCloud CDN","website":"https://www.ucloud.cn/site/product/ucdn.html"},{"domain":"ucloud.com.cn","name":"UCloud CDN","website":"https://www.ucloud.cn/site/product/ucdn.html"},{"domain":"cdndo.com","name":"UCloud CDN","website":"https://www.ucloud.cn/site/product/ucdn.html"},{"domain":"zenlogic.net","name":"Zenlayer CDN","website":"https://www.zenlayer.com"},{"domain":"ogslb.com","name":"Zenlayer CDN","website":"https://www.zenlayer.com"},{"domain":"uxengine.net","name":"Zenlayer CDN","website":"https://www.zenlayer.com"},{"domain":"tan14.net","name":"TAN14 CDN","website":"http://www.tan14.cn/"},{"domain":"verycloud.cn","name":"VeryCloud 云分发","website":"https://www.verycloud.cn/"},{"domain":"verycdn.net","name":"VeryCloud 云分发","website":"https://www.verycloud.cn/"},{"domain":"verygslb.com","name":"VeryCloud 云分发","website":"https://www.verycloud.cn/"},{"domain":"xundayun.cn","name":"SpeedyCloud CDN","website":"https://www.speedycloud.cn/zh/Products/CDN/CloudDistribution.html"},{"domain":"xundayun.com","name":"SpeedyCloud CDN","website":"https://www.speedycloud.cn/zh/Products/CDN/CloudDistribution.html"},{"domain":"speedycloud.cc","name":"SpeedyCloud CDN","website":"https://www.speedycloud.cn/zh/Products/CDN/CloudDistribution.html"},{"domain":"mucdn.net","name":"Verizon CDN (Edgecast)","website":"https://www.verizondigitalmedia.com/platform/edgecast-cdn/"},{"domain":"nucdn.net","name":"Verizon CDN (Edgecast)","website":"https://www.verizondigitalmedia.com/platform/edgecast-cdn/"},{"domain":"alphacdn.net","name":"Verizon CDN (Edgecast)","website":"https://www.verizondigitalmedia.com/platform/edgecast-cdn/"},{"domain":"systemcdn.net","name":"Verizon CDN (Edgecast)","website":"https://www.verizondigitalmedia.com/platform/edgecast-cdn/"},{"domain":"edgecastcdn.net","name":"Verizon CDN (Edgecast)","website":"https://www.verizondigitalmedia.com/platform/edgecast-cdn/"},{"domain":"zetacdn.net","name":"Verizon CDN (Edgecast)","website":"https://www.verizondigitalmedia.com/platform/edgecast-cdn/"},{"domain":"coding.io","name":"Coding Pages","website":"https://coding.net/pages"},{"domain":"coding.me","name":"Coding Pages","website":"https://coding.net/pages"},{"domain":"gitlab.io","name":"GitLab Pages","website":"https://docs.gitlab.com/ee/user/project/pages/"},{"domain":"github.io","name":"GitHub Pages","website":"https://pages.github.com/"},{"domain":"herokuapp.com","name":"Heroku SaaS","website":"https://www.heroku.com"},{"domain":"googleapis.com","name":"Google Cloud Storage","website":"https://cloud.google.com/storage/"},{"domain":"netdna.com","name":"Stackpath (原 MaxCDN)","website":"https://www.stackpath.com/maxcdn/"},{"domain":"netdna-cdn.com","name":"Stackpath (原 MaxCDN)","website":"https://www.stackpath.com/maxcdn/"},{"domain":"netdna-ssl.com","name":"Stackpath (原 MaxCDN)","website":"https://www.stackpath.com/maxcdn/"},{"domain":"cdntip.com","name":"腾讯云 CDN","website":"https://cloud.tencent.com/product/cdn-scd"},{"domain":"dnsv1.com","name":"腾讯云 CDN","website":"https://cloud.tencent.com/product/cdn-scd"},{"domain":"tencdns.net","name":"腾讯云 CDN","website":"https://cloud.tencent.com/product/cdn-scd"},{"domain":"dayugslb.com","name":"腾讯云大禹 BGP 高防","website":"https://cloud.tencent.com/product/ddos-advanced"},{"domain":"tcdnvod.com","name":"腾讯云视频 CDN","website":"https://lab.skk.moe/cdn"},{"domain":"tdnsv5.com","name":"腾讯云 CDN","website":"https://cloud.tencent.com/product/cdn-scd"},{"domain":"ksyuncdn.com","name":"金山云 CDN","website":"https://www.ksyun.com/post/product/CDN"},{"domain":"ks-cdn.com","name":"金山云 CDN","website":"https://www.ksyun.com/post/product/CDN"},{"domain":"ksyuncdn-k1.com","name":"金山云 CDN","website":"https://www.ksyun.com/post/product/CDN"},{"domain":"netlify.com","name":"Netlify","website":"https://www.netlify.com"},{"domain":"zeit.co","name":"ZEIT Now Smart CDN","website":"https://zeit.co"},{"domain":"zeit-cdn.net","name":"ZEIT Now Smart CDN","website":"https://zeit.co"},{"domain":"b-cdn.net","name":"Bunny CDN","website":"https://bunnycdn.com/"},{"domain":"lsycdn.com","name":"蓝视云 CDN","website":"https://cloud.lsy.cn/"},{"domain":"scsdns.com","name":"逸云科技云加速 CDN","website":"http://www.exclouds.com/navPage/wise"},{"domain":"quic.cloud","name":"QUIC.Cloud","website":"https://quic.cloud/"},{"domain":"flexbalancer.net","name":"FlexBalancer - Smart Traffic Routing","website":"https://perfops.net/flexbalancer"},{"domain":"gcdn.co","name":"G - Core Labs","website":"https://gcorelabs.com/cdn/"},{"domain":"sangfordns.com","name":"深信服 AD 系列应用交付产品  单边加速解决方案","website":"http://www.sangfor.com.cn/topic/2011adn/solutions5.html"},{"domain":"stspg-customer.com","name":"StatusPage.io","website":"https://www.statuspage.io"},{"domain":"turbobytes.net","name":"TurboBytes Multi-CDN","website":"https://www.turbobytes.com"},{"domain":"turbobytes-cdn.com","name":"TurboBytes Multi-CDN","website":"https://www.turbobytes.com"},{"domain":"att-dsa.net","name":"AT&T Content Delivery Network","website":"https://www.business.att.com/products/cdn.html"},{"domain":"azioncdn.net","name":"Azion Tech | Edge Computing PLatform","website":"https://www.azion.com"},{"domain":"belugacdn.com","name":"BelugaCDN","website":"https://www.belugacdn.com"},{"domain":"cachefly.net","name":"CacheFly CDN","website":"https://www.cachefly.com/"},{"domain":"inscname.net","name":"Instart CDN","website":"https://www.instart.com/products/web-performance/cdn"},{"domain":"insnw.net","name":"Instart CDN","website":"https://www.instart.com/products/web-performance/cdn"},{"domain":"internapcdn.net","name":"Internap CDN","website":"https://www.inap.com/network/content-delivery-network"},{"domain":"footprint.net","name":"CenturyLink CDN (原 Level 3)","website":"https://www.centurylink.com/business/networking/cdn.html"},{"domain":"llnwi.net","name":"Limelight Network","website":"https://www.limelight.com"},{"domain":"llnwd.net","name":"Limelight Network","website":"https://www.limelight.com"},{"domain":"unud.net","name":"Limelight Network","website":"https://www.limelight.com"},{"domain":"lldns.net","name":"Limelight Network","website":"https://www.limelight.com"},{"domain":"stackpathdns.com","name":"Stackpath CDN","website":"https://www.stackpath.com"},{"domain":"stackpathcdn.com","name":"Stackpath CDN","website":"https://www.stackpath.com"},{"domain":"mncdn.com","name":"Medianova","website":"https://www.medianova.com"},{"domain":"rncdn1.com","name":"Relected Networks","website":"https://reflected.net/globalcdn"},{"domain":"simplecdn.net","name":"Relected Networks","website":"https://reflected.net/globalcdn"},{"domain":"swiftserve.com","name":"Conversant - SwiftServe CDN","website":"https://reflected.net/globalcdn"},{"domain":"bitgravity.com","name":"Tata communications CDN","website":"https://cdn.tatacommunications.com"},{"domain":"zenedge.net","name":"Oracle Dyn Web Application Security suite (原 Zenedge CDN)","website":"https://cdn.tatacommunications.com"},{"domain":"biliapi.com","name":"Bilibili 业务 GSLB","website":"https://lab.skk.moe/cdn"},{"domain":"hdslb.net","name":"Bilibili 高可用负载均衡","website":"https://github.com/bilibili/overlord"},{"domain":"hdslb.com","name":"Bilibili 高可用地域负载均衡","website":"https://github.com/bilibili/overlord"},{"domain":"xwaf.cn","name":"极御云安全(浙江壹云云计算有限公司)","website":"https://www.stopddos.cn"},{"domain":"shifen.com","name":"百度旗下业务地域负载均衡系统","website":"https://lab.skk.moe/cdn"},{"domain":"sinajs.cn","name":"新浪静态域名","website":"https://lab.skk.moe/cdn"},{"domain":"tencent-cloud.net","name":"腾讯旗下业务地域负载均衡系统","website":"https://lab.skk.moe/cdn"},{"domain":"elemecdn.com","name":"饿了么静态域名与地域负载均衡","website":"https://lab.skk.moe/cdn"},{"domain":"sinaedge.com","name":"新浪科技融合CDN负载均衡","website":"https://lab.skk.moe/cdn"},{"domain":"sina.com.cn","name":"新浪科技融合CDN负载均衡","website":"https://lab.skk.moe/cdn"},{"domain":"sinacdn.com","name":"新浪云 CDN","website":"https://www.sinacloud.com/doc/sae/php/cdn.html"},{"domain":"sinasws.com","name":"新浪云 CDN","website":"https://www.sinacloud.com/doc/sae/php/cdn.html"},{"domain":"saebbs.com","name":"新浪云 SAE 云引擎","website":"https://www.sinacloud.com/doc/sae/php/cdn.html"},{"domain":"websitecname.cn","name":"美橙互联旗下建站之星","website":"https://www.sitestar.cn"},{"domain":"cdncenter.cn","name":"美橙互联CDN","website":"https://www.cndns.com"},{"domain":"vhostgo.com","name":"西部数码虚拟主机","website":"https://www.west.cn"},{"domain":"jsd.cc","name":"上海云盾YUNDUN","website":"https://www.yundun.com"},{"domain":"powercdn.cn","name":"动力在线CDN","website":"http://www.powercdn.com"},{"domain":"21vokglb.cn","name":"世纪互联云快线业务","website":"https://www.21vianet.com"},{"domain":"21vianet.com.cn","name":"世纪互联云快线业务","website":"https://www.21vianet.com"},{"domain":"21okglb.cn","name":"世纪互联云快线业务","website":"https://www.21vianet.com"},{"domain":"21speedcdn.com","name":"世纪互联云快线业务","website":"https://www.21vianet.com"},{"domain":"21cvcdn.com","name":"世纪互联云快线业务","website":"https://www.21vianet.com"},{"domain":"okcdn.com","name":"世纪互联云快线业务","website":"https://www.21vianet.com"},{"domain":"okglb.com","name":"世纪互联云快线业务","website":"https://www.21vianet.com"},{"domain":"cdnetworks.net","name":"北京同兴万点网络技术","website":"http://www.txnetworks.cn/"},{"domain":"txnetworks.cn","name":"北京同兴万点网络技术","website":"http://www.txnetworks.cn/"},{"domain":"cdnnetworks.com","name":"北京同兴万点网络技术","website":"http://www.txnetworks.cn/"},{"domain":"txcdn.cn","name":"北京同兴万点网络技术","website":"http://www.txnetworks.cn/"},{"domain":"cdnunion.net","name":"宝腾互联旗下上海万根网络(CDN 联盟)","website":"http://www.cdnunion.com"},{"domain":"cdnunion.com","name":"宝腾互联旗下上海万根网络(CDN 联盟)","website":"http://www.cdnunion.com"},{"domain":"mygslb.com","name":"宝腾互联旗下上海万根网络(YaoCDN)","website":"http://www.vangen.cn"},{"domain":"cdnudns.com","name":"宝腾互联旗下上海万根网络(YaoCDN)","website":"http://www.vangen.cn"},{"domain":"sprycdn.com","name":"宝腾互联旗下上海万根网络(YaoCDN)","website":"http://www.vangen.cn"},{"domain":"chuangcdn.com","name":"创世云融合 CDN","website":"https://www.chuangcache.com/index.html"},{"domain":"aocde.com","name":"创世云融合 CDN","website":"https://www.chuangcache.com"},{"domain":"ctxcdn.cn","name":"中国电信天翼云CDN","website":"https://www.ctyun.cn/product2/#/product/10027560"},{"domain":"yfcdn.net","name":"云帆加速CDN","website":"https://www.yfcloud.com"},{"domain":"mmycdn.cn","name":"蛮蛮云 CDN(中联利信)","website":"https://www.chinamaincloud.com/cloudDispatch.html"},{"domain":"chinamaincloud.com","name":"蛮蛮云 CDN(中联利信)","website":"https://www.chinamaincloud.com/cloudDispatch.html"},{"domain":"cnispgroup.com","name":"中联数据(中联利信)","website":"http://www.cnispgroup.com/"},{"domain":"cdnle.com","name":"新乐视云联(原乐视云)CDN","website":"http://www.lecloud.com/zh-cn"},{"domain":"gosuncdn.com","name":"高升控股CDN技术","website":"http://www.gosun.com"},{"domain":"mmtrixopt.com","name":"mmTrix性能魔方(高升控股旗下)","website":"http://www.mmtrix.com"},{"domain":"cloudfence.cn","name":"蓝盾云CDN","website":"https://www.cloudfence.cn/#/cloudWeb/yaq/yaqyfx"},{"domain":"ngaagslb.cn","name":"新流云(新流万联)","website":"https://www.ngaa.com.cn"},{"domain":"p2cdn.com","name":"星域云P2P CDN","website":"https://www.xycloud.com"},{"domain":"00cdn.com","name":"星域云P2P CDN","website":"https://www.xycloud.com"},{"domain":"sankuai.com","name":"美团云(三快科技)负载均衡","website":"https://www.mtyun.com"},{"domain":"lccdn.org","name":"领智云 CDN(杭州领智云画)","website":"http://www.linkingcloud.com"},{"domain":"nscloudwaf.com","name":"绿盟云 WAF","website":"https://cloud.nsfocus.com"},{"domain":"2cname.com","name":"网堤安全","website":"https://www.ddos.com"},{"domain":"ucloudgda.com","name":"UCloud 罗马 Rome 全球网络加速","website":"https://www.ucloud.cn/site/product/rome.html"},{"domain":"google.com","name":"Google Web 业务","website":"https://lab.skk.moe/cdn"},{"domain":"1e100.net","name":"Google Web 业务","website":"https://lab.skk.moe/cdn"},{"domain":"ncname.com","name":"NodeCache","website":"https://www.nodecache.com"},{"domain":"alipaydns.com","name":"蚂蚁金服旗下业务地域负载均衡系统","website":"https://lab.skk.moe/cdn/"},{"domain":"wscloudcdn.com","name":"全速云(网宿)CloudEdge 云加速","website":"https://www.quansucloud.com/product.action?product.id=270"}]'
`
)

const (
	Session  = "session"
	View     = "view"
	Resource = "resource"
	Action   = "action"
	LongTask = "long task"
	Error    = "error"
)

var (
	log        = logger.DefaultSLogger(inputName)
	wkpool     *workerpool.WorkerPool
	localCache *storage.Storage

	CDNList = &CDNPool{
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
)

var kunlunCDNGlob = glob.MustCompile(`*.kunlun*.com`)

type Input struct {
	Endpoints              []string                     `toml:"endpoints"`
	SessionReplayEndpoints []string                     `toml:"session_replay_endpoints"`
	JavaHome               string                       `toml:"java_home"`
	AndroidCmdLineHome     string                       `toml:"android_cmdline_home"`
	ProguardHome           string                       `toml:"proguard_home"`
	NDKHome                string                       `toml:"ndk_home"`
	AtosBinPath            string                       `toml:"atos_bin_path"`
	WPConfig               *workerpool.WorkerPoolConfig `toml:"threads"`
	LocalCacheConfig       *storage.StorageConfig       `toml:"storage"`
	CDNMap                 string                       `toml:"cdn_map"`
}

type CDN struct {
	Domain  string `json:"domain"`
	Name    string `json:"name"`
	Website string `json:"website"`
}

type CDNPool struct {
	literal map[string]CDN
	glob    map[*glob.Glob]CDN
}

var errLimitReader = errors.New("limit reader err")

type limitReader struct {
	r io.ReadCloser
}

func newLimitReader(r io.ReadCloser, max int64) io.ReadCloser {
	return &limitReader{
		r: http.MaxBytesReader(nil, r, max),
	}
}

func (l *limitReader) Read(p []byte) (int, error) {
	n, err := l.r.Read(p)
	if err != nil {
		if err == io.EOF { //nolint:errorlint
			return n, err
		}
		// wrap the errLimitReader
		return n, fmt.Errorf("%w: %s", errLimitReader, err)
	}
	return n, nil
}

func (l *limitReader) Close() error {
	return l.r.Close()
}

func (*Input) Catalog() string { return inputName }

func (*Input) AvailableArchs() []string { return datakit.AllOS }

func (*Input) SampleConfig() string { return sampleConfig }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&trace.TraceMeasurement{Name: inputName}}
}

func replayUploadHandler() (*httputil.ReverseProxy, error) {
	endpoints := config.Cfg.DataWay.GetAvailableEndpoints()

	if len(endpoints) == 0 {
		return nil, fmt.Errorf("no available dataway endpoint now")
	}

	var (
		validURL *url.URL
		lastErr  error
	)
	for _, ep := range endpoints {
		replayURL := ep.GetCategoryURL()[datakit.SessionReplayUpload]
		if replayURL == "" {
			lastErr = fmt.Errorf("empty category url")
			continue
		}
		parsedURL, err := url.Parse(replayURL)
		if err == nil {
			validURL = parsedURL
			break
		}
		lastErr = err
	}

	if validURL == nil {
		if lastErr != nil {
			return nil, lastErr
		}
		return nil, fmt.Errorf("no available dataway endpoint")
	}

	return &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			if req.ContentLength > ReplayFileMaxSize {
				req.URL = nil // this will trigger a proxy err, and let request complete earlier
				req.Header.Set(ProxyErrorHeader, fmt.Sprintf("request body size [%d] exceeds the limit", req.ContentLength))
				return
			}

			req.URL = validURL
			req.Host = validURL.Host
			req.Body = newLimitReader(req.Body, ReplayFileMaxSize)
		},

		ErrorHandler: func(w http.ResponseWriter, req *http.Request, err error) {
			proxyErr := req.Header.Get(ProxyErrorHeader)
			if proxyErr != "" {
				log.Errorf("proxy error: %s", proxyErr)
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(proxyErr))
				return
			}
			if errors.Is(err, errLimitReader) {
				log.Errorf("request body is too large: %s", err)
				w.WriteHeader(http.StatusBadRequest)
			} else {
				log.Errorf("other rum replay err: %s", err)
				w.WriteHeader(http.StatusBadGateway)
			}
			_, _ = w.Write([]byte(err.Error()))
		},
	}, nil
}

func (ipt *Input) RegHTTPHandler() {
	log = logger.SLogger(inputName)

	var err error
	if ipt.WPConfig != nil {
		if wkpool, err = workerpool.NewWorkerPool(ipt.WPConfig, log); err != nil {
			log.Errorf("### new worker-pool failed: %s", err.Error())
		} else if err = wkpool.Start(); err != nil {
			log.Errorf("### start worker-pool failed: %s", err.Error())
		}
	}
	if ipt.LocalCacheConfig != nil {
		if localCache, err = storage.NewStorage(ipt.LocalCacheConfig, log); err != nil {
			log.Errorf("### new local-cache failed: %s", err.Error())
		} else {
			localCache.RegisterConsumer(storage.HTTP_KEY, func(buf []byte) error {
				start := time.Now()
				reqpb := &storage.Request{}
				if err := proto.Unmarshal(buf, reqpb); err != nil {
					return err
				} else {
					req := &http.Request{
						Method:           reqpb.Method,
						Proto:            reqpb.Proto,
						ProtoMajor:       int(reqpb.ProtoMajor),
						ProtoMinor:       int(reqpb.ProtoMinor),
						Header:           storage.ConvertMapEntriesToMap(reqpb.Header),
						Body:             io.NopCloser(bytes.NewBuffer(reqpb.Body)),
						ContentLength:    reqpb.ContentLength,
						TransferEncoding: reqpb.TransferEncoding,
						Close:            reqpb.Close,
						Host:             reqpb.Host,
						Form:             storage.ConvertMapEntriesToMap(reqpb.Form),
						PostForm:         storage.ConvertMapEntriesToMap(reqpb.PostForm),
						RemoteAddr:       reqpb.RemoteAddr,
						RequestURI:       reqpb.RequestUri,
					}
					if req.URL, err = url.Parse(reqpb.Url); err != nil {
						log.Errorf("### parse raw URL: %s failed: %s", reqpb.Url, err.Error())
					}
					ipt.handleRUM(&ihttp.NopResponseWriter{}, req)

					log.Debugf("### process status: buffer-size: %dkb, cost: %dms, err: %v", len(reqpb.Body)>>10, time.Since(start)/time.Millisecond, err)

					return nil
				}
			})
			if err = localCache.RunConsumeWorker(); err != nil {
				log.Errorf("### run local-cache consumer failed: %s", err.Error())
			}
		}
	}

	for _, endpoint := range ipt.Endpoints {
		dkhttp.RegHTTPHandler(http.MethodPost, endpoint,
			workerpool.HTTPWrapper(httpStatusRespFunc, wkpool,
				storage.HTTPWrapper(storage.HTTP_KEY, httpStatusRespFunc, localCache, ipt.handleRUM)))

		log.Infof("### register RUM endpoint: %s", endpoint)
	}

	proxy, err := replayUploadHandler()
	if err != nil {
		log.Errorf("register rum replay upload proxy fail: %s", err)
	} else {
		for _, endpoint := range ipt.SessionReplayEndpoints {
			dkhttp.RegHTTPHandler(http.MethodPost, endpoint, proxy.ServeHTTP)
			log.Infof("register RUM replay upload endpoint: %s", endpoint)
		}
	}
}

func (ipt *Input) loadCDNListConf() error {
	var cdnVector []CDN
	if err := json.Unmarshal([]byte(ipt.CDNMap), &cdnVector); err != nil {
		return fmt.Errorf("json unmarshal cdn_map config fail: %w", err)
	}

	if len(cdnVector) == 0 {
		return fmt.Errorf("cdn_map resolved length is 0")
	}

	literalCDNMap := make(map[string]CDN, len(cdnVector))
	globCDNMap := make(map[*glob.Glob]CDN, 0)
	for _, cdn := range cdnVector {
		cdn.Domain = strings.TrimSpace(cdn.Domain)
		if cdn.Domain == "" {
			continue
		}
		if strings.ContainsRune(cdn.Domain, '*') {
			domain := cdn.Domain
			// Prepend prefix `*.` to domain, if the domain is `kunlun*.com`, the result will be `*.kunlun*.com`
			if domain[0] != '*' {
				if domain[0] == '.' {
					domain = "*" + domain
				} else {
					domain = "*." + domain
				}
			}
			g, err := glob.Compile(domain)
			if err == nil {
				globCDNMap[&g] = cdn
				continue
			}
		}
		literalCDNMap[strings.ToLower(cdn.Domain)] = cdn
	}
	CDNList.literal = literalCDNMap
	CDNList.glob = globCDNMap
	return nil
}

func lookupCDNNameForeach(cname string) (string, error) {
	for domain, cdn := range CDNList.literal {
		if strings.Contains(domain, strings.ToLower(cname)) {
			return cdn.Name, nil
		}
	}
	for pattern, cdn := range CDNList.glob {
		if (*pattern).Match(cname) {
			return cdn.Name, nil
		}
	}
	return "", fmt.Errorf("unable to resolve cdn name for domain: %s", cname)
}

func lookupCDNName(domain string) (string, string, error) {
	cname, err := net.LookupCNAME(domain)
	if err != nil {
		return "", "", fmt.Errorf("unable to lookup cname for %s: %w", domain, err)
	}
	cname = strings.TrimRight(cname, ".")

	segments := strings.Split(cname, ".")

	// O(1)
	if len(segments) >= 2 {
		secondLevel := segments[len(segments)-2] + "." + segments[len(segments)-1]
		if cdn, ok := CDNList.literal[secondLevel]; ok {
			return cname, cdn.Name, nil
		}
	}

	// O(n)
	cdnName, err := lookupCDNNameForeach(cname)
	return cname, cdnName, err
}

func (ipt *Input) Run() {
	log.Infof("### RUM agent serving on: %+#v", ipt.Endpoints)

	if err := loadSourcemapFile(); err != nil {
		log.Errorf("load source map file failed: %s", err.Error())
	}

	if ipt.CDNMap != "" {
		if err := ipt.loadCDNListConf(); err != nil {
			log.Errorf("load cdn map config err: %s", err)
		}
	}

	<-datakit.Exit.Wait()
	ipt.Terminate()
}

func (*Input) Terminate() {
	if wkpool != nil {
		wkpool.Shutdown()
		log.Debug("### workerpool closed")
	}
	if localCache != nil {
		if err := localCache.Close(); err != nil {
			log.Error(err.Error())
		}
		log.Debug("### local storage closed")
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}
