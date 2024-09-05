package dotenv

import (
	"fmt"
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

// Decoder decodes the contents of an env file into a map.
type Decoder interface {
	Decode(b []byte, v map[string]any) error
}

// DefaultDecoder is the default decoder used by the library.
type DefaultDecoder struct{}

// Decode decodes the contents of b into v.
func (d *DefaultDecoder) Decode(b []byte, v map[string]any) error {
	data := string(b)
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		err := parseLine(line, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func parseLine(line string, envMap map[string]any) error {
	line = strings.TrimSpace(line)
	// Skip empty lines and comments
	if line == "" || line[0] == '#' {
		return nil
	}

	// find the first occurrence of an equal sign or colon
	key, val, ok := strings.Cut(line, "=")
	if !ok {
		key, val, ok = strings.Cut(line, ":")
		if !ok {
			return fmt.Errorf("invalid format: %s", line)
		}
	}
	val = parseValue(val)
	if strings.HasPrefix(key, "export ") {
		_ = os.Setenv(key[7:], val)
		return nil
	}
	envMap[strings.ToUpper(strings.TrimSpace(key))] = val
	return nil
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
