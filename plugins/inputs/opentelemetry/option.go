package opentelemetry

type option struct {
	// add to point.
	globalTags map[string]string

	// that want to send to data center.
	customTags map[string]struct{}
}
