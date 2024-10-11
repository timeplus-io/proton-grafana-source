package timeplus

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"

	"github.com/reactivex/rxgo/v2"
	protonDriver "github.com/timeplus-io/proton-go-driver/v2"
)

type Column struct {
	Name string
	Type string
}

type TimeplusEngine struct {
	connection *sql.DB
	logger     log.Logger
}

type TimeplusQueryState struct {
	Query       string
	AddNow      bool
	Stream      chan rxgo.Item
	ColumnArray []Column
	Cancel      context.CancelFunc
}

func NewEngine(logger log.Logger, host string, port int, username, password string) *TimeplusEngine {
	connection := protonDriver.OpenDB(&protonDriver.Options{
		Addr: []string{fmt.Sprintf("%s:%d", host, port)},
		Auth: protonDriver.Auth{
			Username: username,
			Password: password,
		},
		DialTimeout: 10 * time.Second,
		Debug:       false,
	})

	return &TimeplusEngine{
		connection: connection,
		logger:     logger,
	}
}

func (e *TimeplusEngine) Ping() error {
	return e.connection.Ping()
}

func (e *TimeplusEngine) RunQuery(ctx context.Context, sql string) ([]*sql.ColumnType, chan []any, error) {
	ckCtx := protonDriver.Context(ctx)

	rows, err := e.connection.QueryContext(ckCtx, sql)
	if err != nil {
		return nil, nil, err
	}

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, nil, err
	}

	ch := make(chan []any, 1000)

	go func() {
		defer func() {
			close(ch)
		}()

		count := len(columnTypes)
		for rows.Next() {
			values := make([]any, count) // values is raw data
			valuePtrs := make([]any, count)

			row := make([]any, count) // row is string data
			for i := range columnTypes {
				valuePtrs[i] = &values[i]
			}
			if err = rows.Scan(valuePtrs...); err != nil {
				return
			}
			for i := range columnTypes {
				rawValue := values[i]
				row[i] = rawValue
			}

			ch <- row
		}

		rows.Close()
	}()

	return columnTypes, ch, nil
}

func (e *TimeplusEngine) Dispose() error {
	e.logger.Info("Dispose!!!!!")
	return e.connection.Close()
}
