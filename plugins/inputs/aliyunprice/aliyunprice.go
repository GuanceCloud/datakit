package aliyunprice

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/bssopenapi"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName = `aliyunprice`

	moduleLogger *logger.Logger
)

type (
	priceReq struct {
		subscriptionReq *bssopenapi.GetSubscriptionPriceRequest
		payasyougoReq   *bssopenapi.GetPayAsYouGoPriceRequest
		payAsYouGo      bool

		//有时拿module信息和拿price的不一样，比如eip
		productCodeForPriceModulesSubscript string
		productTypeForPriceModulesSubscript string

		productCodeForPriceModulesPayasugo string
		productTypeForPriceModulesPayasugo string

		region     string
		metricName string
		interval   time.Duration

		lastTime time.Time

		m priceMod

		//产品的计价模块信息，分为Subscription(预付费)和PayAsYouGo(后付费)
		priceModuleInfos map[string]*bssopenapi.ModuleList
		//如果产品的计价模块信息获取失败了，记录下来，隔一段时间再做尝试
		fetchModulePriceHistory map[string]time.Time
	}

	AliyunPriceAgent struct {
		AccessID     string `toml:"access_key_id"`
		AccessSecret string `toml:"access_key_secret"`
		RegionID     string `toml:"region_id"`

		EcsCfg []*Ecs `toml:"ecs"`
		RDSCfg []*Rds `toml:"rds"`
		EipCfg []*Eip `toml:"eip"`
		SlbCfg []*Slb `toml:"slb"`

		reqs []*priceReq

		client *bssopenapi.Client

		ctx       context.Context
		cancelFun context.CancelFunc

		wg sync.WaitGroup

		mode string

		testError error
	}
)

func (a *AliyunPriceAgent) isTest() bool {
	return a.mode == "test"
}

func (a *AliyunPriceAgent) isDebug() bool {
	return a.mode == "debug"
}

func (r *priceReq) String() string {
	return ``
}

func (_ *AliyunPriceAgent) Catalog() string {
	return "aliyun"
}

func (_ *AliyunPriceAgent) SampleConfig() string {
	return globalConfig + ecsSampleConfig + rdsSampleConfig + eipSampleConfig + slbSampleConfig
}

