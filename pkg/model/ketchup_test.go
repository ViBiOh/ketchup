package model

import (
	"reflect"
	"sort"
	"testing"

	"github.com/ViBiOh/ketchup/pkg/github"
)

func TestKetchupByRepositoryID(t *testing.T) {
	type args struct {
		array []Ketchup
	}

	var cases = []struct {
		intention string
		args      args
		want      []Ketchup
	}{
		{
			"simple",
			args{
				array: []Ketchup{
					{Repository: Repository{ID: 10}},
					{Repository: Repository{ID: 1}},
				},
			},
			[]Ketchup{
				{Repository: Repository{ID: 1}},
				{Repository: Repository{ID: 10}},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			sort.Sort(KetchupByRepositoryID(tc.args.array))
			if got := tc.args.array; !reflect.DeepEqual(got, tc.want) {
				t.Errorf("KetchupByRepositoryID() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestReleaseByRepositoryID(t *testing.T) {
	type args struct {
		array []Release
	}

	var cases = []struct {
		intention string
		args      args
		want      []Release
	}{
		{
			"simple",
			args{
				array: []Release{
					NewRelease(Repository{ID: 10}, github.Release{}),
					NewRelease(Repository{ID: 1}, github.Release{}),
				},
			},
			[]Release{
				NewRelease(Repository{ID: 1}, github.Release{}),
				NewRelease(Repository{ID: 10}, github.Release{}),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			sort.Sort(ReleaseByRepositoryID(tc.args.array))
			if got := tc.args.array; !reflect.DeepEqual(got, tc.want) {
				t.Errorf("ReleaseByRepositoryID() = %+v, want %+v", got, tc.want)
			}
		})
	}
}
