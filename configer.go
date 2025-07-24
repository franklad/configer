package configer

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Configer defines an interface for accessing configuration values.
// It provides type-safe methods for common data types, checks for key existence
type Configer interface {
	Bool(key string) bool
	Int(key string) int
	Int64(key string) int64
	Uint(key string) uint
	Float64(key string) float64
	String(key string) string
	Strings(key string) []string
	StringMap(key string) map[string]any
	Duration(key string) time.Duration
	Time(key string) time.Time
	Exists(key string) bool
}

type configer struct {
	viper *viper.Viper
}

// config holds internal configuration options for building the Configer.
type config struct {
	configFileType   string
	configFilePrefix string
	envVarName       string
	autoEnv          bool
	envPrefix        string
	bindEnvKeys      []string
}

// option is a functional option for configuring the Configer.
type option func(*config)

// WithConfigFileType sets the configuration file type (e.g., "toml", "yaml").
func WithConfigFileType(t string) option {
	return func(c *config) {
		c.configFileType = t
	}
}

// WithEnvConfigFilePrefix sets the prefix for environment-specific config files.
func WithEnvConfigFilePrefix(prefix string) option {
	return func(c *config) {
		c.configFilePrefix = prefix
	}
}

// WithEnvVarName sets the environment variable name for determining the environment-specific config file.
func WithEnvVarName(name string) option {
	return func(c *config) {
		c.envVarName = name
	}
}

// WithAutomaticEnv enables automatic binding of environment variables.
// Recommended to pair with WithEnvPrefix to avoid collisions.
func WithAutomaticEnv() option {
	return func(c *config) {
		c.autoEnv = true
	}
}

// WithEnvPrefix sets the prefix for environment variables.
func WithEnvPrefix(prefix string) option {
	return func(c *config) {
		c.envPrefix = prefix
	}
}

// WithBindEnv sets specific keys to bind to environment variables explicitly.
// This provides more control than AutomaticEnv, binding only listed keys.
// Uses the standard naming convention (e.g., "database.host" binds to "DATABASE_HOST").
// Can be used with or without AutomaticEnv.
func WithBindEnv(keys ...string) option {
	return func(c *config) {
		c.bindEnvKeys = keys
	}
}

func defaults() *config {
	return &config{
		configFileType:   "toml",
		configFilePrefix: "config/",
		envVarName:       "ENV",
		autoEnv:          true,
		envPrefix:        "",
		bindEnvKeys:      []string{},
	}
}

// New creates a new Configer instance with the given options.
// It loads the default config file, merges an environment-specific file if applicable,
// binds env vars (automatic or explicit), and applies other configurations.
// Returns an error on failure instead of panicking.
func New(opts ...option) (Configer, error) {
	v := viper.New()

	config := defaults()
	for _, opt := range opts {
		opt(config)
	}

	v.SetConfigType(config.configFileType)
	v.SetConfigFile(config.configFilePrefix + "default." + config.configFileType)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to load default config file: %w", err)
	}

	if value, ok := os.LookupEnv(config.envVarName); ok {
		envFile := config.configFilePrefix + strings.ToLower(value) + "." + config.configFileType
		v.SetConfigFile(envFile)
		if err := v.MergeInConfig(); err != nil {
			if _, notFound := err.(viper.ConfigFileNotFoundError); !notFound {
				return nil, fmt.Errorf("failed to load env config file %s: %w", envFile, err)
			}
		}
	}

	if config.envPrefix != "" {
		v.SetEnvPrefix(config.envPrefix)
	}

	for _, key := range config.bindEnvKeys {
		if err := v.BindEnv(key); err != nil {
			return nil, fmt.Errorf("failed to bind env for key %s: %w", key, err)
		}
	}

	if config.autoEnv {
		v.AutomaticEnv()
	}

	return &configer{viper: v}, nil
}

func (c *configer) Bool(key string) bool {
	return c.viper.GetBool(key)
}

func (c *configer) Int(key string) int {
	return c.viper.GetInt(key)
}

func (c *configer) Int64(key string) int64 {
	return c.viper.GetInt64(key)
}

func (c *configer) Uint(key string) uint {
	return c.viper.GetUint(key)
}

func (c *configer) Float64(key string) float64 {
	return c.viper.GetFloat64(key)
}

func (c *configer) String(key string) string {
	return c.viper.GetString(key)
}

func (c *configer) Strings(key string) []string {
	return c.viper.GetStringSlice(key)
}

func (c *configer) StringMap(key string) map[string]interface{} {
	return c.viper.GetStringMap(key)
}

func (c *configer) Duration(key string) time.Duration {
	return c.viper.GetDuration(key)
}

func (c *configer) Time(key string) time.Time {
	return c.viper.GetTime(key)
}

func (c *configer) Exists(key string) bool {
	return c.viper.IsSet(key)
}
