package ffyaml

import (
	"fmt"
	"io"
	"strconv"

	"github.com/peterbourgon/ff/v3"
	"gopkg.in/yaml.v2"
)

// ConfigParser supports parsing YAML files for github.com/peterbourgon/ff/v3. Key difference is support for nested yaml
// entries in the same vein as the fftoml package.
type ConfigParser struct {
	delimiter string
}

// Option is a func for configuring the ConfigParser
type Option func(*ConfigParser)

// New returns a ConfigParser, after applying the given options.
func New(opts ...Option) ConfigParser {
	c := ConfigParser{delimiter: "."}
	for _, o := range opts {
		o(&c)
	}

	return c
}

// WithDelimiter is an option which configures a delimiter
// used to construct nested YAML keys with their associated flag name.
// The default delimiter is "."
//
// For example, given the following YAML
//
//     section:
//       subsection:
//         value: 10
//
// Parse will match to a flag with the name `-section.subsection.value` by default.
// If the delimiter is "-", Parse will match to `-section-subsection-value` instead.
func WithDelimiter(d string) Option {
	return func(c *ConfigParser) {
		c.delimiter = d
	}
}

// Parser is a default parser func for ff config files. Use New, and pass the returned config's .Parse function if
// choosing to use set options.
func Parser(r io.Reader, set func(name, value string) error) error {
	return New().Parse(r, set)
}

// Parse parses config from the given io.Reader, handling nested values and assigning based on the delimiter given.
func (c ConfigParser) Parse(r io.Reader, set func(name, value string) error) error {
	var m map[string]interface{}
	d := yaml.NewDecoder(r)
	if err := d.Decode(&m); err != nil && err != io.EOF {
		return ParseError{err}
	}

	for key, val := range m {
		err := parseVals(val, key, c.delimiter, set)
		if err != nil {
			return ParseError{err}
		}
	}

	return nil
}

func parseVals(val interface{}, key, delimiter string, set func(name, value string) error) error {
	name := key
	if m, ok := val.(map[interface{}]interface{}); ok {
		for k, v := range m {
			n, err := valToStr(k)
			if err != nil {
				return err
			}
			name = key + delimiter + n
			if err := parseVals(v, name, delimiter, set); err != nil {
				return err
			}
		}
		return nil
	}

	if vals, ok := val.([]interface{}); ok {
		for i := range vals {
			s, err := valToStr(vals[i])
			if err != nil {
				return err
			}
			if err := set(name, s); err != nil {
				return err
			}
		}
		return nil
	}

	s, err := valToStr(val)
	if err != nil {
		return err
	}
	return set(name, s)
}

func valToStr(val interface{}) (string, error) {
	switch v := val.(type) {
	case byte:
		return string([]byte{v}), nil
	case string:
		return v, nil
	case bool:
		return strconv.FormatBool(v), nil
	case uint64:
		return strconv.FormatUint(v, 10), nil
	case int:
		return strconv.Itoa(v), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case float64:
		return strconv.FormatFloat(v, 'g', -1, 64), nil
	case nil:
		return "", nil
	default:
		return "", ff.StringConversionError{Value: val}
	}
}

// ParseError wraps all errors originating from the YAML parser.
type ParseError struct {
	Inner error
}

// Error implements the error interface.
func (e ParseError) Error() string {
	return fmt.Sprintf("error parsing YAML config: %v", e.Inner)
}
