<!-- markdownlint-disable -->

```toml
  cdn_map='''
  [
    {
        "domain":"15cdn.com",
        "name":"腾正安全加速(原 15CDN)",
        "website":"https://www.15cdn.com"
    },
    {
        "domain":"tzcdn.cn",
        "name":"腾正安全加速(原 15CDN)",
        "website":"https://www.15cdn.com"
    },
    {
        "domain":"cedexis.net",
        "name":"Cedexis GSLB",
        "website":"https://www.cedexis.com/"
    },
    {
        "domain":"cdxcn.cn",
        "name":"Cedexis GSLB (For China)",
        "website":"https://www.cedexis.com/"
    },
    {
        "domain":"qhcdn.com",
        "name":"360 云 CDN (由奇安信运营)",
        "website":"https://cloud.360.cn/doc?name=cdn"
    },
    {
        "domain":"qh-cdn.com",
        "name":"360 云 CDN (由奇虎 360 运营)",
        "website":"https://cloud.360.cn/doc?name=cdn"
    },
    {
        "domain":"qihucdn.com",
        "name":"360 云 CDN (由奇虎 360 运营)",
        "website":"https://cloud.360.cn/doc?name=cdn"
    },
    {
        "domain":"360cdn.com",
        "name":"360 云 CDN (由奇虎 360 运营)",
        "website":"https://cloud.360.cn/doc?name=cdn"
    },
    {
        "domain":"360cloudwaf.com",
        "name":"奇安信网站卫士",
        "website":"https://wangzhan.qianxin.com"
    },
    {
        "domain":"360anyu.com",
        "name":"奇安信网站卫士",
        "website":"https://wangzhan.qianxin.com"
    },
    {
        "domain":"360safedns.com",
        "name":"奇安信网站卫士",
        "website":"https://wangzhan.qianxin.com"
    },
    {
        "domain":"360wzws.com",
        "name":"奇安信网站卫士",
        "website":"https://wangzhan.qianxin.com"
    },
    {
        "domain":"akamai.net",
        "name":"Akamai CDN",
        "website":"https://www.akamai.com"
    },
    {
        "domain":"akamaiedge.net",
        "name":"Akamai CDN",
        "website":"https://www.akamai.com"
    },
    {
        "domain":"ytcdn.net",
        "name":"Akamai CDN",
        "website":"https://www.akamai.com"
    },
    {
        "domain":"edgesuite.net",
        "name":"Akamai CDN",
        "website":"https://www.akamai.com"
    },
    {
        "domain":"akamaitech.net",
        "name":"Akamai CDN",
        "website":"https://www.akamai.com"
    },
    {
        "domain":"akamaitechnologies.com",
        "name":"Akamai CDN",
        "website":"https://www.akamai.com"
    },
    {
        "domain":"edgekey.net",
        "name":"Akamai CDN",
        "website":"https://www.akamai.com"
    },
    {
        "domain":"tl88.net",
        "name":"易通锐进(Akamai 中国)由网宿承接",
        "website":"https://www.akamai.com"
    },
    {
        "domain":"cloudfront.net",
        "name":"AWS CloudFront",
        "website":"https://aws.amazon.com/cn/cloudfront/"
    },
    {
        "domain":"worldcdn.net",
        "name":"CDN.NET",
        "website":"https://cdn.net"
    },
    {
        "domain":"worldssl.net",
        "name":"CDN.NET / CDNSUN / ONAPP",
        "website":"https://cdn.net"
    },
    {
        "domain":"cdn77.org",
        "name":"CDN77",
        "website":"https://www.cdn77.com/"
    },
    {
        "domain":"panthercdn.com",
        "name":"CDNetworks",
        "website":"https://www.cdnetworks.com"
    },
    {
        "domain":"cdnga.net",
        "name":"CDNetworks",
        "website":"https://www.cdnetworks.com"
    },
    {
        "domain":"cdngc.net",
        "name":"CDNetworks",
        "website":"https://www.cdnetworks.com"
    },
    {
        "domain":"gccdn.net",
        "name":"CDNetworks",
        "website":"https://www.cdnetworks.com"
    },
    {
        "domain":"gccdn.cn",
        "name":"CDNetworks",
        "website":"https://www.cdnetworks.com"
    },
    {
        "domain":"akamaized.net",
        "name":"Akamai CDN",
        "website":"https://www.akamai.com"
    },
    {
        "domain":"126.net",
        "name":"网易云 CDN",
        "website":"https://www.163yun.com/product/cdn"
    },
    {
        "domain":"163jiasu.com",
        "name":"网易云 CDN",
        "website":"https://www.163yun.com/product/cdn"
    },
    {
        "domain":"amazonaws.com",
        "name":"AWS Cloud",
        "website":"https://aws.amazon.com/cn/cloudfront/"
    },
    {
        "domain":"cdn77.net",
        "name":"CDN77",
        "website":"https://www.cdn77.com/"
    },
    {
        "domain":"cdnify.io",
        "name":"CDNIFY",
        "website":"https://cdnify.com"
    },
    {
        "domain":"cdnsun.net",
        "name":"CDNSUN",
        "website":"https://cdnsun.com"
    },
    {
        "domain":"bdydns.com",
        "name":"百度云 CDN",
        "website":"https://cloud.baidu.com/product/cdn.html"
    },
    {
        "domain":"ccgslb.com.cn",
        "name":"蓝汛 CDN",
        "website":"https://cn.chinacache.com/"
    },
    {
        "domain":"ccgslb.net",
        "name":"蓝汛 CDN",
        "website":"https://cn.chinacache.com/"
    },
    {
        "domain":"ccgslb.com",
        "name":"蓝汛 CDN",
        "website":"https://cn.chinacache.com/"
    },
    {
        "domain":"ccgslb.cn",
        "name":"蓝汛 CDN",
        "website":"https://cn.chinacache.com/"
    },
    {
        "domain":"c3cache.net",
        "name":"蓝汛 CDN",
        "website":"https://cn.chinacache.com/"
    },
    {
        "domain":"c3dns.net",
        "name":"蓝汛 CDN",
        "website":"https://cn.chinacache.com/"
    },
    {
        "domain":"chinacache.net",
        "name":"蓝汛 CDN",
        "website":"https://cn.chinacache.com/"
    },
    {
        "domain":"wswebcdn.com",
        "name":"网宿 CDN",
        "website":"https://www.wangsu.com/"
    },
    {
        "domain":"lxdns.com",
        "name":"网宿 CDN",
        "website":"https://www.wangsu.com/"
    },
    {
        "domain":"wswebpic.com",
        "name":"网宿 CDN",
        "website":"https://www.wangsu.com/"
    },
    {
        "domain":"cloudflare.net",
        "name":"Cloudflare",
        "website":"https://www.cloudflare.com"
    },
    {
        "domain":"akadns.net",
        "name":"Akamai CDN",
        "website":"https://www.akamai.com"
    },
    {
        "domain":"chinanetcenter.com",
        "name":"网宿 CDN",
        "website":"https://www.wangsu.com"
    },
    {
        "domain":"customcdn.com.cn",
        "name":"网宿 CDN",
        "website":"https://www.wangsu.com"
    },
    {
        "domain":"customcdn.cn",
        "name":"网宿 CDN",
        "website":"https://www.wangsu.com"
    },
    {
        "domain":"51cdn.com",
        "name":"网宿 CDN",
        "website":"https://www.wangsu.com"
    },
    {
        "domain":"wscdns.com",
        "name":"网宿 CDN",
        "website":"https://www.wangsu.com"
    },
    {
        "domain":"cdn20.com",
        "name":"网宿 CDN",
        "website":"https://www.wangsu.com"
    },
    {
        "domain":"wsdvs.com",
        "name":"网宿 CDN",
        "website":"https://www.wangsu.com"
    },
    {
        "domain":"wsglb0.com",
        "name":"网宿 CDN",
        "website":"https://www.wangsu.com"
    },
    {
        "domain":"speedcdns.com",
        "name":"网宿 CDN",
        "website":"https://www.wangsu.com"
    },
    {
        "domain":"wtxcdn.com",
        "name":"网宿 CDN",
        "website":"https://www.wangsu.com"
    },
    {
        "domain":"wsssec.com",
        "name":"网宿 WAF CDN",
        "website":"https://www.wangsu.com"
    },
    {
        "domain":"fastly.net",
        "name":"Fastly",
        "website":"https://www.fastly.com"
    },
    {
        "domain":"fastlylb.net",
        "name":"Fastly",
        "website":"https://www.fastly.com/"
    },
    {
        "domain":"hwcdn.net",
        "name":"Stackpath (原 Highwinds)",
        "website":"https://www.stackpath.com/highwinds"
    },
    {
        "domain":"incapdns.net",
        "name":"Incapsula CDN",
        "website":"https://www.incapsula.com"
    },
    {
        "domain":"kxcdn.com.",
        "name":"KeyCDN",
        "website":"https://www.keycdn.com/"
    },
    {
        "domain":"lswcdn.net",
        "name":"LeaseWeb CDN",
        "website":"https://www.leaseweb.com/cdn"
    },
    {
        "domain":"mwcloudcdn.com",
        "name":"QUANTIL (网宿)",
        "website":"https://www.quantil.com/"
    },
    {
        "domain":"mwcname.com",
        "name":"QUANTIL (网宿)",
        "website":"https://www.quantil.com/"
    },
    {
        "domain":"azureedge.net",
        "name":"Microsoft Azure CDN",
        "website":"https://azure.microsoft.com/en-us/services/cdn/"
    },
    {
        "domain":"msecnd.net",
        "name":"Microsoft Azure CDN",
        "website":"https://azure.microsoft.com/en-us/services/cdn/"
    },
    {
        "domain":"mschcdn.com",
        "name":"Microsoft Azure CDN",
        "website":"https://azure.microsoft.com/en-us/services/cdn/"
    },
    {
        "domain":"v0cdn.net",
        "name":"Microsoft Azure CDN",
        "website":"https://azure.microsoft.com/en-us/services/cdn/"
    },
    {
        "domain":"azurewebsites.net",
        "name":"Microsoft Azure App Service",
        "website":"https://azure.microsoft.com/en-us/services/app-service/"
    },
    {
        "domain":"azurewebsites.windows.net",
        "name":"Microsoft Azure App Service",
        "website":"https://azure.microsoft.com/en-us/services/app-service/"
    },
    {
        "domain":"trafficmanager.net",
        "name":"Microsoft Azure Traffic Manager",
        "website":"https://azure.microsoft.com/en-us/services/traffic-manager/"
    },
    {
        "domain":"cloudapp.net",
        "name":"Microsoft Azure",
        "website":"https://azure.microsoft.com"
    },
    {
        "domain":"chinacloudsites.cn",
        "name":"世纪互联旗下上海蓝云(承载 Azure 中国)",
        "website":"https://www.21vbluecloud.com/"
    },
    {
        "domain":"spdydns.com",
        "name":"云端智度融合 CDN",
        "website":"https://www.isurecloud.net/index.html"
    },
    {
        "domain":"jiashule.com",
        "name":"知道创宇云安全加速乐 CDN",
        "website":"https://www.yunaq.com/jsl/"
    },
    {
        "domain":"jiasule.org",
        "name":"知道创宇云安全加速乐 CDN",
        "website":"https://www.yunaq.com/jsl/"
    },
    {
        "domain":"365cyd.cn",
        "name":"知道创宇云安全创宇盾(政务专用)",
        "website":"https://www.yunaq.com/cyd/"
    },
    {
        "domain":"huaweicloud.com",
        "name":"华为云 WAF 高防云盾",
        "website":"https://www.huaweicloud.com/product/aad.html"
    },
    {
        "domain":"cdnhwc1.com",
        "name":"华为云 CDN",
        "website":"https://www.huaweicloud.com/product/cdn.html"
    },
    {
        "domain":"cdnhwc2.com",
        "name":"华为云 CDN",
        "website":"https://www.huaweicloud.com/product/cdn.html"
    },
    {
        "domain":"cdnhwc3.com",
        "name":"华为云 CDN",
        "website":"https://www.huaweicloud.com/product/cdn.html"
    },
    {
        "domain":"dnion.com",
        "name":"帝联科技",
        "website":"http://www.dnion.com/"
    },
    {
        "domain":"ewcache.com",
        "name":"帝联科技",
        "website":"http://www.dnion.com/"
    },
    {
        "domain":"globalcdn.cn",
        "name":"帝联科技",
        "website":"http://www.dnion.com/"
    },
    {
        "domain":"tlgslb.com",
        "name":"帝联科技",
        "website":"http://www.dnion.com/"
    },
    {
        "domain":"fastcdn.com",
        "name":"帝联科技",
        "website":"http://www.dnion.com/"
    },
    {
        "domain":"flxdns.com",
        "name":"帝联科技",
        "website":"http://www.dnion.com/"
    },
    {
        "domain":"dlgslb.cn",
        "name":"帝联科技",
        "website":"http://www.dnion.com/"
    },
    {
        "domain":"newdefend.cn",
        "name":"牛盾云安全",
        "website":"https://www.newdefend.com"
    },
    {
        "domain":"ffdns.net",
        "name":"CloudXNS",
        "website":"https://www.cloudxns.net"
    },
    {
        "domain":"aocdn.com",
        "name":"可靠云 CDN (贴图库)",
        "website":"http://www.kekaoyun.com/"
    },
    {
        "domain":"bsgslb.cn",
        "name":"白山云 CDN",
        "website":"https://zh.baishancloud.com/"
    },
    {
        "domain":"qingcdn.com",
        "name":"白山云 CDN",
        "website":"https://zh.baishancloud.com/"
    },
    {
        "domain":"bsclink.cn",
        "name":"白山云 CDN",
        "website":"https://zh.baishancloud.com/"
    },
    {
        "domain":"trpcdn.net",
        "name":"白山云 CDN",
        "website":"https://zh.baishancloud.com/"
    },
    {
        "domain":"anquan.io",
        "name":"牛盾云安全",
        "website":"https://www.newdefend.com"
    },
    {
        "domain":"cloudglb.com",
        "name":"快网 CDN",
        "website":"http://www.fastweb.com.cn/"
    },
    {
        "domain":"fastweb.com",
        "name":"快网 CDN",
        "website":"http://www.fastweb.com.cn/"
    },
    {
        "domain":"fastwebcdn.com",
        "name":"快网 CDN",
        "website":"http://www.fastweb.com.cn/"
    },
    {
        "domain":"cloudcdn.net",
        "name":"快网 CDN",
        "website":"http://www.fastweb.com.cn/"
    },
    {
        "domain":"fwcdn.com",
        "name":"快网 CDN",
        "website":"http://www.fastweb.com.cn/"
    },
    {
        "domain":"fwdns.net",
        "name":"快网 CDN",
        "website":"http://www.fastweb.com.cn/"
    },
    {
        "domain":"hadns.net",
        "name":"快网 CDN",
        "website":"http://www.fastweb.com.cn/"
    },
    {
        "domain":"hacdn.net",
        "name":"快网 CDN",
        "website":"http://www.fastweb.com.cn/"
    },
    {
        "domain":"cachecn.com",
        "name":"快网 CDN",
        "website":"http://www.fastweb.com.cn/"
    },
    {
        "domain":"qingcache.com",
        "name":"青云 CDN",
        "website":"https://www.qingcloud.com/products/cdn/"
    },
    {
        "domain":"qingcloud.com",
        "name":"青云 CDN",
        "website":"https://www.qingcloud.com/products/cdn/"
    },
    {
        "domain":"frontwize.com",
        "name":"青云 CDN",
        "website":"https://www.qingcloud.com/products/cdn/"
    },
    {
        "domain":"msscdn.com",
        "name":"美团云 CDN",
        "website":"https://www.mtyun.com/product/cdn"
    },
    {
        "domain":"800cdn.com",
        "name":"西部数码",
        "website":"https://www.west.cn"
    },
    {
        "domain":"tbcache.com",
        "name":"阿里云 CDN",
        "website":"https://www.aliyun.com/product/cdn"
    },
    {
        "domain":"aliyun-inc.com",
        "name":"阿里云 CDN",
        "website":"https://www.aliyun.com/product/cdn"
    },
    {
        "domain":"aliyuncs.com",
        "name":"阿里云 CDN",
        "website":"https://www.aliyun.com/product/cdn"
    },
    {
        "domain":"alikunlun.net",
        "name":"阿里云 CDN",
        "website":"https://www.aliyun.com/product/cdn"
    },
    {
        "domain":"alikunlun.com",
        "name":"阿里云 CDN",
        "website":"https://www.aliyun.com/product/cdn"
    },
    {
        "domain":"alicdn.com",
        "name":"阿里云 CDN",
        "website":"https://www.aliyun.com/product/cdn"
    },
    {
        "domain":"aligaofang.com",
        "name":"阿里云盾高防",
        "website":"https://www.aliyun.com/product/ddos"
    },
    {
        "domain":"yundunddos.com",
        "name":"阿里云盾高防",
        "website":"https://www.aliyun.com/product/ddos"
    },
    {
        "domain":"kunlun*.com",
        "name":"阿里云 CDN",
        "website":"https://www.aliyun.com/product/cdn"
    },
    {
        "domain":"cdngslb.com",
        "name":"阿里云 CDN",
        "website":"https://www.aliyun.com/product/cdn"
    },
    {
        "domain":"yunjiasu-cdn.net",
        "name":"百度云加速",
        "website":"https://su.baidu.com"
    },
    {
        "domain":"momentcdn.com",
        "name":"魔门云 CDN",
        "website":"https://www.cachemoment.com"
    },
    {
        "domain":"aicdn.com",
        "name":"又拍云",
        "website":"https://www.upyun.com"
    },
    {
        "domain":"qbox.me",
        "name":"七牛云",
        "website":"https://www.qiniu.com"
    },
    {
        "domain":"qiniu.com",
        "name":"七牛云",
        "website":"https://www.qiniu.com"
    },
    {
        "domain":"qiniudns.com",
        "name":"七牛云",
        "website":"https://www.qiniu.com"
    },
    {
        "domain":"jcloudcs.com",
        "name":"京东云 CDN",
        "website":"https://www.jdcloud.com/cn/products/cdn"
    },
    {
        "domain":"jdcdn.com",
        "name":"京东云 CDN",
        "website":"https://www.jdcloud.com/cn/products/cdn"
    },
    {
        "domain":"qianxun.com",
        "name":"京东云 CDN",
        "website":"https://www.jdcloud.com/cn/products/cdn"
    },
    {
        "domain":"jcloudlb.com",
        "name":"京东云 CDN",
        "website":"https://www.jdcloud.com/cn/products/cdn"
    },
    {
        "domain":"jcloud-cdn.com",
        "name":"京东云 CDN",
        "website":"https://www.jdcloud.com/cn/products/cdn"
    },
    {
        "domain":"maoyun.tv",
        "name":"猫云融合 CDN",
        "website":"https://www.maoyun.com/"
    },
    {
        "domain":"maoyundns.com",
        "name":"猫云融合 CDN",
        "website":"https://www.maoyun.com/"
    },
    {
        "domain":"xgslb.net",
        "name":"WebLuker (蓝汛)",
        "website":"http://www.webluker.com"
    },
    {
        "domain":"ucloud.cn",
        "name":"UCloud CDN",
        "website":"https://www.ucloud.cn/site/product/ucdn.html"
    },
    {
        "domain":"ucloud.com.cn",
        "name":"UCloud CDN",
        "website":"https://www.ucloud.cn/site/product/ucdn.html"
    },
    {
        "domain":"cdndo.com",
        "name":"UCloud CDN",
        "website":"https://www.ucloud.cn/site/product/ucdn.html"
    },
    {
        "domain":"zenlogic.net",
        "name":"Zenlayer CDN",
        "website":"https://www.zenlayer.com"
    },
    {
        "domain":"ogslb.com",
        "name":"Zenlayer CDN",
        "website":"https://www.zenlayer.com"
    },
    {
        "domain":"uxengine.net",
        "name":"Zenlayer CDN",
        "website":"https://www.zenlayer.com"
    },
    {
        "domain":"tan14.net",
        "name":"TAN14 CDN",
        "website":"http://www.tan14.cn/"
    },
    {
        "domain":"verycloud.cn",
        "name":"VeryCloud 云分发",
        "website":"https://www.verycloud.cn/"
    },
    {
        "domain":"verycdn.net",
        "name":"VeryCloud 云分发",
        "website":"https://www.verycloud.cn/"
    },
    {
        "domain":"verygslb.com",
        "name":"VeryCloud 云分发",
        "website":"https://www.verycloud.cn/"
    },
    {
        "domain":"xundayun.cn",
        "name":"SpeedyCloud CDN",
        "website":"https://www.speedycloud.cn/zh/Products/CDN/CloudDistribution.html"
    },
    {
        "domain":"xundayun.com",
        "name":"SpeedyCloud CDN",
        "website":"https://www.speedycloud.cn/zh/Products/CDN/CloudDistribution.html"
    },
    {
        "domain":"speedycloud.cc",
        "name":"SpeedyCloud CDN",
        "website":"https://www.speedycloud.cn/zh/Products/CDN/CloudDistribution.html"
    },
    {
        "domain":"mucdn.net",
        "name":"Verizon CDN (Edgecast)",
        "website":"https://www.verizondigitalmedia.com/platform/edgecast-cdn/"
    },
    {
        "domain":"nucdn.net",
        "name":"Verizon CDN (Edgecast)",
        "website":"https://www.verizondigitalmedia.com/platform/edgecast-cdn/"
    },
    {
        "domain":"alphacdn.net",
        "name":"Verizon CDN (Edgecast)",
        "website":"https://www.verizondigitalmedia.com/platform/edgecast-cdn/"
    },
    {
        "domain":"systemcdn.net",
        "name":"Verizon CDN (Edgecast)",
        "website":"https://www.verizondigitalmedia.com/platform/edgecast-cdn/"
    },
    {
        "domain":"edgecastcdn.net",
        "name":"Verizon CDN (Edgecast)",
        "website":"https://www.verizondigitalmedia.com/platform/edgecast-cdn/"
    },
    {
        "domain":"zetacdn.net",
        "name":"Verizon CDN (Edgecast)",
        "website":"https://www.verizondigitalmedia.com/platform/edgecast-cdn/"
    },
    {
        "domain":"coding.io",
        "name":"Coding Pages",
        "website":"https://coding.net/pages"
    },
    {
        "domain":"coding.me",
        "name":"Coding Pages",
        "website":"https://coding.net/pages"
    },
    {
        "domain":"gitlab.io",
        "name":"GitLab Pages",
        "website":"https://docs.gitlab.com/ee/user/project/pages/"
    },
    {
        "domain":"github.io",
        "name":"GitHub Pages",
        "website":"https://pages.github.com/"
    },
    {
        "domain":"herokuapp.com",
        "name":"Heroku SaaS",
        "website":"https://www.heroku.com"
    },
    {
        "domain":"googleapis.com",
        "name":"Google Cloud Storage",
        "website":"https://cloud.google.com/storage/"
    },
    {
        "domain":"netdna.com",
        "name":"Stackpath (原 MaxCDN)",
        "website":"https://www.stackpath.com/maxcdn/"
    },
    {
        "domain":"netdna-cdn.com",
        "name":"Stackpath (原 MaxCDN)",
        "website":"https://www.stackpath.com/maxcdn/"
    },
    {
        "domain":"netdna-ssl.com",
        "name":"Stackpath (原 MaxCDN)",
        "website":"https://www.stackpath.com/maxcdn/"
    },
    {
        "domain":"cdntip.com",
        "name":"腾讯云 CDN",
        "website":"https://cloud.tencent.com/product/cdn-scd"
    },
    {
        "domain":"dnsv1.com",
        "name":"腾讯云 CDN",
        "website":"https://cloud.tencent.com/product/cdn-scd"
    },
    {
        "domain":"tencdns.net",
        "name":"腾讯云 CDN",
        "website":"https://cloud.tencent.com/product/cdn-scd"
    },
    {
        "domain":"dayugslb.com",
        "name":"腾讯云大禹 BGP 高防",
        "website":"https://cloud.tencent.com/product/ddos-advanced"
    },
    {
        "domain":"tcdnvod.com",
        "name":"腾讯云视频 CDN",
        "website":"https://lab.skk.moe/cdn"
    },
    {
        "domain":"tdnsv5.com",
        "name":"腾讯云 CDN",
        "website":"https://cloud.tencent.com/product/cdn-scd"
    },
    {
        "domain":"ksyuncdn.com",
        "name":"金山云 CDN",
        "website":"https://www.ksyun.com/post/product/CDN"
    },
    {
        "domain":"ks-cdn.com",
        "name":"金山云 CDN",
        "website":"https://www.ksyun.com/post/product/CDN"
    },
    {
        "domain":"ksyuncdn-k1.com",
        "name":"金山云 CDN",
        "website":"https://www.ksyun.com/post/product/CDN"
    },
    {
        "domain":"netlify.com",
        "name":"Netlify",
        "website":"https://www.netlify.com"
    },
    {
        "domain":"zeit.co",
        "name":"ZEIT Now Smart CDN",
        "website":"https://zeit.co"
    },
    {
        "domain":"zeit-cdn.net",
        "name":"ZEIT Now Smart CDN",
        "website":"https://zeit.co"
    },
    {
        "domain":"b-cdn.net",
        "name":"Bunny CDN",
        "website":"https://bunnycdn.com/"
    },
    {
        "domain":"lsycdn.com",
        "name":"蓝视云 CDN",
        "website":"https://cloud.lsy.cn/"
    },
    {
        "domain":"scsdns.com",
        "name":"逸云科技云加速 CDN",
        "website":"http://www.exclouds.com/navPage/wise"
    },
    {
        "domain":"quic.cloud",
        "name":"QUIC.Cloud",
        "website":"https://quic.cloud/"
    },
    {
        "domain":"flexbalancer.net",
        "name":"FlexBalancer - Smart Traffic Routing",
        "website":"https://perfops.net/flexbalancer"
    },
    {
        "domain":"gcdn.co",
        "name":"G - Core Labs",
        "website":"https://gcorelabs.com/cdn/"
    },
    {
        "domain":"sangfordns.com",
        "name":"深信服 AD 系列应用交付产品  单边加速解决方案",
        "website":"http://www.sangfor.com.cn/topic/2011adn/solutions5.html"
    },
    {
        "domain":"stspg-customer.com",
        "name":"StatusPage.io",
        "website":"https://www.statuspage.io"
    },
    {
        "domain":"turbobytes.net",
        "name":"TurboBytes Multi-CDN",
        "website":"https://www.turbobytes.com"
    },
    {
        "domain":"turbobytes-cdn.com",
        "name":"TurboBytes Multi-CDN",
        "website":"https://www.turbobytes.com"
    },
    {
        "domain":"att-dsa.net",
        "name":"AT&T Content Delivery Network",
        "website":"https://www.business.att.com/products/cdn.html"
    },
    {
        "domain":"azioncdn.net",
        "name":"Azion Tech | Edge Computing PLatform",
        "website":"https://www.azion.com"
    },
    {
        "domain":"belugacdn.com",
        "name":"BelugaCDN",
        "website":"https://www.belugacdn.com"
    },
    {
        "domain":"cachefly.net",
        "name":"CacheFly CDN",
        "website":"https://www.cachefly.com/"
    },
    {
        "domain":"inscname.net",
        "name":"Instart CDN",
        "website":"https://www.instart.com/products/web-performance/cdn"
    },
    {
        "domain":"insnw.net",
        "name":"Instart CDN",
        "website":"https://www.instart.com/products/web-performance/cdn"
    },
    {
        "domain":"internapcdn.net",
        "name":"Internap CDN",
        "website":"https://www.inap.com/network/content-delivery-network"
    },
    {
        "domain":"footprint.net",
        "name":"CenturyLink CDN (原 Level 3)",
        "website":"https://www.centurylink.com/business/networking/cdn.html"
    },
    {
        "domain":"llnwi.net",
        "name":"Limelight Network",
        "website":"https://www.limelight.com"
    },
    {
        "domain":"llnwd.net",
        "name":"Limelight Network",
        "website":"https://www.limelight.com"
    },
    {
        "domain":"unud.net",
        "name":"Limelight Network",
        "website":"https://www.limelight.com"
    },
    {
        "domain":"lldns.net",
        "name":"Limelight Network",
        "website":"https://www.limelight.com"
    },
    {
        "domain":"stackpathdns.com",
        "name":"Stackpath CDN",
        "website":"https://www.stackpath.com"
    },
    {
        "domain":"stackpathcdn.com",
        "name":"Stackpath CDN",
        "website":"https://www.stackpath.com"
    },
    {
        "domain":"mncdn.com",
        "name":"Medianova",
        "website":"https://www.medianova.com"
    },
    {
        "domain":"rncdn1.com",
        "name":"Relected Networks",
        "website":"https://reflected.net/globalcdn"
    },
    {
        "domain":"simplecdn.net",
        "name":"Relected Networks",
        "website":"https://reflected.net/globalcdn"
    },
    {
        "domain":"swiftserve.com",
        "name":"Conversant - SwiftServe CDN",
        "website":"https://reflected.net/globalcdn"
    },
    {
        "domain":"bitgravity.com",
        "name":"Tata communications CDN",
        "website":"https://cdn.tatacommunications.com"
    },
    {
        "domain":"zenedge.net",
        "name":"Oracle Dyn Web Application Security suite (原 Zenedge CDN)",
        "website":"https://cdn.tatacommunications.com"
    },
    {
        "domain":"biliapi.com",
        "name":"Bilibili 业务 GSLB",
        "website":"https://lab.skk.moe/cdn"
    },
    {
        "domain":"hdslb.net",
        "name":"Bilibili 高可用负载均衡",
        "website":"https://github.com/bilibili/overlord"
    },
    {
        "domain":"hdslb.com",
        "name":"Bilibili 高可用地域负载均衡",
        "website":"https://github.com/bilibili/overlord"
    },
    {
        "domain":"xwaf.cn",
        "name":"极御云安全(浙江壹云云计算有限公司)",
        "website":"https://www.stopddos.cn"
    },
    {
        "domain":"shifen.com",
        "name":"百度旗下业务地域负载均衡系统",
        "website":"https://lab.skk.moe/cdn"
    },
    {
        "domain":"sinajs.cn",
        "name":"新浪静态域名",
        "website":"https://lab.skk.moe/cdn"
    },
    {
        "domain":"tencent-cloud.net",
        "name":"腾讯旗下业务地域负载均衡系统",
        "website":"https://lab.skk.moe/cdn"
    },
    {
        "domain":"elemecdn.com",
        "name":"饿了么静态域名与地域负载均衡",
        "website":"https://lab.skk.moe/cdn"
    },
    {
        "domain":"sinaedge.com",
        "name":"新浪科技融合 CDN 负载均衡",
        "website":"https://lab.skk.moe/cdn"
    },
    {
        "domain":"sina.com.cn",
        "name":"新浪科技融合 CDN 负载均衡",
        "website":"https://lab.skk.moe/cdn"
    },
    {
        "domain":"sinacdn.com",
        "name":"新浪云 CDN",
        "website":"https://www.sinacloud.com/doc/sae/php/cdn.html"
    },
    {
        "domain":"sinasws.com",
        "name":"新浪云 CDN",
        "website":"https://www.sinacloud.com/doc/sae/php/cdn.html"
    },
    {
        "domain":"saebbs.com",
        "name":"新浪云 SAE 云引擎",
        "website":"https://www.sinacloud.com/doc/sae/php/cdn.html"
    },
    {
        "domain":"websitecname.cn",
        "name":"美橙互联旗下建站之星",
        "website":"https://www.sitestar.cn"
    },
    {
        "domain":"cdncenter.cn",
        "name":"美橙互联 CDN",
        "website":"https://www.cndns.com"
    },
    {
        "domain":"vhostgo.com",
        "name":"西部数码虚拟主机",
        "website":"https://www.west.cn"
    },
    {
        "domain":"jsd.cc",
        "name":"上海云盾 YUNDUN",
        "website":"https://www.yundun.com"
    },
    {
        "domain":"powercdn.cn",
        "name":"动力在线 CDN",
        "website":"http://www.powercdn.com"
    },
    {
        "domain":"21vokglb.cn",
        "name":"世纪互联云快线业务",
        "website":"https://www.21vianet.com"
    },
    {
        "domain":"21vianet.com.cn",
        "name":"世纪互联云快线业务",
        "website":"https://www.21vianet.com"
    },
    {
        "domain":"21okglb.cn",
        "name":"世纪互联云快线业务",
        "website":"https://www.21vianet.com"
    },
    {
        "domain":"21speedcdn.com",
        "name":"世纪互联云快线业务",
        "website":"https://www.21vianet.com"
    },
    {
        "domain":"21cvcdn.com",
        "name":"世纪互联云快线业务",
        "website":"https://www.21vianet.com"
    },
    {
        "domain":"okcdn.com",
        "name":"世纪互联云快线业务",
        "website":"https://www.21vianet.com"
    },
    {
        "domain":"okglb.com",
        "name":"世纪互联云快线业务",
        "website":"https://www.21vianet.com"
    },
    {
        "domain":"cdnetworks.net",
        "name":"北京同兴万点网络技术",
        "website":"http://www.txnetworks.cn/"
    },
    {
        "domain":"txnetworks.cn",
        "name":"北京同兴万点网络技术",
        "website":"http://www.txnetworks.cn/"
    },
    {
        "domain":"cdnnetworks.com",
        "name":"北京同兴万点网络技术",
        "website":"http://www.txnetworks.cn/"
    },
    {
        "domain":"txcdn.cn",
        "name":"北京同兴万点网络技术",
        "website":"http://www.txnetworks.cn/"
    },
    {
        "domain":"cdnunion.net",
        "name":"宝腾互联旗下上海万根网络(CDN 联盟)",
        "website":"http://www.cdnunion.com"
    },
    {
        "domain":"cdnunion.com",
        "name":"宝腾互联旗下上海万根网络(CDN 联盟)",
        "website":"http://www.cdnunion.com"
    },
    {
        "domain":"mygslb.com",
        "name":"宝腾互联旗下上海万根网络(YaoCDN)",
        "website":"http://www.vangen.cn"
    },
    {
        "domain":"cdnudns.com",
        "name":"宝腾互联旗下上海万根网络(YaoCDN)",
        "website":"http://www.vangen.cn"
    },
    {
        "domain":"sprycdn.com",
        "name":"宝腾互联旗下上海万根网络(YaoCDN)",
        "website":"http://www.vangen.cn"
    },
    {
        "domain":"chuangcdn.com",
        "name":"创世云融合 CDN",
        "website":"https://www.chuangcache.com/index.html"
    },
    {
        "domain":"aocde.com",
        "name":"创世云融合 CDN",
        "website":"https://www.chuangcache.com"
    },
    {
        "domain":"ctxcdn.cn",
        "name":"中国电信天翼云 CDN",
        "website":"https://www.ctyun.cn/product2/#/product/10027560"
    },
    {
        "domain":"yfcdn.net",
        "name":"云帆加速 CDN",
        "website":"https://www.yfcloud.com"
    },
    {
        "domain":"mmycdn.cn",
        "name":"蛮蛮云 CDN(中联利信)",
        "website":"https://www.chinamaincloud.com/cloudDispatch.html"
    },
    {
        "domain":"chinamaincloud.com",
        "name":"蛮蛮云 CDN(中联利信)",
        "website":"https://www.chinamaincloud.com/cloudDispatch.html"
    },
    {
        "domain":"cnispgroup.com",
        "name":"中联数据(中联利信)",
        "website":"http://www.cnispgroup.com/"
    },
    {
        "domain":"cdnle.com",
        "name":"新乐视云联(原乐视云)CDN",
        "website":"http://www.lecloud.com/zh-cn"
    },
    {
        "domain":"gosuncdn.com",
        "name":"高升控股 CDN 技术",
        "website":"http://www.gosun.com"
    },
    {
        "domain":"mmtrixopt.com",
        "name":"mmTrix 性能魔方(高升控股旗下)",
        "website":"http://www.mmtrix.com"
    },
    {
        "domain":"cloudfence.cn",
        "name":"蓝盾云 CDN",
        "website":"https://www.cloudfence.cn/#/cloudWeb/yaq/yaqyfx"
    },
    {
        "domain":"ngaagslb.cn",
        "name":"新流云(新流万联)",
        "website":"https://www.ngaa.com.cn"
    },
    {
        "domain":"p2cdn.com",
        "name":"星域云 P2P CDN",
        "website":"https://www.xycloud.com"
    },
    {
        "domain":"00cdn.com",
        "name":"星域云 P2P CDN",
        "website":"https://www.xycloud.com"
    },
    {
        "domain":"sankuai.com",
        "name":"美团云(三快科技)负载均衡",
        "website":"https://www.mtyun.com"
    },
    {
        "domain":"lccdn.org",
        "name":"领智云 CDN(杭州领智云画)",
        "website":"http://www.linkingcloud.com"
    },
    {
        "domain":"nscloudwaf.com",
        "name":"绿盟云 WAF",
        "website":"https://cloud.nsfocus.com"
    },
    {
        "domain":"2cname.com",
        "name":"网堤安全",
        "website":"https://www.ddos.com"
    },
    {
        "domain":"ucloudgda.com",
        "name":"UCloud 罗马 Rome 全球网络加速",
        "website":"https://www.ucloud.cn/site/product/rome.html"
    },
    {
        "domain":"google.com",
        "name":"Google Web 业务",
        "website":"https://lab.skk.moe/cdn"
    },
    {
        "domain":"1e100.net",
        "name":"Google Web 业务",
        "website":"https://lab.skk.moe/cdn"
    },
    {
        "domain":"ncname.com",
        "name":"NodeCache",
        "website":"https://www.nodecache.com"
    },
    {
        "domain":"alipaydns.com",
        "name":"蚂蚁金服旗下业务地域负载均衡系统",
        "website":"https://lab.skk.moe/cdn/"
    },
    {
        "domain":"wscloudcdn.com",
        "name":"全速云(网宿)CloudEdge 云加速",
        "website":"https://www.quansucloud.com/product.action?product.id=270"
    }
]
'''
```
