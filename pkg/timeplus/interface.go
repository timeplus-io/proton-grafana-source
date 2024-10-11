package timeplus

import (
	"context"
	"database/sql"
)

type Engine interface {
	RunQuery(ctx context.Context, query string) ([]*sql.ColumnType, chan []any, error)

	Ping() error
	Dispose() error
}
