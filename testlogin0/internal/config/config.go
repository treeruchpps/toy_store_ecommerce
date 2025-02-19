// config.go
package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	AppPort        string
	DatabaseURL    string
	GoogleClientID string
	JWTSecret      string
	APIKey         string
}

func New() (*Config, error) {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	config := &Config{
		AppPort:        viper.GetString("APP_PORT"),
		GoogleClientID: viper.GetString("GOOGLE_CLIENT_ID"),
		JWTSecret:      viper.GetString("JWT_SECRET"),
		APIKey:         viper.GetString("API_KEY"),
	}

	// Construct DatabaseURL
	config.DatabaseURL = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		viper.GetString("POSTGRES_HOST"),
		viper.GetString("POSTGRES_PORT"),
		viper.GetString("POSTGRES_USER"),
		viper.GetString("POSTGRES_PASSWORD"),
		viper.GetString("POSTGRES_DBNAME"),
		viper.GetString("POSTGRES_SSLMODE"),
	)

	return config, nil
}
