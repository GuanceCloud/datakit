package protocol

var ProtocolDic = map[string]bool{
	"HTTP":  true,
	"TCP":   true,
	"UDP":   true,
	"MYSQL": true,
	"DNS":   true,
	"SSH":   false,
}

type Item struct {
	Header string
	Field  string
	Tag    bool
	Type   string
}

var CommonItems = []*Item{
	&Item{
		Header: "timestamp",
		Field:  "frame.time_epoch",
		Type:   "Timestamp",
	},
	&Item{
		Header: "seq",
		Field:  "frame.number",
		Type:   "String",
	},
	&Item{
		Header: "source",
		Field:  "_ws.col.Source",
		Type:   "String",
	},
	&Item{
		Header: "destination",
		Field:  "_ws.col.Destination",
		Type:   "String",
	},
	&Item{
		Header: "protocol",
		Field:  "_ws.col.Protocol",
		Tag:    true,
		Type:   "String",
	},
	&Item{
		Header: "length",
		Field:  "_ws.col.Length",
		Type:   "Int",
	},
	&Item{
		Header: "info",
		Field:  "_ws.col.Info",
		Type:   "String",
	},
}

func GetFiled() []string {
	args := make([]string, 0)
	for _, item := range CommonItems {
		args = append(args, "-e", item.Field)
	}

	return args
}