func (a *AliyunPriceAgent) Run() {

	moduleLogger = logger.SLogger(inputName)

	go func() {
		<-datakit.Exit.Wait()
		a.cancelFun()
	}()

	for {
		select {
		case <-a.ctx.Done():
			return
		default:
		}

		if cli, err := bssopenapi.NewClientWithAccessKey(a.RegionID, a.AccessID, a.AccessSecret); err != nil {
			moduleLogger.Errorf("fail to create client, %s", err)
			time.Sleep(time.Second)
		} else {
			a.client = cli
			break
		}
	}

	for _, item := range a.EcsCfg {
		q, err := item.toRequest()
		if err != nil {
			moduleLogger.Warnf("invalid ecs config, %s", err)
		}
		a.reqs = append(a.reqs, q)
	}

	for _, item := range a.RDSCfg {
		q, err := item.toRequest()
		if err != nil {
			moduleLogger.Warnf("invalid rds config, %s", err)
		}
		a.reqs = append(a.reqs, q)
	}

	for _, item := range a.EipCfg {
		q, err := item.toRequest()
		if err != nil {
			moduleLogger.Warnf("invalid eip config, %s", err)
		}
		a.reqs = append(a.reqs, q)
	}

	for _, item := range a.SlbCfg {
		q, err := item.toRequest()
		if err != nil {
			moduleLogger.Warnf("invalid slb config, %s", err)
		}
		a.reqs = append(a.reqs, q)
	}

	a.wg.Add(1)

	go func() {
		defer a.wg.Done()

		defer func() {
			if e := recover(); e != nil {
				moduleLogger.Errorf("panic %v", e)
			}
		}()

		for {

			select {
			case <-datakit.Exit.Wait():
				return
			default:
			}

			for _, req := range a.reqs {
				select {
				case <-datakit.Exit.Wait():
					return
				default:
				}

				if req.lastTime.IsZero() {
					req.lastTime = time.Now()
				} else if time.Now().Sub(req.lastTime) < req.interval {
					continue
				}

				var err error
				var tempDelay time.Duration

				var respPayasyougo *bssopenapi.GetPayAsYouGoPriceResponse
				var respSubscript *bssopenapi.GetSubscriptionPriceResponse

				for i := 0; i < 5; i++ {

					if req.payAsYouGo {
						respPayasyougo, err = a.client.GetPayAsYouGoPrice(req.payasyougoReq)
						if err == nil && respPayasyougo != nil && !respPayasyougo.Success {
							err = fmt.Errorf("%s", respPayasyougo.String())
						}
					} else {
						respSubscript, err = a.client.GetSubscriptionPrice(req.subscriptionReq)
						if err == nil && respSubscript != nil && !respSubscript.Success {
							err = fmt.Errorf("%s", respSubscript.String())
						}
					}

					if tempDelay == 0 {
						tempDelay = time.Millisecond * 50
					} else {
						tempDelay *= 2
					}

					if max := time.Second; tempDelay > max {
						tempDelay = max
					}

					if err != nil {
						moduleLogger.Warnf("get price failed")
						if a.isTest() {
							a.testError = err
							return
						}
						time.Sleep(tempDelay)
					} else {
						if i != 0 {
							moduleLogger.Debugf("retry successed, %d", i)
						}
						break
					}
				}

				if err != nil {
					moduleLogger.Errorf("get price failed, %s", err)
				} else {
					if req.payAsYouGo {
						a.handleResponse(&respPayasyougo.Data, req)
					} else {
						a.handleResponse(&respSubscript.Data, req)
					}
				}

				req.lastTime = time.Now()

				select {
				case <-datakit.Exit.Wait():
					return
				default:
				}
			}

			if a.isTest() {
				break
			}

			datakit.SleepContext(a.ctx, time.Second*5)
		}

	}()

}

