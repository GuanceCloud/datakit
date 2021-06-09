package parser

// 高亮字段的查询
// "highlight": {
//   "fields": {
//     "f1": {
//       "fragment_size": 1000
//     },
//     "f2": {
//       "fragment_size": 1000
//     }
//   }
// }
func highlightTransport(m *DFQuery, esPtr *EST) (interface{}, error) {

	if (!esPtr.dfState.aggs) && (esPtr.IsHighlight) {
		if len(esPtr.HighlightFields) > 0 {
			inner := map[string]interface{}{}
			for _, v := range esPtr.HighlightFields {
				iinner := map[string]int{
					"fragment_size": HighlightFragmentSize,
				}
				inner[v] = iinner
			}
			outer := map[string]interface{}{
				"fields": inner,
			}
			return outer, nil
		}
	}
	return nil, nil
}
