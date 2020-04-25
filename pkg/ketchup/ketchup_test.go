package ketchup

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
			"Usage of simple:\n  -hour string\n    \t[ketchup] Hour of cron, 24-hour format {SIMPLE_HOUR} (default \"08:00\")\n  -timezone string\n    \t[ketchup] Timezone {SIMPLE_TIMEZONE} (default \"Europe/Paris\")\n  -to string\n    \t[ketchup] Email to send notification {SIMPLE_TO}\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			fs := flag.NewFlagSet(testCase.intention, flag.ContinueOnError)
			Flags(fs, "")

			var writer strings.Builder
			fs.SetOutput(&writer)
			fs.Usage()

			result := writer.String()

			if result != testCase.want {
				t.Errorf("Flags() = `%s`, want `%s`", result, testCase.want)
			}
		})
	}
}
