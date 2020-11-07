package dotenv

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	DefaultConfigFile = ".env"
	DefaultSeparator  = "="
	// multiple config files cache: <file: <key: value>>
	cachedConfig map[string]map[string]string
)

type DotEnv struct {
	Config     map[string]string
	ConfigFile string
	Separator  string
	prefix     string
}

// Init will read your env file and cache the config in DotEnv.Config.
// Call this function as close as possible to the start of your program (ideally in main where your config file resides)
// If you call Init without any args it will default to loading .env in the current path
// You can otherwise tell it which files to load like
//
//		dotenv.Init("file.env")
func Init(file ...string) (*DotEnv, error) {
	var configFile string
	if len(file) > 0 {
		configFile = file[0]
	}

	if configFile == "" {
		configFile = DefaultConfigFile
	}

	dotenv := &DotEnv{
		ConfigFile: configFile,
		Separator:  DefaultSeparator,
	}

	if !CheckFileExists(configFile) {
		return nil, os.ErrNotExist
	}

	dotenv.Config = readConfig(configFile)
	return dotenv, nil
}

func (e *DotEnv) SetPrefix(prefix string) {
	e.prefix = prefix + "_"
}

func (e *DotEnv) GetPrefix() string {
	return strings.TrimSuffix(e.prefix, "_")
}

func (e *DotEnv) addPrefix(key string) string {
	if e.prefix != "" {
		if !strings.HasPrefix(e.prefix, key) {
			key = e.prefix + key
		}
	}
	return key
}

func (e *DotEnv) OverrideConfigFile(configFile string) {
	e.ConfigFile = configFile
	if !CheckFileExists(configFile) {
		e.Config = readConfig(configFile)
	}
}

// Get returns env variable value. It first looks for the key from the OS env var
// before searching from the config file
func (e *DotEnv) Get(key string) string {
	if key != "" {
		key = e.addPrefix(key)

		if e.Config != nil {
			return e.Config[key]
		}

		// first get os env var
		env := os.Getenv(key)

		if env == "" {
			// Find config variable in config file
			env, _, _ = GetFromFile(e.ConfigFile, key)
		}
		return env
	}

	return ""
}

// Set writes or update env variable to config file
func (e *DotEnv) Set(key, value string) error {
	defer InvalidateEnvCacheForFile(e.ConfigFile)

	key = e.addPrefix(key)

	data, err := ioutil.ReadFile(e.ConfigFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read/update env config: %q", err)
	}

	file := string(data)
	temp := strings.Split(file, "\n")
	newData := ""
	keyExists := false
	newConfig := key + e.Separator + (value) + "\n"
	for _, item := range temp {
		if item == "" {
			continue
		}

		env := strings.SplitN(item, e.Separator, 2)
		if env[0] == key {
			newData += newConfig
			keyExists = true
		} else {
			newData += item + "\n"
		}
	}
	if !keyExists {
		newData += newConfig
	}
	_ = os.MkdirAll(filepath.Join(e.ConfigFile, ".."), 0755)
	if err = WriteFile(e.ConfigFile, []byte(newData), 0666); err != nil {
		return fmt.Errorf("failed to update config file: %q", err)
	}

	return nil
}

// CheckFileExists returns true if a file exists and is not a directory.
func CheckFileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// InvalidateEnvCacheForFile invalidates the cached content of a file used by eg. GetKeyValueInFile
func InvalidateEnvCacheForFile(filePath string) {
	delete(cachedConfig, filePath)
}

// GetFromFile retrieves the value of the config variable named by the key from the config file
// If the variable is present in the environment the value (which may be empty) is returned and the boolean is true.
// Otherwise the returned value will be empty and the boolean will be false.
func GetFromFile(filePath, key string) (string, bool, error) {
	if !CheckFileExists(filePath) {
		return "", false, os.ErrNotExist
	}

	configCache, okConfig := cachedConfig[filePath]
	if !okConfig {
		configCache = readConfig(filePath)
		if cachedConfig == nil {
			cachedConfig = make(map[string]map[string]string)
		}
		cachedConfig[filePath] = configCache
	}

	if cachedEnv, okEnv := configCache[key]; okEnv {
		return cachedEnv, true, nil
	}

	return "", false, nil
}

func readConfig(filePath string) map[string]string {
	var config = make(map[string]string)
	data, _ := ioutil.ReadFile(filePath)
	file := string(data)
	temp := strings.Split(file, "\n")
	for _, item := range temp {
		env := strings.SplitN(item, "=", 2)
		if len(env) > 1 {
			config[env[0]] = env[1]
		}
	}
	return config
}
