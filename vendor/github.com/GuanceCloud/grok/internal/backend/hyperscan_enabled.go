//go:build hyperscan && cgo

package backend

const HyperscanPlaceholderBackendName = "hyperscan-placeholder"

type hyperscanBackend struct{}

type placeholderMultiPatternFilter struct{}

func Default() Backend {
	return hyperscanBackend{}
}

func (hyperscanBackend) Name() string {
	return HyperscanPlaceholderBackendName
}

func (hyperscanBackend) Available() bool {
	return false
}

func (hyperscanBackend) Compile(patterns []string) (Filter, error) {
	return &placeholderMultiPatternFilter{}, nil
}

func (hyperscanBackend) Load(database []byte) (Filter, error) {
	return &placeholderMultiPatternFilter{}, nil
}

func (f *placeholderMultiPatternFilter) Backend() string {
	return HyperscanPlaceholderBackendName
}

func (f *placeholderMultiPatternFilter) MatchString(string) bool {
	return true
}

func (f *placeholderMultiPatternFilter) Close() error {
	return nil
}
