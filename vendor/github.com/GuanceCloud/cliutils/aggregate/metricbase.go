package aggregate

import (
	"fmt"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
)

type MetricBase struct {
	pt *point.PBPoint

	aggrTags [][2]string // hash tags
	key,
	name string

	hash uint64

	window,
	nextWallTime int64
	heapIdx int
}

// build used to delay build the tags.
func (mb *MetricBase) build() {
	for _, kv := range mb.pt.Fields {
		if kv.IsTag {
			mb.aggrTags = append(mb.aggrTags, [2]string{kv.Key, kv.GetS()})
		}
	}
}

func (mb *MetricBase) String() string {
	arr := []string{}
	arr = append(arr,
		fmt.Sprintf("aggrTags: %+#v", mb.aggrTags),
		fmt.Sprintf("key: %s", mb.key),
		fmt.Sprintf("name: %s", mb.name),
		fmt.Sprintf("hash: %d", mb.hash),
		fmt.Sprintf("window: %s", time.Duration(mb.window)),
		fmt.Sprintf("nextWallTime: %s", time.Unix(0, mb.nextWallTime)),
		fmt.Sprintf("heap index: %d", mb.heapIdx),
	)
	return strings.Join(arr, "\n")
}
