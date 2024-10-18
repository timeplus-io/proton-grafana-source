package timeplus

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"

	protonDriver "github.com/timeplus-io/proton-go-driver/v2"
)

type Column struct {
	Name string
	Type string
}

type TimeplusEngine struct {
	connection *sql.DB
	logger     log.Logger
	analyzeURL string
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
		analyzeURL: fmt.Sprintf("http://%s:%d/proton/v1/sqlanalyzer", host, 3218),
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
	return e.connection.Close()
}

func (e *TimeplusEngine) IsStreamingQuery(ctx context.Context, query string) (bool, error) {
	queryMap := map[string]string{"query": query}
	jsonData, err := json.Marshal(queryMap)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.analyzeURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	var response map[string]interface{}
	if err = json.Unmarshal(body, &response); err != nil {
		return false, err
	}

	isStreaming, ok := response["is_streaming"].(bool)
	if !ok {
		return false, fmt.Errorf("invalid response %s", response)
	}

	return isStreaming, nil
}
