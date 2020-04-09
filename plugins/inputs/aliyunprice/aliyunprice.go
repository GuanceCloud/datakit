package aliyunprice

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/influxdata/telegraf"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/bssopenapi"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName = `aliyunprice`
)

type (
	priceReq struct {
		subscriptionReq *bssopenapi.GetSubscriptionPriceRequest
		payasyougoReq   *bssopenapi.GetPayAsYouGoPriceRequest
		payAsYouGo      bool
		region          string
		metricName      string
		interval        time.Duration

		lastTime time.Time
	}

	AliyunPriceAgent struct {
		AccessID     string `toml:"access_id"`
		RegionID     string `toml:"region_id"`
		AccessSecret string `toml:"access_key"`

		EcsCfg []*Ecs `toml:"ecs"`
		RDSCfg []*Rds `toml:"rds"`

		reqs []*priceReq

		client *bssopenapi.Client

		ctx       context.Context
		cancelFun context.CancelFunc

		wg sync.WaitGroup

		logger *models.Logger

		accumulator telegraf.Accumulator
	}
)

func (r *priceReq) String() string {
	return ``
}

func (_ *AliyunPriceAgent) SampleConfig() string {
	return sampleConfig
}

func (_ *AliyunPriceAgent) Description() string {
	return ""
}

func (_ *AliyunPriceAgent) Gather(telegraf.Accumulator) error {
	return nil
}

