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
	t.Parallel()

	type args struct {
		versions         map[string]semver.Version
		compiledPatterns map[string]semver.Pattern
		version          semver.Version
	}

	cases := map[string]struct {
		args args
		want map[string]semver.Version
	}{
		"no match": {
			args{
				versions: map[string]semver.Version{},
				compiledPatterns: map[string]semver.Pattern{
					"stable": safeParsePattern("stable"),
				},
				version: safeParse("1.2.3"),
			},
			map[string]semver.Version{},
		},
		"greater": {
			args{
				versions: map[string]semver.Version{
					"stable": {},
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
		"mutiple match": {
			args{
				versions: map[string]semver.Version{
					"^2.0":   {},
					"stable": {},
					"~1.2":   {},
				},
				compiledPatterns: map[string]semver.Pattern{
					"stable": safeParsePattern("stable"),
					"^2.0":   safeParsePattern("^2.0"),
					"~1.0":   safeParsePattern("~1.0"),
				},
				version: safeParse("1.2.3"),
			},
			map[string]semver.Version{
				"^2.0":   {},
				"stable": safeParse("1.2.3"),
				"~1.2":   safeParse("1.2.3"),
			},
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			CheckPatternsMatching(testCase.args.versions, testCase.args.compiledPatterns, testCase.args.version)

			if !reflect.DeepEqual(testCase.args.versions, testCase.want) {
				t.Errorf("CheckPatternsMatching() = %+v, want %+v", testCase.args.versions, testCase.want)
			}
		})
	}
}
