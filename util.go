package dotenv

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/spf13/cast"
)

func readConfig(filePath, separator string) (map[string]interface{}, error) {
	var config = make(map[string]interface{})
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	file := string(data)
	temp := strings.Split(file, "\n")
	for _, item := range temp {
		env := strings.SplitN(item, separator, 2)
		if len(env) > 1 {
			config[strings.ToUpper(strings.TrimSpace(env[0]))] = strings.TrimSpace(env[1])
		}
	}
	return config, nil
}

func writeConfig(cfgFile, data string) error {
	defer InvalidateEnvCacheForFile(cfgFile)

	_ = os.MkdirAll(filepath.Join(cfgFile, ".."), 0755)
	if err := WriteFile(cfgFile, []byte(data), 0666); err != nil {
		return fmt.Errorf("failed to write to config file: %q", err)
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
