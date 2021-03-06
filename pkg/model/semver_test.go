package model

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/ViBiOh/ketchup/pkg/semver"
)

func safeParse(version string) semver.Version {
	output, err := semver.Parse(version)
	if err != nil {
		fmt.Println(err)
	}
	return output
}

func safeParsePattern(pattern string) semver.Pattern {
	output, err := semver.ParsePattern(pattern)
	if err != nil {
		fmt.Println(err)
	}
	return output
}

func TestCheckPatternsMatching(t *testing.T) {
	type args struct {
		versions         map[string]semver.Version
		compiledPatterns map[string]semver.Pattern
		version          semver.Version
	}

	var cases = []struct {
		intention string
		args      args
		want      map[string]semver.Version
	}{
		{
			"no match",
			args{
				versions: map[string]semver.Version{},
				compiledPatterns: map[string]semver.Pattern{
					"stable": safeParsePattern("stable"),
				},
				version: safeParse("1.2.3"),
			},
			map[string]semver.Version{},
		},
		{
			"greater",
			args{
				versions: map[string]semver.Version{
					"stable": semver.NoneVersion,
				},
				compiledPatterns: map[string]semver.Pattern{
					"stable": safeParsePattern("stable"),
				},
				version: safeParse("1.2.3"),
			},
			map[string]semver.Version{
				"stable": safeParse("1.2.3"),
			},
		},
		{
			"mutiple match",
			args{
				versions: map[string]semver.Version{
					"^2.0":   semver.NoneVersion,
					"stable": semver.NoneVersion,
					"~1.2":   semver.NoneVersion,
				},
				compiledPatterns: map[string]semver.Pattern{
					"stable": safeParsePattern("stable"),
					"^2.0":   safeParsePattern("^2.0"),
					"~1.0":   safeParsePattern("~1.0"),
				},
				version: safeParse("1.2.3"),
			},
			map[string]semver.Version{
				"^2.0":   semver.NoneVersion,
				"stable": safeParse("1.2.3"),
				"~1.2":   safeParse("1.2.3"),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			CheckPatternsMatching(tc.args.versions, tc.args.compiledPatterns, tc.args.version)

			if !reflect.DeepEqual(tc.args.versions, tc.want) {
				t.Errorf("CheckPatternsMatching() = %+v, want %+v", tc.args.versions, tc.want)
			}
		})
	}
}
