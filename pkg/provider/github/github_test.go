package github

import (
	"flag"
	"strings"
	"testing"
)

func TestFlags(t *testing.T) {
	var cases = []struct {
		intention string
		want      string
	}{
		{
			"simple",
			"Usage of simple:\n  -token string\n    \t[github] OAuth Token {SIMPLE_TOKEN}\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			fs := flag.NewFlagSet(tc.intention, flag.ContinueOnError)
			Flags(fs, "")

			var writer strings.Builder
			fs.SetOutput(&writer)
			fs.Usage()

			result := writer.String()

			if result != tc.want {
				t.Errorf("Flags() = `%s`, want `%s`", result, tc.want)
			}
		})
	}
}
