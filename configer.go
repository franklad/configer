package configer

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Configer defines an interface for accessing configuration values.
// It provides type-safe methods for common data types, checks for key existence,
// supports nested configurations via Sub, and allows retrieving all settings.
type Configer interface {
	Bool(key string) bool
	Int(key string) int
	Int64(key string) int64
	Uint(key string) uint
	Float64(key string) float64
	String(key string) string
	Strings(key string) []string
	StringMap(key string) map[string]interface{}
	Duration(key string) time.Duration
	Time(key string) time.Time
	Exists(key string) bool
	Sub(key string) Configer
	AllSettings() map[string]interface{}
}

type viperConfiger struct {
	v *viper.Viper
}

// config holds internal configuration options for building the Configer.
type config struct {
	configType          string
	defaultConfigFile   string
	envVarName          string
	envConfigFilePrefix string
	envConfigFileSuffix string
	autoEnv             bool
	envPrefix           string
	bindEnvKeys         []string
	viperConfig         func(*viper.Viper)
}

// option is a functional option for configuring the Configer.
type option func(*config)

// WithConfigType sets the configuration file type (e.g., "toml", "yaml").
func WithConfigType(t string) option {
	return func(c *config) {
		c.configType = t
	}
}

// WithDefaultConfigFile sets the path to the default configuration file.
func WithDefaultConfigFile(path string) option {
	return func(c *config) {
		c.defaultConfigFile = path
	}
}

// WithEnvVarName sets the environment variable name for determining the environment-specific config file.
func WithEnvVarName(name string) option {
	return func(c *config) {
		c.envVarName = name
	}
}

// WithEnvConfigFilePrefix sets the prefix for environment-specific config files.
func WithEnvConfigFilePrefix(prefix string) option {
	return func(c *config) {
		c.envConfigFilePrefix = prefix
	}
}

// WithEnvConfigFileSuffix sets the suffix for environment-specific config files (e.g., ".toml").
func WithEnvConfigFileSuffix(suffix string) option {
	return func(c *config) {
		c.envConfigFileSuffix = suffix
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

// WithCustomViper allows custom configuration of the underlying Viper instance.
// This can be used for advanced features like adding more paths, binding flags, or setting up watching.
func WithCustomViper(f func(*viper.Viper)) option {
	return func(c *config) {
		c.viperConfig = f
	}
}

func defaults() *config {
	return &config{
		configType:          "toml",
		defaultConfigFile:   "config/default.toml",
		envVarName:          "ENV",
		envConfigFilePrefix: "config/",
		envConfigFileSuffix: ".toml",
		autoEnv:             true,
		envPrefix:           "",
		bindEnvKeys:         []string{},
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

	if config.viperConfig != nil {
		config.viperConfig(v)
	}

	v.SetConfigType(config.configType)
	v.SetConfigFile(config.defaultConfigFile)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to load default config file %s: %w", config.defaultConfigFile, err)
	}

	if value, ok := os.LookupEnv(config.envVarName); ok {
		envFile := config.envConfigFilePrefix + strings.ToLower(value) + config.envConfigFileSuffix
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

	return &viperConfiger{v: v}, nil
}

func (c *viperConfiger) Bool(key string) bool {
	return c.v.GetBool(key)
}

func (c *viperConfiger) Int(key string) int {
	return c.v.GetInt(key)
}

func (c *viperConfiger) Int64(key string) int64 {
	return c.v.GetInt64(key)
}

func (c *viperConfiger) Uint(key string) uint {
	return c.v.GetUint(key)
}

func (c *viperConfiger) Float64(key string) float64 {
	return c.v.GetFloat64(key)
}

func (c *viperConfiger) String(key string) string {
	return c.v.GetString(key)
}

func (c *viperConfiger) Strings(key string) []string {
	return c.v.GetStringSlice(key)
}

func (c *viperConfiger) StringMap(key string) map[string]interface{} {
	return c.v.GetStringMap(key)
}

func (c *viperConfiger) Duration(key string) time.Duration {
	return c.v.GetDuration(key)
}

func (c *viperConfiger) Time(key string) time.Time {
	return c.v.GetTime(key)
}

func (c *viperConfiger) Exists(key string) bool {
	return c.v.IsSet(key)
}

func (c *viperConfiger) Sub(key string) Configer {
	sub := c.v.Sub(key)
	if sub == nil {
		return &viperConfiger{v: viper.New()}
	}
	return &viperConfiger{v: sub}
}

func (c *viperConfiger) AllSettings() map[string]interface{} {
	return c.v.AllSettings()
}
