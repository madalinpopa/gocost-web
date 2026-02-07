package config

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/madalinpopa/gocost-web/internal/platform/money"
	"github.com/spf13/viper"
)

const defaultCurrency = "USD"

type Config struct {
	// Version specifies the application version
	Version string

	// Addr specifies the address for the service
	Addr string

	// Port specifies the port number for the service
	Port int

	// Dsn specifies the database connection string
	Dsn string

	// AllowedHosts specifies the allowed hosts for the application.
	AllowedHosts []string

	// TrustedProxies specifies the trusted proxies for the application.
	TrustedProxies []string

	// Domain specifies the application web domain
	Domain string

	// Currency specifies the currency symbol
	Currency string

	// logger is used for config-level logging.
	logger *slog.Logger

	// environment specifies the application Environment
	environment string
}

func New() *Config {
	return NewWithLogger(nil)
}

func NewWithLogger(logger *slog.Logger) *Config {
	return &Config{
		AllowedHosts:   make([]string, 0),
		TrustedProxies: make([]string, 0),
		logger:         logger,
	}
}

func (c *Config) WithEnvironment(env string) *Config {
	c.environment = env
	return c
}

func (c *Config) WithServerAddr(addr string, port int) *Config {
	c.Addr = addr
	c.Port = port
	return c
}

func (c *Config) WithDatabaseDsn(dsn string) *Config {
	c.Dsn = dsn
	return c
}

// LoadEnvironments loads environment variables into the Config struct using Viper's automatic environment handling.
func (c *Config) LoadEnvironments() error {
	viper.AutomaticEnv()

	// Helps with handling underscores in env vars
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	required := []string{
		"ALLOWED_HOSTS",
		"DOMAIN",
	}

	for _, env := range required {
		if viper.GetString(env) == "" {
			return fmt.Errorf("env %s is not set", env)
		}
	}

	c.AllowedHosts = c.getStringSliceFromEnv("ALLOWED_HOSTS")
	c.TrustedProxies = c.getStringSliceFromEnv("TRUSTED_PROXIES")
	c.Domain = viper.GetString("DOMAIN")

	c.Currency = c.currencyCodeOrDefault(viper.GetString("CURRENCY"))

	return nil
}

func (c *Config) getStringSliceFromEnv(key string) []string {
	slice := viper.GetStringSlice(key)
	// If the environment variable is passed as a comma-separated string (e.g. via Docker),
	// Viper might treat it as a single element slice. We need to handle this manually.
	if len(slice) == 1 && strings.Contains(slice[0], ",") {
		return strings.Split(slice[0], ",")
	}
	// Also support cases where Viper returns empty slice but the env var is set as string
	if len(slice) == 0 {
		str := viper.GetString(key)
		if str != "" {
			return strings.Split(str, ",")
		}
	}
	return slice
}

func (c *Config) GetEnvironment() string {
	return c.environment
}

func (c *Config) currencyCodeOrDefault(value string) string {
	currency, err := money.New(0, value)
	if err != nil {
		c.loggerOrDefault().Error("invalid currency code; using default", "currency", value, "err", err)
		return defaultCurrency
	}

	return currency.Currency()
}

func (c *Config) loggerOrDefault() *slog.Logger {
	if c.logger == nil {
		return slog.Default()
	}
	return c.logger
}
