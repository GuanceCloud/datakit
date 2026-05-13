//go:build !hyperscan || !cgo

package backend

const HyperscanDisabledBackendName = "hyperscan-disabled"

type hyperscanBackend struct{}

type noOpMultiPatternFilter struct {
	backend string
}

func Default() Backend {
	return hyperscanBackend{}
}

func (hyperscanBackend) Name() string {
	return HyperscanDisabledBackendName
}

func (hyperscanBackend) Available() bool {
	return false
}

func (hyperscanBackend) Compile(patterns []string) (Filter, error) {
	return &noOpMultiPatternFilter{backend: HyperscanDisabledBackendName}, nil
}

func (hyperscanBackend) Load(database []byte) (Filter, error) {
	return &noOpMultiPatternFilter{backend: HyperscanDisabledBackendName}, nil
}

func (f *noOpMultiPatternFilter) Backend() string {
	return f.backend
}

func (f *noOpMultiPatternFilter) MatchString(string) bool {
	return true
}

func (f *noOpMultiPatternFilter) Close() error {
	return nil
}
