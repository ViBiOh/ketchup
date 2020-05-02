package ketchup

import (
	"testing"

	"github.com/ViBiOh/ketchup/pkg/model"
)

func TestComputeSemver(t *testing.T) {
	type args struct {
		item model.Ketchup
	}

	var cases = []struct {
		intention string
		args      args
		want      string
	}{
		{
			"version match",
			args{
				item: model.Ketchup{
					Version: "1.0.0",
					Repository: model.Repository{
						Version: "1.0.0",
					},
				},
			},
			"",
		},
		{
			"no repo semver",
			args{
				item: model.Ketchup{
					Version: "1.0.0",
					Repository: model.Repository{
						Version: "1.0",
					},
				},
			},
			"",
		},
		{
			"no semver",
			args{
				item: model.Ketchup{
					Version: "1.0",
					Repository: model.Repository{
						Version: "1.0.0",
					},
				},
			},
			"",
		},
		{
			"major",
			args{
				item: model.Ketchup{
					Version: "1.0.0",
					Repository: model.Repository{
						Version: "2.1.3",
					},
				},
			},
			"Major",
		},
		{
			"minor",
			args{
				item: model.Ketchup{
					Version: "2.0.3",
					Repository: model.Repository{
						Version: "2.1.3",
					},
				},
			},
			"Minor",
		},
		{
			"patch",
			args{
				item: model.Ketchup{
					Version: "2.1.2",
					Repository: model.Repository{
						Version: "2.1.3",
					},
				},
			},
			"Patch",
		},
		{
			"complex semver",
			args{
				item: model.Ketchup{
					Version: "2.1.2+abcdef",
					Repository: model.Repository{
						Version: "v2.1.3",
					},
				},
			},
			"Patch",
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := computeSemver(tc.args.item); got != tc.want {
				t.Errorf("computeSemver() = %s, want %s", got, tc.want)
			}
		})
	}
}
