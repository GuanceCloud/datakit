package jvm

import "fmt"

var reqObjs = []jolokiaRequest{
	{
		"read",
		"java.lang:type=Memory",
		"HeapMemoryUsage",
		"init",
		nil,

		"heap_memory_init",
	},

	{
		"read",
		"java.lang:type=Memory",
		"HeapMemoryUsage",
		"committed",
		nil,

		"heap_memory_committed",
	},

	{
		"read",
		"java.lang:type=Memory",
		"HeapMemoryUsage",
		"max",
		nil,

		"heap_memory_max",
	},

	{
		"read",
		"java.lang:type=Memory",
		"HeapMemoryUsage",
		"used",
		nil,

		"heap_memory",
	},

	{
		"read",
		"java.lang:type=Memory",
		"NonHeapMemoryUsage",
		"init",
		nil,
		"non_heap_memory_init",
	},

	{
		"read",
		"java.lang:type=Memory",
		"NonHeapMemoryUsage",
		"committed",
		nil,
		"non_heap_memory_committed",
	},

	{
		"read",
		"java.lang:type=Memory",
		"NonHeapMemoryUsage",
		"max",
		nil,
		"non_heap_memory_max",
	},

	{
		"read",
		"java.lang:type=Memory",
		"NonHeapMemoryUsage",
		"used",
		nil,
		"non_heap_memory",
	},

	{
		"read",
		"java.lang:type=Threading",
		"ThreadCount",
		"",
		nil,
		"thread_count",
	},

	{
		"read",
		"java.lang:name=G1 Young Generation,type=GarbageCollector",
		"CollectionCount",
		"",
		nil,
		"minor_collection_count",
	},

	{
		"read",
		"java.lang:name=G1 Young Generation,type=GarbageCollector",
		"CollectionTime",
		"",
		nil,
		"minor_collection_time",
	},

	{
		"read",
		"java.lang:name=G1 Old Generation,type=GarbageCollector",
		"CollectionCount",
		"",
		nil,
		"major_collection_count",
	},

	{
		"read",
		"java.lang:name=G1 Old Generation,type=GarbageCollector",
		"CollectionTime",
		"",
		nil,
		"major_collection_time",
	},
}

var convertDict = map[string]string{}

func initConvertDict() {
	for _, obj := range reqObjs {
		key := genKey(obj.Mbean, obj.Attribute, obj.Path)
		convertDict[key] = obj.name
	}
}

func genKey(mbean string, attr interface{}, path string) string {
	return fmt.Sprintf("%s-%s-%s", mbean, attr, path)
}
