package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Configuration struct {
	ApplicationConfig ApplicationConfig
	DatabaseConfig    DatabaseConfig
	ServerConfig      ServerConfig
}

type ApplicationConfig struct {
	Name           string
	Env            string
	EnvConfig      EvnConfig
	TelegramConfig TelegramConfig
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	TLS          TLSServerConfig
}

type TLSServerConfig struct {
	Mode     string
	CertFile string
	KeyFile  string
}
type EvnConfig struct {
	Username                  string
	Password                  string
	CustomerCode              string
	BaseURL                   string
	LoginAPI                  string
	ElectricityConsumptionAPI string
}

type TelegramConfig struct {
	TelegramAPIKey string
}

type DatabaseConfig struct {
	Host         string
	Port         string
	User         string
	Password     string
	DatabaseName string
	SSLMode      string

	MaxOpenConnections int
	MaxIdleConnections int
	MaxIdleTime        time.Duration
}

func Load() *Configuration {
	return &Configuration{
		ApplicationConfig: ApplicationConfig{
			Name: GetEnv("APP_NAME", "social"),
			Env:  GetEnv("APP_ENV", "development"),
			EnvConfig: EvnConfig{
				Username:                  GetEnv("EVN_USERNAME", ""),
				Password:                  GetEnv("EVN_PASSWORD", ""),
				CustomerCode:              GetEnv("EVN_CUSTOMER", ""),
				BaseURL:                   GetEnv("EVN_BASE_URL", ""),
				LoginAPI:                  GetEnv("EVN_LOGIN_API", ""),
				ElectricityConsumptionAPI: GetEnv("EVN_ELECTRICITY_CONSUMPTION_API", ""),
			},
			TelegramConfig: TelegramConfig{
				TelegramAPIKey: GetEnv("TELEGRAM_API_KEY", ""),
			},
		},
		DatabaseConfig: DatabaseConfig{
			Host:               GetEnv("DB_HOST", "localhost"),
			Port:               GetEnv("DB_PORT", "5432"),
			User:               GetEnv("DB_USER", "admin"),
			Password:           GetEnv("DB_PASS", "password"),
			DatabaseName:       GetEnv("DB_NAME", "social"),
			SSLMode:            GetEnv("DB_SSL_MODE", "disable"),
			MaxOpenConnections: GetEnvInt("DB_MAX_OPEN_CONNECTIONS", 10),
			MaxIdleConnections: GetEnvInt("DB_MAX_IDLE_CONNECTIONS", 10),
			MaxIdleTime:        GetEnvDuration("DB_MAX_IDLE_TIME", 15*time.Minute),
		},
		ServerConfig: ServerConfig{
			Port:         GetEnv("SERVER_PORT", "8080"),
			ReadTimeout:  GetEnvDuration("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: GetEnvDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
			TLS: TLSServerConfig{
				Mode:     GetEnv("SERVER_TLS_MODE", "off"),
				CertFile: GetEnv("SERVER_TLS_CERT_FILE", ""),
				KeyFile:  GetEnv("SERVER_TLS_KEY_FILE", ""),
			},
		},
	}
}

func (c DatabaseConfig) DatabaseDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.DatabaseName,
		c.SSLMode,
	)
}

func GetEnv(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return value
}

func GetEnvInt(key string, fallback int) int {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	valAsInt, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return valAsInt
}

func GetEnvDuration(key string, fallback time.Duration) time.Duration {
	val := GetEnv(key, "")
	if val == "" {
		return fallback
	}

	d, err := time.ParseDuration(val)
	if err != nil {
		return fallback
	}

	return d
}
