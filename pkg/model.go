package agents

import "github.com/logkn/agents-go/internal/types"

type (
	Model       = types.ModelConfig
	ModelOption = types.ModelOption
)

func NewModel(model string, opts ...types.ModelOption) Model {
	config := types.DefaultModel(model)
	config.Apply(opts...)
	return config
}

type modelOptionFunc func(*Model) error

func (f modelOptionFunc) Apply(config *Model) error {
	return f(config)
}

func WithBaseURL(baseURL string) ModelOption {
	return modelOptionFunc(func(config *Model) error {
		config.BaseURL = baseURL
		return nil
	})
}

func WithTemperature(temperature float32) ModelOption {
	return modelOptionFunc(func(config *Model) error {
		config.Temperature = temperature
		return nil
	})
}
