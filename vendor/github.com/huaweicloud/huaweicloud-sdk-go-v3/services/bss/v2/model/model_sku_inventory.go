/*
 * BSS
 *
 * Business Support System API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type SkuInventory struct {
	// |参数名称：产品ID| |参数约束及描述：产品ID|
	ProductId string `json:"product_id"`
	// |参数名称：SKU编码| |参数约束及描述：SKU编码|
	SkuCode string `json:"sku_code"`
	// |参数名称：可售库存数| |参数的约束及描述：可售库存数|
	SaleableQuantity int32 `json:"saleable_quantity"`
}

func (o SkuInventory) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SkuInventory struct{}"
	}

	return strings.Join([]string{"SkuInventory", string(data)}, " ")
}
