package dotenv_test

import (
	"encoding"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/profclems/go-dotenv"
)

func testReadEnvAndCompare(t *testing.T, envFileName string, expectedValues map[string]string) {
	dotenv := dotenv.New()
	dotenv.SetConfigFile(envFileName)
	err := dotenv.Load()
	if err != nil {
		t.Error("Error loading config", err)
	}

	for key, value := range expectedValues {
		assert.Equal(t, value, dotenv.GetString(key))
	}
}

func TestReadPlainEnv(t *testing.T) {
	envFileName := "fixtures/plain.env"
	expectedValues := map[string]string{
		"OPTION_A": "1",
		"OPTION_B": "2",
		"OPTION_C": "3",
		"OPTION_D": "4",
		"OPTION_E": "5",
		"OPTION_F": "",
		"OPTION_G": "",
		"OPTION_H": "my string",
	}

	testReadEnvAndCompare(t, envFileName, expectedValues)
}

func TestLoadUnquotedEnv(t *testing.T) {
	envFileName := "fixtures/unquoted.env"
	expectedValues := map[string]string{
		"OPTION_A": "some quoted phrase",
		"OPTION_B": "first one with an unquoted phrase",
		"OPTION_C": "then another one with an unquoted phrase",
		"OPTION_D": "then another one with an unquoted phrase special è char",
		"OPTION_E": "then another one quoted phrase",
	}

	testReadEnvAndCompare(t, envFileName, expectedValues)
}

func TestLoadQuotedEnv(t *testing.T) {
	//t.Skip()
	envFileName := "fixtures/quoted.env"
	expectedValues := map[string]string{
		"OPTION_A": "1",
		"OPTION_B": "2",
		"OPTION_C": "",
		"OPTION_D": "\\n",
		"OPTION_E": "1",
		"OPTION_F": "2",
		"OPTION_G": "",
		"OPTION_H": "\n",
		"OPTION_I": "echo 'asd'",
		"OPTION_J": `first line
second line
third line
and so on`,
		"OPTION_K": "Test#123",
		"OPTION_Z": "last value",
	}

	testReadEnvAndCompare(t, envFileName, expectedValues)
}

func TestLoadExportedEnv(t *testing.T) {
	envFileName := "fixtures/exported.env"
	expectedValues := map[string]string{
		"OPTION_A": "2",
		"OPTION_B": "\\n",
	}

	dotenv := dotenv.New()
	dotenv.SetConfigFile(envFileName)
	err := dotenv.Load()
	if err != nil {
		t.Fatal(err)
	}

	for key, value := range expectedValues {
		if dotenv.Get(key) != value {
			t.Errorf("Expected: %q got %q", value, dotenv.Get(key))
		}
	}
}

func TestErrorParsing(t *testing.T) {
	envFileName := "fixtures/invalid.env"
	dotenv := dotenv.New()
	dotenv.SetConfigFile(envFileName)
	err := dotenv.Load()
	assert.ErrorContains(t, err, "line 7: key cannot contain spaces")
}

func TestUnMarshal(t *testing.T) {
	type DB struct {
		Host     string `env:"DB_HOST" default:"localhost"`
		Port     int    `env:"DB_PORT"`
		User     string `env:"DB_USERNAME"`
		Password string `env:"DB_PASSWORD"`
		Database string `env:"DB_DATABASE"`
		Driver   string `env:"DB_DRIVER"`
	}

	type Log struct {
		Level   string `env:"LOG_LEVEL" default:"info"`
		Channel string `env:"LOG_CHANNEL" default:"stdout"`
		Path    string `env:"LOG_PATH" default:"/var/log/app.log"`
	}

	type Config struct {
		APIEndpoint  string `env:"API_ENDPOINT" default:"http://localhost:8080"`
		AuthEndpoint string `env:"AUTH_ENDPOINT" default:"http://localhost:8080"`

		DoesNotExit  string        `env:"DOES_NOT_EXIT" default:"default"`
		SomeDuration time.Duration `env:"SOME_DURATION" default:"1s"`

		DB  DB
		Log Log
	}

	expectedConfig := Config{
		APIEndpoint:  "http://localhost:8000/api",
		AuthEndpoint: "http://localhost:8000/auth",
		DoesNotExit:  "default",
		SomeDuration: time.Second,
		DB: DB{
			Host:     "localhost",
			Port:     3306,
			User:     "root",
			Password: "my-secret-pw",
			Database: "app",
			Driver:   "mysql",
		},
		Log: Log{
			Level:   "debug",
			Channel: "stack",
			Path:    "storage/logs/app.log",
		},
	}
	config := Config{}

	dotenv := dotenv.New()
	dotenv.SetConfigFile("fixtures/test.env")
	err := dotenv.Load()
	if err != nil {
		t.Fatal(err)
	}

	dotenv.SetPrefix("APP")

	err = dotenv.Unmarshal(&config)
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, expectedConfig, config)
}

type customDuration struct {
	value time.Duration
}

// check that it implements encoding.TextUnmarshaler
var _ encoding.TextUnmarshaler = (*customDuration)(nil)

func (c *customDuration) UnmarshalText(text []byte) error {
	d, err := time.ParseDuration(string(text))
	if err != nil {
		return err
	}
	log.Println(d)
	c.value = d
	return nil
}

func TestUnMarshal_fieldWithTextUnmarshaler(t *testing.T) {
	type config struct {
		Interval customDuration `env:"INTERVAL" default:"15m"`
	}

	expectedConfig := config{
		Interval: customDuration{value: 15 * time.Minute},
	}
	cfg := config{}

	dotenv := dotenv.New()
	err := dotenv.Unmarshal(&cfg)
	require.NoError(t, err)
	require.Equal(t, expectedConfig, cfg)
}

func TestGet_NoConfigFile(t *testing.T) {
	env := dotenv.New()
	val := env.GetString("SOME_KEY")
	assert.Equal(t, "", val)

	// set os env
	err := os.Setenv("SOME_KEY", "some value")
	require.NoError(t, err)

	val = env.GetString("SOME_KEY")
	assert.Equal(t, "some value", val)

	// use global instance
	val = dotenv.GetString("SOME_KEY")
	assert.Equal(t, "some value", val)
}
