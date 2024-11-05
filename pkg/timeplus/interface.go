package timeplus

import (
	"context"
	"database/sql"
)

type Engine interface {
	RunQuery(ctx context.Context, query string) ([]*sql.ColumnType, chan []any, error)
	IsStreamingQuery(ctx context.Context, query string) (bool, error)

	Ping(ctx context.Context) error
	Dispose() error
}
