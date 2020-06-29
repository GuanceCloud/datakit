package inputs

import (
	"time"
)

type RunningInput struct {
	Input       Input
	Config      *InputConfig
	defaultTags map[string]string
}

func NewRunningInput(input Input, config *InputConfig) *RunningInput {
	tags := map[string]string{"input": config.Name}
	if config.Alias != "" {
		tags["alias"] = config.Alias
	}

	return &RunningInput{
		Input:  input,
		Config: config,
	}
}

// InputConfig is the common config for all inputs.
type InputConfig struct {
	Name     string
	Alias    string
	Interval time.Duration

	Tags map[string]string

	OutputFile string
}

func (r *RunningInput) SetDefaultTags(tags map[string]string) {
	r.defaultTags = tags
}
