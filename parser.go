package dotenv

import (
	"os"
	"regexp"
	"strings"
	"unicode"

	"github.com/spf13/cast"
)

var (
	singleQuotesRegex  = regexp.MustCompile(`\A'(.*)'\z`)
	doubleQuotesRegex  = regexp.MustCompile(`\A"(.*)"\z`)
	escapeRegex        = regexp.MustCompile(`\\.`)
	unescapeCharsRegex = regexp.MustCompile(`\\([^$])`)
)

func readAndParseConfig(filePath, separator string) (map[string]interface{}, error) {
	var config = make(map[string]interface{})
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	temp := strings.Split(string(data), "\n")
	for _, line := range temp {
		parseLine(line, separator, config)
	}
	return config, nil
}

func parseLine(line, separator string, envMap map[string]interface{}) {
	line = strings.TrimSpace(line)
	// Skip empty lines and comments
	if line == "" || line[0] == '#' {
		return
	}
	key, val, ok := strings.Cut(line, separator)
	if ok {
		val = parseValue(val)
		if strings.HasPrefix(key, "export ") {
			_ = os.Setenv(key[7:], val)
			return
		}
		envMap[strings.ToUpper(strings.TrimSpace(key))] = val
	}
}

func parseValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	// remove comments but only if the value is not quoted
	if !(strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`)) &&
		!(strings.HasPrefix(value, `'`) && strings.HasSuffix(value, `'`)) {
		if i := strings.Index(value, "#"); i >= 0 {
			value = value[:i]
		}
	}
	// remove leading and trailing spaces
	value = strings.TrimSpace(value)
	if len(value) > 1 {
		// check if we have quoted values or possible escape characters
		singleQuotes := singleQuotesRegex.FindStringSubmatch(value)
		doubleQuotes := doubleQuotesRegex.FindStringSubmatch(value)
		if singleQuotes != nil || doubleQuotes != nil {
			value = value[1 : len(value)-1]

			if doubleQuotes != nil {
				value = escapeRegex.ReplaceAllStringFunc(value, func(s string) string {
					c := strings.TrimPrefix(s, "\\")
					switch c {
					case "n":
						return "\n"
					case "r":
						return "\r"
					default:
						return s
					}
				})
				// unescape characters
				value = unescapeCharsRegex.ReplaceAllString(value, "$1")
			}
		}
	}
	return value
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
