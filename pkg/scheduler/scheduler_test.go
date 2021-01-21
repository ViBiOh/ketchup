package scheduler

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
			"Usage of simple:\n  -hour string\n    \t[scheduler] Hour of cron, 24-hour format {SIMPLE_HOUR} (default \"08:00\")\n  -loginID uint\n    \t[scheduler] Scheduler user ID {SIMPLE_LOGIN_ID} (default 1)\n  -timezone string\n    \t[scheduler] Timezone {SIMPLE_TIMEZONE} (default \"Europe/Paris\")\n",
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
