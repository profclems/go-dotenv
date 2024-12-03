package dotenv

import (
	"bytes"
	"encoding"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/spf13/cast"
)

const (
	// DefaultConfigFile is the default name of the configuration file.
	DefaultConfigFile = ".env"
)

// DotEnv is a prioritized .env configuration registry.
// It maintains a set of configuration sources, fetches
// values to populate those, and provides them according
// to the source's priority.
// The priority of the sources is the following:
// 1. env. variables
// 2. key/value cache/store (loaded from config file or set explicitly with Set())
// 3. defaults(when using structures)
//
// For example, if values from the following sources were loaded:
//
//	Defaults
//		USER=default
//		ENDPOINT=https://localhost
//
//	Config
//		USER=root
//		SECRET=secretFromConfig
//
//	Environment
//		SECRET=secretFromEnv
//
// The resulting config will have the following values:
//
//	SECRET=secretFromEnv
//	USER=root
//	ENDPOINT=https://localhost
//
// DotEnv is safe for concurrent Get___() and Set() operations by multiple goroutines.
type DotEnv struct {
	decoder Decoder

	configFile        string
	prefix            string
	allowEmptyEnvVars bool

	mu           sync.RWMutex
	cachedConfig map[string]any
}

// global DotEnv instance
var (
	_globalMu sync.RWMutex
	_global   = New()
)

// ReplaceDefault replaces the default DotEnv instance with a new one
// and returns a function to restore the previous instance.
// This is useful for customizing the default DotEnv instance.
// It's safe for concurrent use.
func ReplaceDefault(env *DotEnv) func() {
	_globalMu.Lock()
	prev := _global
	_global = env
	_globalMu.Unlock()
	return func() {
		ReplaceDefault(prev)
	}
}

// New returns an initialized DotEnv instance.
// This does not load the config file. You call Load() to do that.
func New() *DotEnv {
	return &DotEnv{
		decoder:    &DefaultDecoder{},
		configFile: DefaultConfigFile,
	}
}

var utf8BOM = []byte("\uFEFF")

// Load reads the config file(s) and loads the configuration
// in the order of the files provided.
// It returns os.ErrNotExist if config file does not exist.
// If no config file is specified, it loads the .env file from the current directory by default.
func Load(files ...string) error {
	return GetDotEnv().Load(files...)
}

func (e *DotEnv) Load(files ...string) error {
	config := make(map[string]any)
	if len(files) == 0 {
		files = []string{e.configFile}
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		data = bytes.TrimPrefix(data, utf8BOM)

		err = e.decoder.Decode(data, config)
		if err != nil {
			return err
		}
	}

	e.mu.Lock()
	if e.cachedConfig == nil {
		e.cachedConfig = make(map[string]any)
	}

	for key, val := range config {
		e.cachedConfig[key] = val
	}
	e.mu.Unlock()

	return nil
}

// LoadWithDecoder is like Load but uses the provided decoder to decode the config file(s).
func LoadWithDecoder(decoder Decoder, files ...string) error {
	return GetDotEnv().LoadWithDecoder(decoder, files...)
}

func (e *DotEnv) LoadWithDecoder(decoder Decoder, files ...string) error {
	e.decoder = decoder
	return e.Load(files...)
}

// GetDotEnv returns the global DotEnv instance which can reconfigured with ReplaceDefault.
// It's safe for concurrent use.
func GetDotEnv() *DotEnv {
	_globalMu.Lock()
	d := _global
	_globalMu.Unlock()

	return d
}

// SetPrefix defines a prefix that ENVIRONMENT variables will use.
// E.g. if your prefix is "pro", the env registry will look for env
// variables that start with "PRO_".
func SetPrefix(prefix string) { GetDotEnv().SetPrefix(prefix) }

func (e *DotEnv) SetPrefix(prefix string) {
	e.prefix = strings.ToUpper(prefix) + "_"
}

// GetPrefix returns the prefix that ENVIRONMENT variables will use which is set with SetPrefix.
func GetPrefix() string { return GetDotEnv().GetPrefix() }

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

