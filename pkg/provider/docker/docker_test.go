package docker

import (
	"net/http"
	"testing"
)

func TestGetNextURL(t *testing.T) {
	type args struct {
		headers  http.Header
		registry string
	}

	upperCaseHeader := http.Header{}
	upperCaseHeader.Add("Link", `rel="next"; /v2/test`)

	lowerCaseHeader := http.Header{}
	lowerCaseHeader.Add("link", `rel="prev"; /v2/prev`)
	lowerCaseHeader.Add("link", `rel="next"; /v2/next`)

	cases := []struct {
		intention string
		args      args
		want      string
	}{
		{
			"empty",
			args{
				headers:  http.Header{},
				registry: "http://localhost",
			},
			"",
		},
		{
			"uppercase link with next",
			args{
				headers:  upperCaseHeader,
				registry: "http://localhost",
			},
			"http://localhost/v2/test",
		},
		{
			"lowercase link with previous and next inverted",
			args{
				headers:  lowerCaseHeader,
				registry: "http://localhost",
			},
			"http://localhost/v2/next",
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := getNextURL(tc.args.headers, tc.args.registry); got != tc.want {
				t.Errorf("getNextURL() =`%s`, want`%s`", got, tc.want)
			}
		})
	}
}
