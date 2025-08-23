package config

import "time"

type (
	App struct {
		Name        string      `mapstructure:"name"`
		Version     string      `mapstructure:"version"`
		Environment Environment `mapstructure:"environment"`
	}

	HTTP struct {
		Host            string        `mapstructure:"host"`
		Port            string        `mapstructure:"port"`
		ReadTimeout     time.Duration `mapstructure:"read_timeout"`
		WriteTimeout    time.Duration `mapstructure:"write_timeout"`
		ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	}

	DB struct {
		DSN string `mapstructure:"dsn"`
	}

	// parallelism: 1
	// memory: 128MB
	// iterations: 2
	// salt_length: 16
	// key_length: 16

	Argon struct {
		Parallelism int `mapstructure:"parallelism"`
		Memory      int `mapstructure:"memory"`
		Iterations  int `mapstructure:"iterations"`
		SaltLength  int `mapstructure:"salt_length"`
		KeyLength   int `mapstructure:"key_length"`
	}

	TokenConfig struct {
		Algorithm string `mapstructure:"algorithm"`
		CertPath  string `mapstructure:"cert_path"`
		KeyPath   string `mapstructure:"key_path"`
		SecretKey string `mapstructure:"secret_key"`
	}
)