// AllowEmptyEnv tells Dotenv to consider set, but empty environment variables
// as valid values instead of falling back to config value.
// This is set to true by default.
func AllowEmptyEnv(allowEmptyEnvVars bool) { GetDotEnv().AllowEmptyEnvVars(allowEmptyEnvVars) }

func (e *DotEnv) AllowEmptyEnvVars(allowEmptyEnvVars bool) {
	e.allowEmptyEnvVars = allowEmptyEnvVars
}

// SetConfigFile explicitly defines the path, name and extension of the config file.
// Dotenv will use this and not check .env from the current directory.
// You need to call Load() to read the config file.
// Or you could directly load the config file by calling Load("path/to/config/file").
func SetConfigFile(configFile string) {
	if _global == nil {
		_global = New()
	}
	_global.SetConfigFile(configFile)
}

func (e *DotEnv) SetConfigFile(configFile string) {
	e.configFile = configFile
}

// Unmarshal unmarshals the config file into a struct.
// Recognizes the following struct tags:
//   - env:"KEY" to specify the key name to look up in the config file
//   - default:"value" to specify a default value if the key is not found
func Unmarshal(v any) error {
	return GetDotEnv().Unmarshal(v)
}

func (e *DotEnv) Unmarshal(v any) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	vPtr := reflect.ValueOf(v)
	val := vPtr.Elem()

	if vk := val.Kind(); vk != reflect.Struct {
		return fmt.Errorf("expected a struct, got %T", vk)
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		getConfigVal := func() string {
			tag := field.Tag.Get("env")
			if tag != "" {
				if envVal := e.GetString(tag); envVal != "" {
					return envVal
				}
			}
			// set default value
			if def := field.Tag.Get("default"); def != "" {
				return def
			}
			return ""
		}

		if fieldVal.CanAddr() {
			if m, ok := fieldVal.Addr().Interface().(encoding.TextUnmarshaler); ok {
				val := getConfigVal()
				if val == "" {
					continue
				}
				if err := m.UnmarshalText([]byte(getConfigVal())); err != nil {
					return err
				}
				continue
			}
		}

		if field.Type.Kind() == reflect.Struct {
			if err := e.Unmarshal(fieldVal.Addr().Interface()); err != nil {
				return err
			}
			continue
		}

		configVal := getConfigVal()
		if configVal == "" {
			continue
		}

		// set the value based on the field type
		switch field.Type {
		case reflect.TypeOf(time.Time{}):
			fieldVal.Set(reflect.ValueOf(cast.ToTime(configVal)))
		case reflect.TypeOf(time.Duration(0)):
			fieldVal.Set(reflect.ValueOf(cast.ToDuration(configVal)))
		case reflect.TypeOf([]int{}):
			fieldVal.Set(reflect.ValueOf(cast.ToIntSlice(configVal)))
		case reflect.TypeOf([]string{}):
			fieldVal.Set(reflect.ValueOf(cast.ToStringSlice(configVal)))
		default:
			switch field.Type.Kind() {
			case reflect.String:
				fieldVal.SetString(cast.ToString(configVal))
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				fieldVal.SetInt(cast.ToInt64(configVal))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				fieldVal.SetUint(cast.ToUint64(configVal))
			case reflect.Float32, reflect.Float64:
				fieldVal.SetFloat(cast.ToFloat64(configVal))
			case reflect.Bool:
				fieldVal.SetBool(cast.ToBool(configVal))
			default:
				return fmt.Errorf("unsupported type %s", field.Type)
			}
		}
	}

	return err
}

// Get can retrieve any value given the key to use.
// Get is case-insensitive for a key.
// Dotenv will check in the following order:
// configOverride cache, env, key/value store, config file
//
// Get returns an interface. For a specific value use one of the Get___ methods e.g. GetBool(key) for a boolean value
func Get(key string) any { return GetDotEnv().Get(key) }