func (a *AliyunPriceAgent) Start(acc telegraf.Accumulator) error {

	a.logger.Info("starting...")

	a.accumulator = acc

	if cli, err := bssopenapi.NewClientWithAccessKey(a.RegionID, a.AccessID, a.AccessSecret); err != nil {
		a.logger.Errorf("fail to create client, %s", err)
		return err
	} else {
		a.client = cli
	}

	for _, item := range a.EcsCfg {
		q, err := item.toRequest()
		if err != nil {
			a.logger.Warnf("invalid config, %s", err)
		}
		a.reqs = append(a.reqs, q)
	}

	a.wg.Add(1)

	go func() {
		defer a.wg.Done()

		for {

			select {
			case <-a.ctx.Done():
				return
			default:
			}

			for _, req := range a.reqs {
				select {
				case <-a.ctx.Done():
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

				if req.payAsYouGo {

					var resp *bssopenapi.GetPayAsYouGoPriceResponse
					for i := 0; i < 5; i++ {

						resp, err = a.client.GetPayAsYouGoPrice(req.payasyougoReq)

						if tempDelay == 0 {
							tempDelay = time.Millisecond * 50
						} else {
							tempDelay *= 2
						}

						if max := time.Second; tempDelay > max {
							tempDelay = max
						}

						if err != nil {
							a.logger.Warnf("GetPayAsYouGoPrice failed: %s", err)
							time.Sleep(tempDelay)
						} else {
							if i != 0 {
								a.logger.Debugf("retry successed, %d", i)
							}
							break
						}
					}

					if err != nil {
						a.logger.Errorf("GetPayAsYouGoPrice for %s failed, %s", req.String(), err)
					} else {
						a.handlePayasyougoResp(resp, req)
					}

				} else {
					var resp *bssopenapi.GetSubscriptionPriceResponse

					for i := 0; i < 5; i++ {

						resp, err = a.client.GetSubscriptionPrice(req.subscriptionReq)

						if tempDelay == 0 {
							tempDelay = time.Millisecond * 50
						} else {
							tempDelay *= 2
						}

						if max := time.Second; tempDelay > max {
							tempDelay = max
						}

						if err != nil {
							a.logger.Warnf("%s", err)
							time.Sleep(tempDelay)
						} else {
							if i != 0 {
								a.logger.Debugf("retry successed, %d", i)
							}
							break
						}
					}

					if err != nil {
						a.logger.Errorf("GetSubscriptionPrice for %s failed, %s", req.String(), err)
					} else {
						if !resp.Success {
							a.logger.Errorf("GetSubscriptionPrice failed, %s", resp.String())
						} else {
							a.handleSubscriptionResp(resp, req)
						}
					}
				}

				req.lastTime = time.Now()

				select {
				case <-a.ctx.Done():
					return
				default:
				}
			}

			internal.SleepContext(a.ctx, time.Second*10)
		}

	}()

	return nil
}

func (a *AliyunPriceAgent) Stop() {
	a.cancelFun()
	a.wg.Wait()
}

func (a *AliyunPriceAgent) handleSubscriptionResp(resp *bssopenapi.GetSubscriptionPriceResponse, req *priceReq) {
	tags := map[string]string{}
	tags["ProductCode"] = req.subscriptionReq.ProductCode
	tags["SubscriptionType"] = "Subscription"
	tags["Currency"] = resp.Data.Currency
	tags["Region"] = req.region

	fields := map[string]interface{}{}

	for _, mod := range resp.Data.ModuleDetails.ModuleDetail {
		if mod.OriginalCost > 0 {
			//fields[mod.ModuleCode+"_OriginalCost"] = mod.OriginalCost
			fields[mod.ModuleCode+"_CostAfterDiscount"] = mod.CostAfterDiscount
			//fields[mod.ModuleCode+"_InvoiceDiscount"] = mod.InvoiceDiscount
			//fields[mod.ModuleCode+"_UnitPrice"] = mod.UnitPrice
		}
	}
	// for _, prom := range resp.Data.PromotionDetails.PromotionDetail {
	// 	if prom.PromotionName != "" {
	// 		k := fmt.Sprintf("Promotion_%v", prom.PromotionId)
	// 		promj, _ := json.Marshal(&prom)
	// 		fields[k] = promj
	// 	}
	// }
	if resp.Data.TradePrice > 0 {
		fields["TradePrice"] = resp.Data.TradePrice
	}
	if resp.Data.OriginalPrice > 0 {
		fields["OriginalPrice"] = resp.Data.OriginalPrice
	}
	if resp.Data.DiscountPrice > 0 {
		fields["DiscountPrice"] = resp.Data.DiscountPrice
	}

	if len(fields) > 0 {
		metricName := req.metricName
		if metricName == "" {
			metricName = "aliyun_price"
		}
		a.accumulator.AddFields(metricName, fields, tags, time.Now().UTC())
	}
}

func (a *AliyunPriceAgent) handlePayasyougoResp(resp *bssopenapi.GetPayAsYouGoPriceResponse, req *priceReq) {
	tags := map[string]string{}
	tags["ProductCode"] = req.payasyougoReq.ProductCode
	tags["SubscriptionType"] = "PayAsYouGo"
	tags["Currency"] = resp.Data.Currency
	tags["Region"] = req.region

	fields := map[string]interface{}{}

	for _, mod := range resp.Data.ModuleDetails.ModuleDetail {
		if mod.OriginalCost > 0 {
			fields[mod.ModuleCode+"_OriginalCost"] = mod.OriginalCost
			fields[mod.ModuleCode+"_CostAfterDiscount"] = mod.CostAfterDiscount
			fields[mod.ModuleCode+"_InvoiceDiscount"] = mod.InvoiceDiscount
			fields[mod.ModuleCode+"_UnitPrice"] = mod.UnitPrice
		}
	}
	for _, prom := range resp.Data.PromotionDetails.PromotionDetail {
		if prom.PromotionName != "" {
			k := fmt.Sprintf("Promotion_%v", prom.PromotionId)
			promj, _ := json.Marshal(&prom)
			fields[k] = promj
		}
	}
	if len(fields) > 0 {
		metricName := req.metricName
		if metricName == "" {
			metricName = "aliyun_price"
		}
		a.accumulator.AddFields(metricName, fields, tags, time.Now().UTC())
	}
}

func NewAgent() *AliyunPriceAgent {
	ac := &AliyunPriceAgent{
		logger: &models.Logger{
			Name: inputName,
		},
	}
	ac.ctx, ac.cancelFun = context.WithCancel(context.Background())

	return ac
}

func init() {
	inputs.Add(inputName, func() telegraf.Input {
		ac := NewAgent()
		return ac
	})
}
