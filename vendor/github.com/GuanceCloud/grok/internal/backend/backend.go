package backend

type Filter interface {
	Backend() string
	MatchString(content string) bool
	Close() error
}

type Backend interface {
	Name() string
	Available() bool
	Compile(patterns []string) (Filter, error)
	Load(database []byte) (Filter, error)
}

func CompileMultiPatternFilter(patterns []string) (Filter, error) {
	return Default().Compile(patterns)
}

func LoadMultiPatternFilter(database []byte) (Filter, error) {
	return Default().Load(database)
}