func (e *DotEnv) Get(key string) any {
	val, _ := e.LookUp(key)
	return val
}

// GetString returns the value associated with the key as a string.
func GetString(key string) string { return GetDotEnv().GetString(key) }

func (e *DotEnv) GetString(key string) string {
	return cast.ToString(e.Get(key))
}

// GetBool returns the value associated with the key as a boolean.
func GetBool(key string) bool { return GetDotEnv().GetBool(key) }

func (e *DotEnv) GetBool(key string) bool {
	return cast.ToBool(e.Get(key))
}

// GetInt returns the value associated with the key as an integer.
func GetInt(key string) int { return GetDotEnv().GetInt(key) }

func (e *DotEnv) GetInt(key string) int {
	return cast.ToInt(e.Get(key))
}

// GetInt32 returns the value associated with the key as an integer.
func GetInt32(key string) int32 { return GetDotEnv().GetInt32(key) }

func (e *DotEnv) GetInt32(key string) int32 {
	return cast.ToInt32(e.Get(key))
}

// GetInt64 returns the value associated with the key as an integer.
func GetInt64(key string) int64 { return GetDotEnv().GetInt64(key) }

func (e *DotEnv) GetInt64(key string) int64 {
	return cast.ToInt64(e.Get(key))
}

// GetUint returns the value associated with the key as an unsigned integer.
func GetUint(key string) uint { return GetDotEnv().GetUint(key) }

func (e *DotEnv) GetUint(key string) uint {
	return cast.ToUint(e.Get(key))
}

// GetUint32 returns the value associated with the key as an unsigned integer.
func GetUint32(key string) uint32 { return GetDotEnv().GetUint32(key) }

func (e *DotEnv) GetUint32(key string) uint32 {
	return cast.ToUint32(e.Get(key))
}

// GetUint64 returns the value associated with the key as an unsigned integer.
func GetUint64(key string) uint64 { return GetDotEnv().GetUint64(key) }

func (e *DotEnv) GetUint64(key string) uint64 {
	return cast.ToUint64(e.Get(key))
}

// GetFloat64 returns the value associated with the key as a float64.
func GetFloat64(key string) float64 { return GetDotEnv().GetFloat64(key) }

func (e *DotEnv) GetFloat64(key string) float64 {
	return cast.ToFloat64(e.Get(key))
}

// GetTime returns the value associated with the key as time.
func GetTime(key string) time.Time { return GetDotEnv().GetTime(key) }

func (e *DotEnv) GetTime(key string) time.Time {
	return cast.ToTime(e.Get(key))
}

// GetDuration returns the value associated with the key as a duration.
func GetDuration(key string) time.Duration { return GetDotEnv().GetDuration(key) }

func (e *DotEnv) GetDuration(key string) time.Duration {
	return cast.ToDuration(e.Get(key))
}

// GetIntSlice returns the value associated with the key as a slice of int values.
func GetIntSlice(key string) []int { return GetDotEnv().GetIntSlice(key) }

func (e *DotEnv) GetIntSlice(key string) []int {
	return cast.ToIntSlice(toSlice(e.GetString(key)))
}
func toSlice(value string) []string {
	value = strings.TrimPrefix(value, "[")
	value = strings.TrimSuffix(value, "]")
	return strings.Split(value, ",")
}

// GetStringSlice returns the value associated with the key as a slice of strings.
func GetStringSlice(key string) []string { return GetDotEnv().GetStringSlice(key) }

func (e *DotEnv) GetStringSlice(key string) []string {
	return cast.ToStringSlice(toSlice(e.GetString(key)))
}

// GetSizeInBytes returns the size of the value associated with the given key
// in bytes.
func GetSizeInBytes(key string) uint { return GetDotEnv().GetSizeInBytes(key) }

func (e *DotEnv) GetSizeInBytes(key string) uint {
	sizeStr := cast.ToString(e.Get(key))
	return parseSizeInBytes(sizeStr)
}

