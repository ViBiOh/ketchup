package store

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ViBiOh/httputils/v3/pkg/db"
)

func TestStartAtomic(t *testing.T) {
	mockDb, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("unable to create mock database: %s", err)
	}
	defer mockDb.Close()

	mock.ExpectBegin()
	tx, err := mockDb.Begin()
	if err != nil {
		t.Errorf("unable to create tx: %v", err)
	}

	type args struct {
		ctx context.Context
	}

	var cases = []struct {
		intention string
		args      args
		want      bool
		wantErr   error
	}{
		{
			"present",
			args{
				ctx: db.StoreTx(context.Background(), tx),
			},
			true,
			nil,
		},
		{
			"error",
			args{
				ctx: context.Background(),
			},
			false,
			errors.New("call to database transaction Begin was not expected"),
		},
		{
			"create",
			args{
				ctx: context.Background(),
			},
			true,
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if tc.intention == "create" {
				mock.ExpectBegin()
			}

			got, gotErr := StartAtomic(tc.args.ctx, mockDb)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			} else if (db.ReadTx(got) != nil) != tc.want {
				failed = true
			}

			if failed {
				t.Errorf("StartAtomic() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
		})
	}
}

func TestEndAtomic(t *testing.T) {
	mockDb, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("unable to create mock database: %s", err)
	}
	defer mockDb.Close()

	mock.ExpectBegin()
	presentTx, err := mockDb.Begin()
	if err != nil {
		t.Errorf("unable to create tx: %v", err)
	}

	mock.ExpectBegin()
	errorTx, err := mockDb.Begin()
	if err != nil {
		t.Errorf("unable to create tx: %v", err)
	}

	type args struct {
		ctx context.Context
		err error
	}

	var cases = []struct {
		intention string
		args      args
		wantErr   error
	}{
		{
			"absent",
			args{
				ctx: context.Background(),
			},
			nil,
		},
		{
			"present",
			args{
				ctx: db.StoreTx(context.Background(), presentTx),
			},
			nil,
		},
		{
			"error",
			args{
				ctx: db.StoreTx(context.Background(), errorTx),
				err: errors.New("invalid"),
			},
			errors.New("invalid"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if tc.intention == "present" {
				mock.ExpectCommit()
			}

			gotErr := EndAtomic(tc.args.ctx, tc.args.err)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			}

			if failed {
				t.Errorf("EndAtomic() = `%s`, want `%s`", gotErr, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
		})
	}
}
