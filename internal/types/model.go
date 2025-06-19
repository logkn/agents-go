package types

// ModelConfig contains configuration details for an LLM model.
// Model is the identifier of the model to use and BaseUrl is an optional
// override for the API base URL.
type ModelConfig struct {
	Model       string
	BaseURL     string
	Temperature float32
}

type ModelOption interface {
	Apply(config *ModelConfig) error
}

func (config *ModelConfig) Apply(opts ...ModelOption) error {
	for _, opt := range opts {
		if err := opt.Apply(config); err != nil {
			return err
		}
	}
	return nil
}

func DefaultModel(model string) ModelConfig {
	return ModelConfig{
		Model: model,
		// BaseURL:     "http://localhost:11434/v1",
		Temperature: 0.6,
	}
}
