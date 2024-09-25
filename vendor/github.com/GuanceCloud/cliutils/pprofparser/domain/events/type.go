package events

import (
	"github.com/GuanceCloud/cliutils/pprofparser/domain/languages"
	"github.com/GuanceCloud/cliutils/pprofparser/domain/quantity"
)

const (
	DefaultMetaFileName           = "event"
	DefaultMetaFileNameWithExt    = "event.json"
	DefaultProfileFilename        = "prof"
	DefaultProfileFilenameWithExt = "prof.pprof"
)

const (
	CpuSamples       Type = "cpu-samples"
	CpuTime          Type = "cpu-time"
	WallTime         Type = "wall-time"
	HeapLiveSize     Type = "heap-space"
	HeapLiveObjects  Type = "heap-live-objects"
	Mutex            Type = "mutex"
	Block            Type = "block"
	Goroutines       Type = "goroutines"
	AllocatedMemory  Type = "alloc-space"
	Allocations      Type = "alloc-samples"
	ThrownExceptions Type = "exception-samples"
	LockWaitTime     Type = "lock-acquire-wait"
	LockedTime       Type = "lock-release-hold"
	LockAcquires     Type = "lock-acquire"
	LockReleases     Type = "lock-release"
	Other            Type = "other"
	Unknown          Type = "unknown"
)

const (
	ShowNoWay     ShowPlace = 0
	ShowInTrace   ShowPlace = 1
	ShowInProfile ShowPlace = 2
)

var TypeProfileFilename = map[languages.Lang]map[Type]string{
	languages.Python: {},

	languages.GoLang: {},
}

var Metas = map[Type]TypeMetadata{
	CpuSamples: {
		Sort:         sortMap{languages.Python: 0, languages.GoLang: 0, languages.NodeJS: 0},
		Name:         "CPU Samples",
		Description:  descriptionMap{languages.Any: "This is the number of samples each method spent running on the CPU."},
		QuantityKind: quantity.Count,
		ShowPlaces:   ShowNoWay,
	},

	CpuTime: {
		Sort:         sortMap{languages.Python: 10, languages.GoLang: 10, languages.DotNet: 10}, //map[languages.Lang]int{languages.Python: 10, languages.GoLang: 10},
		Name:         "CPU Time",
		Description:  descriptionMap{languages.Any: "This is the time each method spent running on the CPU."},
		QuantityKind: quantity.Duration,
		ShowPlaces:   ShowInTrace | ShowInProfile,
	},

	WallTime: {
		Sort:         sortMap{languages.Python: 20, languages.DotNet: 80}, //map[languages.Lang]int{languages.Python: 20, languages.GoLang: 20},
		Name:         "Wall Time",
		Description:  descriptionMap{languages.Any: "This is the elapsed time spent in each method. Elapsed time includes time when code is running on CPU, waiting for I/O, and anything else that happens while the function is running."},
		QuantityKind: quantity.Duration,
		ShowPlaces:   ShowInProfile,
	},

	HeapLiveSize: {
		Sort: sortMap{languages.Python: 30, languages.NodeJS: 30, languages.DotNet: 70}, //map[languages.Lang]int{languages.Python: 30, languages.GoLang: 30},
		Name: "Heap Live Size",
		Description: descriptionMap{languages.Any: "This is the amount of heap memory allocated that remains in use.",
			languages.GoLang: `This is the amount of heap memory allocated by each function that remains in use. (Go calls this "inuse_space").`,
		},
		QuantityKind: quantity.Memory,
		ShowPlaces:   ShowInProfile,
	},

	HeapLiveObjects: {
		Sort: sortMap{languages.Python: 31, languages.NodeJS: 31, languages.DotNet: 60}, //map[languages.Lang]int{languages.Python: 31, languages.GoLang: 31},
		Name: "Heap Live Objects",
		Description: descriptionMap{languages.Any: `This is the number of objects allocated by each function that remain in use. `,
			languages.GoLang: `This is the number of objects allocated by each function that remain in use. (Go calls this "inuse_objects").`,
		},
		QuantityKind: quantity.Count,
		ShowPlaces:   ShowInProfile,
	},

	Mutex: {
		Sort:         sortMap{languages.Python: 32}, //map[languages.Lang]int{languages.Python: 32, languages.GoLang: 32},
		Name:         "Mutex",
		Description:  descriptionMap{languages.GoLang: `This is the time each function spent waiting on mutexes during the profiling period.`},
		QuantityKind: quantity.Duration,
		ShowPlaces:   ShowInProfile,
	},

	Block: {
		Sort:         sortMap{languages.Python: 33}, //map[languages.Lang]int{languages.Python: 33, languages.GoLang: 33},
		Name:         "Block",
		Description:  descriptionMap{languages.Any: `This is the time each function spent blocked since the start of the process.`},
		QuantityKind: quantity.Duration,
		ShowPlaces:   ShowInProfile,
	},

	Goroutines: {
		Sort:         sortMap{languages.Python: 34}, //map[languages.Lang]int{languages.Python: 34, languages.GoLang: 34},
		Name:         "Goroutines",
		Description:  descriptionMap{languages.Any: `This is the number of goroutines.`},
		QuantityKind: quantity.Count,
		ShowPlaces:   ShowInProfile,
	},

	AllocatedMemory: {
		Sort: sortMap{languages.Python: 40, languages.DotNet: 40}, //map[languages.Lang]int{languages.Python: 40, languages.GoLang: 40},
		Name: "Allocated Memory",
		Description: descriptionMap{languages.Any: "This is the amount of heap memory allocated by each method, including allocations which were subsequently freed.",
			languages.GoLang: `This is the amount of heap memory allocated by each function during the profiling period, including allocations which were subsequently freed. (Go calls this "alloc_space").`,
		},
		QuantityKind: quantity.Memory,
		ShowPlaces:   ShowInProfile,
	},

	Allocations: {
		Sort: sortMap{languages.Python: 50, languages.DotNet: 30}, //map[languages.Lang]int{languages.Python: 50, languages.GoLang: 50},
		Name: "Allocations",
		Description: descriptionMap{languages.Any: "This is the number of heap allocations made by each method, including allocations which were subsequently freed.",
			languages.GoLang: "This is the number of objects allocated by each function during the profiling period, including allocations which were subsequently freed. (Go calls this \"alloc_objects\").",
		},
		QuantityKind: quantity.Count,
		ShowPlaces:   ShowInProfile,
	},

	ThrownExceptions: {
		Sort:         sortMap{languages.Python: 60, languages.DotNet: 20}, //map[languages.Lang]int{languages.Python: 60, languages.GoLang: 60},
		Name:         "Thrown Exceptions",
		Description:  descriptionMap{languages.Any: "This is the number of exceptions thrown by each method."},
		QuantityKind: quantity.Count,
		ShowPlaces:   ShowInProfile,
	},

	LockWaitTime: {
		Sort:         sortMap{languages.Python: 70, languages.DotNet: 110}, //map[languages.Lang]int{languages.Python: 70, languages.GoLang: 70},
		Name:         "Lock Wait Time",
		Description:  descriptionMap{languages.Any: "This is the time each function spent waiting for a lock."},
		QuantityKind: quantity.Duration,
		ShowPlaces:   ShowInTrace | ShowInProfile,
	},

	LockedTime: {
		Sort:         sortMap{languages.Python: 80}, //map[languages.Lang]int{languages.Python: 80, languages.GoLang: 80},
		Name:         "Locked Time",
		Description:  descriptionMap{languages.Any: "This is the time each function spent holding a lock."},
		QuantityKind: quantity.Duration,
		ShowPlaces:   ShowInTrace | ShowInProfile,
	},

	LockAcquires: {
		Sort:         sortMap{languages.Python: 90, languages.DotNet: 100}, //map[languages.Lang]int{languages.Python: 90, languages.GoLang: 90},
		Name:         "Lock Acquires",
		Description:  descriptionMap{languages.Any: "This is the number of lock acquisitions made by each method."},
		QuantityKind: quantity.Count,
		ShowPlaces:   ShowInProfile,
	},

	LockReleases: {
		Sort:         sortMap{languages.Python: 100}, //map[languages.Lang]int{languages.Python: 100, languages.GoLang: 100},
		Name:         "Lock Releases",
		Description:  descriptionMap{languages.Any: "This is the number of times each function released a lock."},
		QuantityKind: quantity.Count,
		ShowPlaces:   ShowInProfile,
	},
	Other: {
		Sort:         sortMap{languages.Python: 110}, //map[languages.Lang]int{languages.Python: 110, languages.GoLang: 110},
		Name:         "Other",
		Description:  descriptionMap{languages.Any: "Methods that used the most uncategorized time."},
		QuantityKind: quantity.Duration,
		ShowPlaces:   ShowInTrace,
	},
}

