package parser

var (
	// limit信息
	DefaultLimit       = 1000  // 查询默认的返回数量
	MaxLimit           = 10000 // 查询的最大返回数量
	DefaultGroupLimit  = 10    // 聚合默认的返回数量
	MaxGroupLimit      = 1000  // 聚合最大的返回数量
	DefaultTophitLimit = 1     // tophit默认的返回数量
	ZeroLimit          = 0     // 有聚合时，查询默认返回数量
	RegexLimit         = 1000  // 正则匹配的最大字符串长度
	BucketDepthSize    = 3     // bucket桶聚合的最大深度

	// from offset
	DefaultOffset = 0     // 默认offset
	MaxOffset     = 10000 // 最大的offset值，即为 offset + limit
	MaxBucket     = 10000 // 桶聚合最大分桶值
)

//获取size
func sizeTransport(m *DFQuery, esPtr *EST) (int, error) {
	if esPtr.dfState.aggs { // 存在聚合,没有 query limit
		return ZeroLimit, nil
	}
	return esPtr.limitSize, nil
}

// 获取offset
func fromTransport(m *DFQuery, esPtr *EST) (int, error) {
	if esPtr.dfState.aggs {
		return DefaultOffset, nil
	}
	return esPtr.fromSize, nil
}

// 获取bucket size
func (esPtr *EST) getBucketSize(i int) int {
	if i > 0 {
		return DefaultGroupLimit
	}
	return esPtr.fromSize + esPtr.limitSize
}

// 获取bucket from size
func (esPtr *EST) getAggsFromSize() int {
	if esPtr.dfState.aggs {
		return esPtr.fromSize
	}
	return DefaultOffset
}
