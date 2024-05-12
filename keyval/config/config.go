package config

import (
	"fmt"
)

// APIConfig
type APIConfig struct {
	Host string
	Port int
}

// Address
func (app *APIConfig) Address() string {
	return fmt.Sprintf("%s:%d", app.Host, app.Port)
}

// Initialize the service
func Init() *APIConfig {
	config := new(APIConfig)
	config.Host = "" // 0.0.0.0
	config.Port = 5000
	return config
}
