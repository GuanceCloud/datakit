package pprof

import (
	"sort"

	"github.com/GuanceCloud/cliutils/pprofparser/domain/events"
	"github.com/GuanceCloud/cliutils/pprofparser/domain/languages"
	"github.com/GuanceCloud/cliutils/pprofparser/domain/quantity"
)

type SummaryValueType struct {
	Type events.Type    `json:"type"`
	Unit *quantity.Unit `json:"unit"`
}

type EventSummary struct {
	*SummaryValueType
	Value int64 `json:"value"`
}

func (es *EventSummary) ConvertToDefaultUnit() {
	if es.Unit == nil {
		return
	}
	es.Value = es.Unit.ConvertToDefaultUnit(es.Value)
	es.Unit = es.Unit.Kind.DefaultUnit
}

type SummaryCollection []*EventSummary

func (s SummaryCollection) SortByType(lang languages.Lang) {
	cs := &CollectionSort{
		collection: s,
		lessFunc: func(a, b *EventSummary) bool {
			return a.Type.GetSort(lang) < b.Type.GetSort(lang)
		},
	}
	sort.Sort(cs)
}

func (s SummaryCollection) SortByValue() {
	cs := &CollectionSort{
		collection: s,
		lessFunc: func(a, b *EventSummary) bool {
			return a.Value > b.Value
		},
	}
	sort.Sort(cs)
}

type LessFunc func(a, b *EventSummary) bool

type CollectionSort struct {
	collection SummaryCollection
	lessFunc   LessFunc
}

func (c *CollectionSort) Len() int {
	return len(c.collection)
}

func (c *CollectionSort) Less(i, j int) bool {
	return c.lessFunc(c.collection[i], c.collection[j])
}

func (c *CollectionSort) Swap(i, j int) {
	c.collection[i], c.collection[j] = c.collection[j], c.collection[i]
}
