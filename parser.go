package dotenv

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

const (
	prefixSingleQuote = '\''
	prefixDoubleQuote = '"'
)

var (
	escapeRegex        = regexp.MustCompile(`\\.`)
	unescapeCharsRegex = regexp.MustCompile(`\\([^$])`)
)

// Decoder decodes the contents of an env file into a map.
type Decoder interface {
	Decode(b []byte, v map[string]any) error
}

// DefaultDecoder is the default decoder used by the library.
type DefaultDecoder struct {
	line int
}

// Decode decodes the contents of b into v.
func (d *DefaultDecoder) Decode(b []byte, v map[string]any) error {
	data := string(b)
	lines := strings.Split(data, "\n")

	var curKey, curVal string
	var curQuote byte

	for _, line := range lines {
		d.line++
		if curQuote == 0 {
			// not in a quoted value block
			line = strings.TrimSpace(line)
			// Skip empty lines and comments
			if line == "" || line[0] == '#' {
				continue
			}

			// find the first occurrence of an equal sign or colon
			key, val, ok := strings.Cut(line, "=")
			if !ok {
				key, val, ok = strings.Cut(line, ":")
				// TODO: support inherited variables
			}
			key = strings.TrimSpace(key)
			if !strings.HasPrefix(key, "export ") && strings.Contains(key, " ") {
				return fmt.Errorf("line %d: key cannot contain spaces", d.line)
			}

			val = strings.TrimSpace(val)
			// check if the value is quoted
			quote, isQuoted := isPrefixQuoted(val)
			if isQuoted {
				// get the value without the quotes
				// if the value is quoted, check if it's a multi-line value
				idx := d.findTerminator(val[1:], quote)
				if idx == -1 {
					// if the value is not terminated, continue to the next line
					curKey = key
					curVal = val
					curQuote = quote
					continue
				}
			}

			val = parseValue(val)
			addEnv(key, val, v)
			continue
		}

		// in a quoted value block
		curVal += "\n" + line
		if d.findTerminator(line, curQuote) == -1 {
			continue
		}

		// value is terminated, parse and add to the environment
		curVal = parseValue(curVal)
		addEnv(curKey, curVal, v)
		curKey, curVal, curQuote = "", "", 0
	}

	if curQuote != 0 {
		return fmt.Errorf("line %d: unterminated quoted value", d.line)

	}
	return nil
}

// addEnv adds the key and value to the environment.
func addEnv(key, value string, v map[string]any) {
	if strings.HasPrefix(key, "export ") {
		_ = os.Setenv(key[7:], value)
		return
	}
	v[strings.ToUpper(key)] = value
}

// findTerminator finds the terminator of a quote in a string
// and returns the index of the terminator.
func (d *DefaultDecoder) findTerminator(str string, quote byte) int {
	previousCharIsEscape := false
	for i := 0; i < len(str); i++ {
		char := str[i]

		if char == quote {
			if !previousCharIsEscape {
				return i
			}
		}

		if !previousCharIsEscape && char == '\\' {
			previousCharIsEscape = true
			continue
		}
		if previousCharIsEscape {
			previousCharIsEscape = false
			continue
		}
	}

	return -1
}

func parseValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	// remove comments but only if the value is not quoted
	if !isQuoted(value) {
		if i := strings.Index(value, "#"); i >= 0 {
			value = value[:i]
		}
	}
	// remove leading and trailing spaces
	value = strings.TrimSpace(value)
	if len(value) > 1 {
		if quote, ok := isPrefixQuoted(value); ok {
			// remove quotes
			value = value[1 : len(value)-1]

			if quote == prefixDoubleQuote {
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

func isPrefixQuoted(s string) (byte, bool) {
	if s == "" {
		return 0, false
	}
	switch quote := s[0]; quote {
	case prefixDoubleQuote, prefixSingleQuote:
		return quote, true
	default:
		return 0, false
	}
}

func isQuoted(s string) bool {
	if len(s) < 2 {
		return false
	}

	return s[0] == s[len(s)-1] && (s[0] == prefixDoubleQuote || s[0] == prefixSingleQuote)
}