type Type string

func ParseType(typ string) Type {
	et := Type(typ)
	if _, ok := Metas[et]; ok {
		return et
	}
	return Unknown
}

func (et Type) String() string {
	return string(et)
}

func (et Type) Equals(target Type) bool {
	return et.String() == target.String()
}

func (et Type) GetSort(lang languages.Lang) int {
	if meta, ok := Metas[et]; ok {
		if sort, ok := meta.Sort[lang]; ok {
			return sort
		}
		if sort, ok := meta.Sort[languages.Any]; ok {
			return sort
		}
	}
	return 1 << 30
}

func (et Type) GetName() string {
	if meta, ok := Metas[et]; ok {
		return meta.Name
	}
	return "unknown"
}

func (et Type) GetDescription(lang languages.Lang) string {
	if meta, ok := Metas[et]; ok {
		if desc, ok := meta.Description[lang]; ok {
			return desc
		}
		if desc, ok := meta.Description[languages.Any]; ok {
			return desc
		}
	}
	return "unknown metric type"
}

func (et Type) GetQuantityKind() *quantity.Kind {
	if meta, ok := Metas[et]; ok {
		return meta.QuantityKind
	}
	return quantity.UnknownKind
}

func (et Type) GetShowPlaces() ShowPlace {
	if meta, ok := Metas[et]; ok {
		return meta.ShowPlaces
	}
	return ShowNoWay
}

type TypeMetadata struct {
	Sort         sortMap
	Name         string
	Description  descriptionMap
	QuantityKind *quantity.Kind
	ShowPlaces   ShowPlace
}

type ShowPlace int

// sortMap used to generate sort map for convenience
type sortMap map[languages.Lang]int

func newSortMap() sortMap {
	return make(sortMap)
}

func (sm sortMap) put(lang languages.Lang, sort int) sortMap {
	sm[lang] = sort
	return sm
}

type descriptionMap map[languages.Lang]string

func newDescriptionMap() descriptionMap {
	return make(descriptionMap)
}

func (dm descriptionMap) put(lang languages.Lang, desc string) descriptionMap {
	dm[lang] = desc
	return dm
}
