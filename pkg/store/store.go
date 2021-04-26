package store

import (
	"fmt"
	"strings"
)

// AddPagination add query pagination to the given query builder
func AddPagination(query *strings.Builder, index int, page, pageSize uint) []interface{} {
	query.WriteString(fmt.Sprintf("LIMIT $%d OFFSET $%d", index+1, index+2))

	return []interface{}{
		pageSize,
		(page - 1) * pageSize,
	}
}
