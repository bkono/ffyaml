package ffyaml_test

import (
	"flag"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/bkono/ffyaml"
	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/fftest"
)

func TestParserWithNested(t *testing.T) {
	t.Parallel()

	type fields struct {
		String        string
		Float         float64
		HyphenatedKey string
		Strings       fftest.StringSlice
	}

	expected := fields{
		String:        "a string",
		Float:         1.23,
		HyphenatedKey: "valid",
		Strings:       fftest.StringSlice{"one", "two", "three"},
	}

	for _, tc := range []struct {
		name       string
		opts       []ffyaml.Option
		stringKey  string
		floatKey   string
		stringsKey string
	}{
		{
			name:       "default '.' delimiter",
			stringKey:  "string.key",
			floatKey:   "float.nested.key",
			stringsKey: "strings.nested.key",
		},
		{
			name:       "with '-' delimiter",
			opts:       []ffyaml.Option{ffyaml.WithDelimiter("-")},
			stringKey:  "string-key",
			floatKey:   "float-nested-key",
			stringsKey: "strings-nested-key",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var (
				found fields
				fs    = flag.NewFlagSet("fftest", flag.ContinueOnError)
			)

			fs.StringVar(&found.HyphenatedKey, "hyphenated-key", "", "a hyphenated yaml key")
			fs.StringVar(&found.String, tc.stringKey, "", "string")
			fs.Float64Var(&found.Float, tc.floatKey, 0, "float64")
			fs.Var(&found.Strings, tc.stringsKey, "string slice")

			err := ff.Parse(fs, []string{}, ff.WithConfigFile("testdata/nested.yaml"),
				ff.WithConfigFileParser(ffyaml.New(tc.opts...).Parse),
			)

			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(expected, found) {
				t.Errorf(`expected %v to be %v`, found, expected)
			}
		})
	}
}

func TestParser(t *testing.T) {
	t.Parallel()

	for _, testcase := range []struct {
		vars func(*flag.FlagSet) *fftest.Vars
		name string
		file string
		miss bool // AllowMissingConfigFiles
		want fftest.Vars
	}{
		{
			name: "empty",
			file: "testdata/empty.yaml",
			want: fftest.Vars{},
		},
		{
			name: "basic KV pairs",
			file: "testdata/basic.yaml",
			want: fftest.Vars{S: "hello", I: 10, B: true, D: 5 * time.Second, F: 3.14},
		},
		{
			name: "invalid prefix",
			file: "testdata/invalid_prefix.yaml",
			want: fftest.Vars{WantParseErrorString: "found character that cannot start any token"},
		},
		{
			vars: fftest.NonzeroDefaultVars,
			name: "no value for s",
			file: "testdata/no_value_s.yaml",
			want: fftest.Vars{S: "", I: 123, F: 9.99, B: true, D: 3 * time.Hour},
		},
		{
			vars: fftest.NonzeroDefaultVars,
			name: "no value for i",
			file: "testdata/no_value_i.yaml",
			want: fftest.Vars{WantParseErrorString: "parse error"},
		},
		{
			name: "basic arrays",
			file: "testdata/basic_array.yaml",
			want: fftest.Vars{S: "c", X: []string{"a", "b", "c"}},
		},
		{
			name: "multiline arrays",
			file: "testdata/multi_line_array.yaml",
			want: fftest.Vars{S: "c", X: []string{"d", "e", "f"}},
		},
		{
			name: "line break arrays",
			file: "testdata/line_break_array.yaml",
			want: fftest.Vars{X: []string{"first string", "second string", "third"}},
		},
		{
			name: "unquoted strings in arrays",
			file: "testdata/unquoted_string_array.yaml",
			want: fftest.Vars{X: []string{"one", "two", "three"}},
		},
		{
			name: "missing config file allowed",
			file: "testdata/this_file_does_not_exist.yaml",
			miss: true,
			want: fftest.Vars{},
		},
		{
			name: "missing config file not allowed",
			file: "testdata/this_file_does_not_exist.yaml",
			miss: false,
			want: fftest.Vars{WantParseErrorIs: os.ErrNotExist},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			if testcase.vars == nil {
				testcase.vars = fftest.DefaultVars
			}
			fs := flag.NewFlagSet("fftest", flag.ContinueOnError)
			vars := testcase.vars(fs)
			vars.ParseError = ff.Parse(fs, []string{},
				ff.WithConfigFile(testcase.file),
				ff.WithConfigFileParser(ffyaml.Parser),
				ff.WithAllowMissingConfigFile(testcase.miss),
			)
			fftest.Compare(t, &testcase.want, vars)
		})
	}
}
