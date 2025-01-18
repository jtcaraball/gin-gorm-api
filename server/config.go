package server

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/sethvargo/go-envconfig"
	"gopkg.in/yaml.v3"
)

// Config for all server dependencies.
type Config struct {
	Debug   bool         `yaml:"debug"   env:"DEBUG, overwrite"`
	Testing bool         `yaml:"testing" env:"TESTING, overwrite"`
	Secret  string       `yaml:"secret"  env:"SECRET, overwrite"`
	DB      DBConfig     `yaml:"db"`
	Engine  EngineConfig `yaml:"engine"`
}

// EngineConfig holds the config info for the http engine.
type EngineConfig struct {
	AllowedHost    []string `yaml:"allowed_hosts" env:"ALLOWED_HOSTS, overwrite"`     //nolint:lll // annotaions dont allow new lines.
	TrustedProxies []string `yaml:"trusted_proxies" env:"TRUSTED_PROXIES, overwrite"` //nolint:lll // annotaions dont allow new lines.
}

// EngineConfig holds the config info for the database.
type DBConfig struct {
	Host     string `yaml:"host"     env:"DB_HOST, overwrite"`
	Port     int    `yaml:"port"     env:"DB_PORT, overwrite"`
	Name     string `yaml:"name"     env:"DB_NAME, overwrite"`
	User     string `yaml:"user"     env:"DB_USER, overwrite"`
	Password string `yaml:"password" env:"DB_PASSWORD, overwrite"`
	SSL      string `yaml:"ssl"      env:"DB_SSL, overwrite"`
}

// LoadConfig from environment variables and, optionally, from a yaml formatted
// word passed as an io.Reader r. Environment variables take precedence over
// those defined in r.
func LoadConfig(r ...io.Reader) (Config, error) {
	var (
		config Config
		err    error
	)
	if err = config.handleReaders(r); err != nil {
		return config,
			fmt.Errorf("failed to parse configuration file: %w", err)
	}
	if err = envconfig.Process(context.Background(), &config); err != nil {
		return config,
			fmt.Errorf("failed to parse environment variables: %w", err)
	}
	return config, err
}

// handleReaders attempts to unmarshal the contents of the first reader of r
// into c. If no readers are passed this is a noop and if more than one is
// passed an error is returned.
func (c *Config) handleReaders(r []io.Reader) error {
	switch len(r) {
	case 0:
		return nil
	case 1:
		return yaml.NewDecoder(r[0]).Decode(c)
	default:
		return errors.New("too many inputs")
	}
}
