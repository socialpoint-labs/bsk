package envconfig

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

const (
	defaultEnv = "_default"
	envFolder  = "envs"
)

type Option func(*loadOptions)

type loadOptions struct {
	dotEnvPath string
}

func WithDotEnvPath(path string) Option {
	return func(o *loadOptions) {
		o.dotEnvPath = path
	}
}

// Load is a helper function that load the config
func LoadConfig(env string, prefix string, config interface{}, opts ...Option) error {
	options := &loadOptions{}
	for _, o := range opts {
		o(options)
	}

	if options.dotEnvPath == "" {
		options.dotEnvPath = "./"
	}

	// first we load application env variables from envs/xxx.env
	err := loadEnvVars(options.dotEnvPath, env)
	if err != nil {
		return err
	}

	// then load the env variables into the Config
	return envconfig.Process(prefix, config)
}

// MustLoad is a helper function that load the config
// and panics if an error is encountered.
func MustLoadConfig(env string, prefix string, config interface{}, opts ...Option) {
	err := LoadConfig(env, prefix, config, opts...)
	if err != nil {
		panic(err)
	}
}

func loadEnvVars(path string, env string) error {
	var configFiles []string

	configFile := fmt.Sprintf("%s%s/%s.env", path, envFolder, env)
	if _, err := os.Stat(configFile); err == nil {
		// load the specific env var if the file exist
		configFiles = append(configFiles, configFile)
	}

	// load the default env var
	configFiles = append(configFiles, fmt.Sprintf("%s%s/%s.env", path, envFolder, defaultEnv))

	// here xxx.env will first be loaded, and then the default.env will be loaded
	// if a var is repeated in both files, only the first load (from the xxx.env) will be the good one
	// and if a var is already loaded by environment variables, it will not be override
	// godotenv.Load does not override existing env var
	// env var priority is:
	// - first: the one defined by environment variables
	// - then: the one defined in xxx.env (pre, pro, int...)
	// - last: the one in _default.env
	err := godotenv.Load(configFiles...)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not load the environment variables from file: %s", err.Error()))
	}

	return nil
}