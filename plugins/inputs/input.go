package inputs

import "time"

type RunningInput struct {
	Input  Input
	Config *InputConfig
}

func NewRunningInput(input Input, config *InputConfig) *RunningInput {

	return &RunningInput{
		Input:  input,
		Config: config,
	}
}

// InputConfig is the common config for all inputs.
type InputConfig struct {
	Name     string
	Tags     map[string]string
	Interval time.Duration
}
