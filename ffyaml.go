package ffyaml

import (
	"fmt"
	"io"
	"strconv"

	"github.com/peterbourgon/ff/v3"
	"gopkg.in/yaml.v2"
)

func Parser(r io.Reader, set func(name, value string) error) error {
	var m map[string]interface{}
	d := yaml.NewDecoder(r)
	if err := d.Decode(&m); err != nil && err != io.EOF {
		return ParseError{err}
	}

	for key, val := range m {
		fmt.Println("parsing")
		err := parseVals(val, key, set)
		if err != nil {
			return ParseError{err}
		}
	}

	return nil
}

func parseVals(val interface{}, key string, set func(name, value string) error) error {
	name := key
	if m, ok := val.(map[interface{}]interface{}); ok {
		fmt.Println("its a map[interface]interface{}")
		for k, v := range m {
			n, err := valToStr(k)
			if err != nil {
				return err
			}
			name = key + "." + n
			fmt.Println("making key: " + name)
			if err := parseVals(v, name, set); err != nil {
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

// Error implenents the error interface.
func (e ParseError) Error() string {
	return fmt.Sprintf("error parsing YAML config: %v", e.Inner)
}
