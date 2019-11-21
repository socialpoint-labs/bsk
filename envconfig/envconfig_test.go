package envconfig_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/socialpoint-labs/bsk/envconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testConfig struct {
	Number int `required:"true"`
	Var string
}

func TestLoadConfig(t *testing.T) {
	testCases := map[string]struct {
		env    string
		prefix    string
		config testConfig
		expected testConfig
	}{
		"load config from an environment with its own .env file": {
			env:    "test",
			prefix:    "myservice",
			config:    testConfig{},
			expected: testConfig{Number: 42, Var: "some_test_value"},
		},
		"load config from an environment without .env file": {
			env:    "other_environment",
			prefix:    "myservice",
			config:    testConfig{},
			expected: testConfig{Number: 42, Var: "some_value"},
		},
	}
	for name, testCase := range testCases {
		tc := testCase // capture range variables
		msg := name
		t.Run(msg, func(t *testing.T) {
			a := assert.New(t)
			r := require.New(t)

			err := envconfig.LoadConfig(tc.env, tc.prefix, &tc.config)
			a.NoError(err)
			a.Equal(tc.expected, tc.config, fmt.Sprintf("checking expected in %s", msg))

			err = cleanEnvVars()
			r.NoError(err)
		})
	}
}

func TestLoadConfig_FromEnvVars(t *testing.T) {
	a := assert.New(t)
	r := require.New(t)

	expectedNumber := 28
	err := os.Setenv("MYSERVICE_NUMBER", fmt.Sprintf("%d", expectedNumber))
	r.NoError(err)

	env := "test"
	prefix := "myservice"
	config := testConfig{}
	expected := testConfig{
		Number: expectedNumber,
		Var:    "some_test_value",
	}

	err = envconfig.LoadConfig(env, prefix, &config)
	a.NoError(err)
	a.Equal(expected, config)

	err = cleanEnvVars()
	a.NoError(err)
}

func TestLoadConfig_Error(t *testing.T) {
	a := assert.New(t)

	path := "missing/path"
	env := "test"
	prefix := "myservice"
	config := testConfig{}

	err := envconfig.LoadConfig(env, prefix, &config, envconfig.WithDotEnvPath(path))
	a.Error(err)

	err = cleanEnvVars()
	a.NoError(err)
}

func TestMustLoadConfig(t *testing.T) {
	a := assert.New(t)

	env := "test"
	prefix := "myservice"
	config := testConfig{}
	expected := testConfig{
		Number: 42,
		Var:    "some_test_value",
	}

	envconfig.MustLoadConfig(env, prefix, &config)
	a.Equal(expected, config)

	err := cleanEnvVars()
	a.NoError(err)
}

func TestMustLoadConfig_Error(t *testing.T) {
	a := assert.New(t)

	path := "missing/path"
	env := "test"
	prefix := "myservice"
	config := testConfig{}

	a.Panics(func() {
		envconfig.MustLoadConfig(env, prefix, &config, envconfig.WithDotEnvPath(path))
	})
	
	err := cleanEnvVars()
	a.NoError(err)
}

func cleanEnvVars() error {
	err := os.Unsetenv("MYSERVICE_NUMBER")
	if err != nil {
		return err
	}

	err = os.Unsetenv("MYSERVICE_VAR")
	if err != nil {
		return err
	}

	return nil
}
