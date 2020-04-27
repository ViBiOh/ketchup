package model

import "testing"

func TestNewSuccessMessage(t *testing.T) {
	type args struct {
		content string
	}

	var cases = []struct {
		intention string
		args      args
		want      Message
	}{
		{
			"simple",
			args{
				content: "Created!",
			},
			Message{
				Level:   "success",
				Content: "Created!",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := NewSuccessMessage(tc.args.content); got != tc.want {
				t.Errorf("NewSuccessMessage() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestNewErrorMessage(t *testing.T) {
	type args struct {
		content string
	}

	var cases = []struct {
		intention string
		args      args
		want      Message
	}{
		{
			"simple",
			args{
				content: "Failed!",
			},
			Message{
				Level:   "error",
				Content: "Failed!",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := NewErrorMessage(tc.args.content); got != tc.want {
				t.Errorf("NewErrorMessage() = %v, want %v", got, tc.want)
			}
		})
	}
}
