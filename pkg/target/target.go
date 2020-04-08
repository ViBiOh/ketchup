package target

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/crud"
)

var _ crud.Service = &app{}

// App of package
type App interface {
	Unmarshal(data []byte) (interface{}, error)
	Check(ctx context.Context, old, new interface{}) []crud.Error
	List(ctx context.Context, page, pageSize uint, sortKey string, sortDesc bool, filters map[string][]string) ([]interface{}, uint, error)
	Get(ctx context.Context, ID uint64) (interface{}, error)
	Create(ctx context.Context, o interface{}) (interface{}, error)
	Update(ctx context.Context, o interface{}) (interface{}, error)
	Delete(ctx context.Context, o interface{}) error
}

type app struct {
	db *sql.DB
}

// New creates new App from Config
func New(db *sql.DB) App {
	return &app{
		db: db,
	}
}

// Unmarshal a Target
func (a app) Unmarshal(content []byte) (interface{}, error) {
	var o Target

	if err := json.Unmarshal(content, &o); err != nil {
		return nil, err
	}

	return o, nil
}

// List Targets
func (a app) List(ctx context.Context, page, pageSize uint, sortKey string, sortAsc bool, _ map[string][]string) ([]interface{}, uint, error) {
	list, total, err := a.listTargets(page, pageSize, sortKey, sortAsc)
	if err != nil {
		return nil, 0, fmt.Errorf("unable to list targets: %w", err)
	}

	itemsList := make([]interface{}, len(list))
	for index, item := range list {
		itemsList[index] = item
	}

	return itemsList, total, nil
}

// Get Target by ID
func (a app) Get(ctx context.Context, ID uint64) (interface{}, error) {
	target, err := a.getTargetByID(ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, crud.ErrNotFound
		}

		return nil, fmt.Errorf("unable to get target: %w", err)
	}

	return target, nil
}

// Create Target
func (a app) Create(ctx context.Context, o interface{}) (item interface{}, err error) {
	var target Target
	target, err = getTargetFromItem(ctx, o)
	if err != nil {
		return
	}

	err = a.saveTarget(&target, nil)
	if err != nil {
		err = fmt.Errorf("unable to create target: %w", err)
		return
	}

	item = target

	return
}

// Update Target
func (a app) Update(ctx context.Context, o interface{}) (item interface{}, err error) {
	var target Target
	target, err = getTargetFromItem(ctx, o)
	if err != nil {
		return
	}

	err = a.saveTarget(&target, nil)
	if err != nil {
		err = fmt.Errorf("unable to update target: %w", err)

		return
	}

	item = target

	return
}

// Delete Target
func (a app) Delete(ctx context.Context, o interface{}) (err error) {
	var target Target
	target, err = getTargetFromItem(ctx, o)
	if err != nil {
		return
	}

	err = a.deleteTarget(target, nil)
	if err != nil {
		err = fmt.Errorf("unable to delete target: %w", err)
	}

	return
}

// Check instance
func (a app) Check(ctx context.Context, old, new interface{}) []crud.Error {
	if old != nil && new == nil {
		return nil
	}

	item, err := getTargetFromItem(ctx, new)
	if err != nil {
		return []crud.Error{crud.NewError("item", err.Error())}
	}

	errors := make([]crud.Error, 0)

	if strings.TrimSpace(item.Repository) == "" {
		errors = append(errors, crud.NewError("repository", "repository is required, in the form user/repository"))
	}

	if strings.TrimSpace(item.CurrentVersion) == "" {
		errors = append(errors, crud.NewError("current_version", "current version is required"))
	}

	return errors
}

func getTargetFromItem(ctx context.Context, o interface{}) (Target, error) {
	item, ok := o.(Target)
	if !ok {
		return item, errors.New("item is not a Target")
	}

	return item, nil
}