func (a *AliyunPriceAgent) handleResponse(respData *bssopenapi.Data, req *priceReq) {
	tags := map[string]string{}

	productCode := ""
	productType := ""
	subscriptionType := ""
	if req.payAsYouGo {
		subscriptionType = req.payasyougoReq.SubscriptionType
		productCode = req.payasyougoReq.ProductCode
		productType = req.payasyougoReq.ProductType
	} else {
		subscriptionType = req.subscriptionReq.SubscriptionType
		productCode = req.subscriptionReq.ProductCode
		productType = req.subscriptionReq.ProductType
	}
	tags["ProductCode"] = productCode
	tags["ProductType"] = productType
	tags["SubscriptionType"] = subscriptionType
	tags["Currency"] = respData.Currency
	tags["Region"] = req.region

	fields := map[string]interface{}{}

	for _, mod := range respData.ModuleDetails.ModuleDetail {
		if mod.OriginalCost > 0 {
			//每个计费模块的信息
			fields["Module_"+mod.ModuleCode+"_OriginalCost"] = mod.OriginalCost           //原价
			fields["Module_"+mod.ModuleCode+"_CostAfterDiscount"] = mod.CostAfterDiscount //折后价
			fields["Module_"+mod.ModuleCode+"_InvoiceDiscount"] = mod.InvoiceDiscount     //打折减去的价钱
			if mod.UnitPrice > 0 {
				fields["Module_"+mod.ModuleCode+"_UnitPrice"] = mod.UnitPrice //单价
			}
		}

		//获取计费模块的信息
		modinfo := req.getModInfo(mod.ModuleCode, req.payAsYouGo)

		if modinfo == nil {

			skip := true
			if ht, ok := req.fetchModulePriceHistory[subscriptionType]; ok {
				if time.Now().Sub(ht) > time.Minute*30 {
					skip = false
				}
			} else {
				skip = false
			}

			if !skip {
				pcode := ""
				ptype := ""
				if req.payAsYouGo {
					pcode = req.productCodeForPriceModulesPayasugo
					ptype = req.productTypeForPriceModulesPayasugo
				} else {
					pcode = req.productCodeForPriceModulesSubscript
					ptype = req.productTypeForPriceModulesSubscript
				}
				modlist, err := a.fetchProductPriceModule(req.region, pcode, ptype, subscriptionType)
				if err != nil {
					moduleLogger.Errorf("fail to get price modules, ProductCode=%s, ProductType=%s, SubscriptionType=%s, %s", pcode, ptype, subscriptionType, err)
					req.fetchModulePriceHistory[subscriptionType] = time.Now()
				} else {
					req.priceModuleInfos[subscriptionType] = modlist
					modinfo = req.getModInfo(mod.ModuleCode, true)
				}
			}
		}

		//eg., 按月付
		if modinfo != nil && modinfo.PriceType != "" {
			fields["Module_"+mod.ModuleCode+"_PriceType"] = modinfo.PriceType
		}
	}

	//优惠活动
	if len(respData.PromotionDetails.PromotionDetail) > 0 {
		prominfo, err := json.Marshal(respData.PromotionDetails)
		if err != nil {
			moduleLogger.Warnf("fail to marshal PromotionDetails, %s", err)
		} else {
			fields["Promotion"] = string(prominfo)
		}
	}

	if respData.TradePrice > 0 {
		fields["TradePrice"] = respData.TradePrice
	}
	if respData.OriginalPrice > 0 {
		fields["OriginalPrice"] = respData.OriginalPrice
	}
	if respData.DiscountPrice > 0 {
		fields["DiscountPrice"] = respData.DiscountPrice
	}

	if req.m != nil {
		tags = req.m.handleTags(tags)
		fields = req.m.handleFields(fields)
	}

	if len(fields) > 0 {
		metricName := req.metricName
		if metricName == "" {
			metricName = "aliyun_price"
		}

		if a.isTest() {
			// pass
		} else {
			io.NamedFeedEx(inputName, datakit.Metric, metricName, tags, fields)
		}
	}
}

func (req *priceReq) getModInfo(code string, payasyoug bool) *bssopenapi.Module {

	subscripType := `Subscription`
	if payasyoug {
		subscripType = `PayAsYouGo`
	}
	if l, ok := req.priceModuleInfos[subscripType]; ok {
		for _, m := range l.Module {
			if m.ModuleCode == code {
				return &m
			}
		}
	}
	return nil
}

func (a *AliyunPriceAgent) fetchProductPriceModule(region, productCode, productType, subscriptionType string) (*bssopenapi.ModuleList, error) {

	req := bssopenapi.CreateDescribePricingModuleRequest()
	req.Scheme = `https`
	req.RegionId = region
	req.ProductCode = productCode
	req.ProductType = productType
	req.SubscriptionType = subscriptionType

	resp, err := a.client.DescribePricingModule(req)
	if err == nil && resp != nil && !resp.Success {
		err = fmt.Errorf("%s", resp.String())
	}

	if err != nil {
		return nil, err
	}

	// for _, attr := range resp.Data.AttributeList.Attribute {
	// 	//attr.Values.AttributeValue //属性的值的取值范围
	// 	if attr.Code == "InstanceType" {
	// 		for _, v := range attr.Values.AttributeValue {
	// 			fmt.Printf("%s : %s - %s\n", v.Name, v.Value, v.Remark)
	// 		}
	// 	}
	// }

	// for _, mod := range resp.Data.ModuleList.Module {
	// 	_ = mod
	// }

	return &resp.Data.ModuleList, nil
}

func NewAgent() *AliyunPriceAgent {
	ac := &AliyunPriceAgent{}
	ac.ctx, ac.cancelFun = context.WithCancel(context.Background())

	return ac
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		ac := NewAgent()
		return ac
	})
}
