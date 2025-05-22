package state

// WeatherStateReader interface for weather-related state access
type WeatherStateReader interface {
	GetWeatherAPIKey() string
	GetDefaultUnits() string
}

// EmailStateReader interface for email-related state access
type EmailStateReader interface {
	GetEmailConfig() EmailConfig
}

// EmailConfig holds email configuration
type EmailConfig struct {
	SMTPHost string
	SMTPPort int
	Username string
	Password string
}

// ExampleGlobalState is an example implementation that implements multiple state interfaces
type ExampleGlobalState struct {
	WeatherAPIKey string
	DefaultUnits  string
	EmailConfig   EmailConfig
}

func (s *ExampleGlobalState) GetWeatherAPIKey() string {
	return s.WeatherAPIKey
}

func (s *ExampleGlobalState) GetDefaultUnits() string {
	return s.DefaultUnits
}

func (s *ExampleGlobalState) GetEmailConfig() EmailConfig {
	return s.EmailConfig
}