// IsSet checks to see if the key has been set in any of the env var, config cache or config file.
// IsSet is case-insensitive for a key.
func IsSet(key string) bool { return GetDotEnv().IsSet(key) }

func (e *DotEnv) IsSet(key string) bool {
	_, set := e.LookUp(key)
	return set
}

// LookUp retrieves the value of the configuration named by the key.
// If the variable is set (which may be empty) is returned and the boolean is true.
// Otherwise the returned value will be empty and the boolean will be false.
func LookUp(key string) (any, bool) { return GetDotEnv().LookUp(key) }

func (e *DotEnv) LookUp(key string) (any, bool) {
	if key != "" {
		key = strings.ToUpper(e.addPrefix(key))

		if val, ok := os.LookupEnv(key); ok {
			if val != "" && !e.allowEmptyEnvVars {
				return val, true
			}
		}

		e.mu.Lock()
		defer e.mu.Unlock()

		if cachedEnv, okEnv := e.cachedConfig[key]; okEnv {
			return cachedEnv, true
		}
	}
	return nil, false
}

// Set sets or update env variable
// This will be used instead of following the normal precedence
// when getting the value
func Set(key string, value any) { GetDotEnv().Set(key, value) }

func (e *DotEnv) Set(key string, value any) {
	key = e.addPrefix(key)
	key = strings.ToUpper(key)

	e.mu.Lock()
	e.cachedConfig[key] = value
	e.mu.Unlock()
}

// Deprecated: to be removed in v2.0.0
//
// Save writes the current configuration to a file.
func Save() error { return GetDotEnv().Save() }

// Deprecated: to be removed in v2.0.0
//
// Save writes the current configuration to a file.
func (e *DotEnv) Save() error {
	cfgData := ""

	e.mu.RLock()
	for key, value := range e.cachedConfig {
		cfgData += fmt.Sprintf("%s=%s\n", key, cast.ToString(value))
	}
	e.mu.RUnlock()

	return writeConfig(e.configFile, cfgData)
}

// Write explicitly sets/update the configuration with the key-value provided
// and writes the current configuration to a file.
// This is the same as
//
//	dotenv.Set(key, value)
//	dotenv.Save()
func Write(key string, value any) error { return GetDotEnv().Write(key, value) }

func (e *DotEnv) Write(key string, value any) error {
	e.Set(key, value)
	return e.Save()
}

func writeConfig(cfgFile, data string) error {
	_ = os.MkdirAll(filepath.Join(cfgFile, ".."), 0755)
	if err := os.WriteFile(cfgFile, []byte(data), 0666); err != nil {
		return fmt.Errorf("failed to write to config file: %q", err)
	}

	return nil
}

func safeMul(a, b uint) uint {
	c := a * b
	if a > 1 && b > 1 && c/b != a {
		return 0
	}
	return c
}

// parseSizeInBytes converts strings like 1GB or 12 mb into an unsigned integer number of bytes
func parseSizeInBytes(sizeStr string) uint {
	sizeStr = strings.TrimSpace(sizeStr)
	lastChar := len(sizeStr) - 1
	multiplier := uint(1)

	if lastChar > 0 {
		if sizeStr[lastChar] == 'b' || sizeStr[lastChar] == 'B' {
			if lastChar > 1 {
				switch unicode.ToLower(rune(sizeStr[lastChar-1])) {
				case 'k':
					multiplier = 1 << 10
					sizeStr = strings.TrimSpace(sizeStr[:lastChar-1])
				case 'm':
					multiplier = 1 << 20
					sizeStr = strings.TrimSpace(sizeStr[:lastChar-1])
				case 'g':
					multiplier = 1 << 30
					sizeStr = strings.TrimSpace(sizeStr[:lastChar-1])
				default:
					multiplier = 1
					sizeStr = strings.TrimSpace(sizeStr[:lastChar])
				}
			}
		}
	}

	size := cast.ToInt(sizeStr)
	if size < 0 {
		size = 0
	}

	return safeMul(uint(size), multiplier)
}
