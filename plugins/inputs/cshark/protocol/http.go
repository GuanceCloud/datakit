package protocol

var HttpItems = []*Item{
	&Item{
		Header: "host",
		Field:  "http.host",
		Type:   "String",
	},
	&Item{
		Header: "uri",
		Field:  "http.request.uri",
		Type:   "String",
	},
	&Item{
		Header: "method",
		Field:  "http.request.method",
		Tag:    true,
		Type:   "String",
	},
	&Item{
		Header: "accept",
		Field:  "http.accept",
		Type:   "String",
	},
	&Item{
		Header: "cache_control",
		Field:  "http.cache_control",
		Type:   "String",
	},
	&Item{
		Header: "content_length",
		Field:  "http.content_length",
		Type:   "Int",
	},
	&Item{
		Header: "content_type",
		Field:  "http.content_type",
		Type:   "String",
	},
	&Item{
		Header: "x_forwarded_for",
		Field:  "http.x_forwarded_for",
		Type:   "String",
	},
}
