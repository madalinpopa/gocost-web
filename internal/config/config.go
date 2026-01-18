package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

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

	// Domain specifies the application web domain
	Domain string

	// Currency specifies the currency symbol
	Currency string

	// environment specifies the application Environment
	environment string
}

func New() *Config {
	return &Config{AllowedHosts: make([]string, 0)}
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

	c.AllowedHosts = viper.GetStringSlice("ALLOWED_HOSTS")
	// If the environment variable is passed as a comma-separated string (e.g. via Docker),
	// Viper might treat it as a single element slice. We need to handle this manually.
	if len(c.AllowedHosts) == 1 && strings.Contains(c.AllowedHosts[0], ",") {
		c.AllowedHosts = strings.Split(c.AllowedHosts[0], ",")
	}
	// Also support cases where Viper returns empty slice but the env var is set as string
	if len(c.AllowedHosts) == 0 {
		str := viper.GetString("ALLOWED_HOSTS")
		if str != "" {
			c.AllowedHosts = strings.Split(str, ",")
		}
	}
	c.Domain = viper.GetString("DOMAIN")

	c.Currency = viper.GetString("CURRENCY")
	if c.Currency == "" {
		c.Currency = "$"
	}

	return nil
}

func (c *Config) GetEnvironment() string {
	return c.environment
}
